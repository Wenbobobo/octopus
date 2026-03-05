package wechatlogin

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/duo/octopus/internal/common"

	log "github.com/sirupsen/logrus"
)

type Manager struct {
	config *common.Configure

	isConnected func() bool
	notify      func(string)

	stop chan struct{}
	done chan struct{}
	once sync.Once

	running atomic.Bool
}

func NewManager(config *common.Configure, isConnected func() bool, notify func(string)) *Manager {
	if isConnected == nil {
		isConnected = func() bool { return false }
	}
	if notify == nil {
		notify = func(msg string) {
			log.Infof("WeChatLogin: %s", msg)
		}
	}

	return &Manager{
		config:      config,
		isConnected: isConnected,
		notify:      notify,
		stop:        make(chan struct{}),
		done:        make(chan struct{}),
	}
}

func (m *Manager) Start() {
	if !m.config.WechatLogin.Enable {
		return
	}

	go func() {
		defer close(m.done)

		if m.config.WechatLogin.Trigger == "" || m.config.WechatLogin.Trigger == "startup_check" {
			m.runFlow("startup", false)
		}

		m.scheduleDailyRelogin()
	}()
}

func (m *Manager) Stop() {
	if !m.config.WechatLogin.Enable {
		return
	}
	m.once.Do(func() {
		close(m.stop)
		<-m.done
	})
}

func (m *Manager) scheduleDailyRelogin() {
	loc, err := time.LoadLocation(m.config.WechatLogin.Timezone)
	if err != nil {
		m.notify(fmt.Sprintf("invalid timezone %q: %v", m.config.WechatLogin.Timezone, err))
		return
	}

	hour, minute, err := parseClock(m.config.WechatLogin.ReloginAt)
	if err != nil {
		m.notify(fmt.Sprintf("invalid relogin_at %q: %v", m.config.WechatLogin.ReloginAt, err))
		return
	}

	for {
		nextRun := nextClockTime(time.Now().In(loc), hour, minute)
		wait := time.Until(nextRun)
		if wait < 0 {
			wait = 0
		}
		timer := time.NewTimer(wait)

		select {
		case <-m.stop:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		case <-timer.C:
			m.runFlow("daily", true)
		}
	}
}

func (m *Manager) runFlow(trigger string, force bool) {
	if !m.running.CompareAndSwap(false, true) {
		m.notify(fmt.Sprintf("skip %s login flow: another flow is running", trigger))
		return
	}
	defer m.running.Store(false)

	m.notify(fmt.Sprintf("start login flow (%s)", trigger))
	if m.checkLoggedIn() && !force {
		m.notify("wechat client already online")
		return
	}

	retries := m.config.WechatLogin.Hooks.Retry + 1
	for attempt := 1; attempt <= retries; attempt++ {
		if err := m.runHook("resume_login", m.config.WechatLogin.Hooks.ResumeLogin); err != nil {
			m.notify(fmt.Sprintf("resume_login attempt %d/%d failed: %v", attempt, retries, err))
		} else if m.waitForOnline() {
			m.notify("wechat login resumed and online")
			return
		}

		if attempt < retries {
			if !m.sleepWithStop(m.config.WechatLogin.Hooks.RetryDelay) {
				return
			}
		}
	}

	if cmd := m.config.WechatLogin.Hooks.RequireScan; cmd != "" {
		if err := m.runHook("require_scan", cmd); err != nil {
			m.notify(fmt.Sprintf("require_scan hook failed: %v", err))
		} else {
			m.notify("require_scan hook executed")
		}
	}

	if m.config.WechatLogin.QRCode.ForwardToTG && m.config.WechatLogin.QRCode.CaptureCmd != "" {
		output, err := m.runCommand(m.config.WechatLogin.QRCode.CaptureCmd)
		if err != nil {
			m.notify(fmt.Sprintf("capture_qrcode failed: %v", err))
		} else {
			m.notify(fmt.Sprintf("qrcode captured: %s", strings.TrimSpace(output)))
		}
	}

	m.notify("wechat login requires scan or manual intervention")
}

func (m *Manager) waitForOnline() bool {
	if m.checkLoggedIn() {
		return true
	}

	deadline := time.Now().Add(m.config.WechatLogin.Hooks.Timeout)
	for time.Now().Before(deadline) {
		if m.checkLoggedIn() {
			return true
		}
		if !m.sleepWithStop(2 * time.Second) {
			return false
		}
	}

	return m.checkLoggedIn()
}

func (m *Manager) checkLoggedIn() bool {
	if m.isConnected() {
		return true
	}
	cmd := strings.TrimSpace(m.config.WechatLogin.Hooks.CheckLoggedIn)
	if cmd == "" {
		return false
	}
	if _, err := m.runCommand(cmd); err != nil {
		return false
	}
	return true
}

func (m *Manager) runHook(name, cmd string) error {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return fmt.Errorf("%s command is empty", name)
	}
	_, err := m.runCommand(cmd)
	return err
}

func (m *Manager) runCommand(command string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.config.WechatLogin.Hooks.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", "-lc", command)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	outText := strings.TrimSpace(output.String())
	if ctx.Err() == context.DeadlineExceeded {
		return outText, fmt.Errorf("command timeout")
	}
	if err != nil {
		if outText == "" {
			return "", err
		}
		return outText, fmt.Errorf("%w: %s", err, outText)
	}
	return outText, nil
}

func (m *Manager) sleepWithStop(d time.Duration) bool {
	if d <= 0 {
		return true
	}
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-m.stop:
		return false
	case <-timer.C:
		return true
	}
}

func parseClock(clock string) (int, int, error) {
	parts := strings.Split(clock, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expect HH:MM")
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("out of range")
	}
	return hour, minute, nil
}

func nextClockTime(now time.Time, hour int, minute int) time.Time {
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
