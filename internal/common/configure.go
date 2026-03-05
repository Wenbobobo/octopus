package common

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultPageSize             = 10
	defaultSendTimeout          = 3 * time.Minute
	defaultQueueMaxEvents       = 4096
	defaultQueueOverflowPolicy  = "block"
	defaultWorkerMaxConcurrency = 8
	defaultWorkerQueueSize      = 4096
	defaultMediaDownloadTimeout = 90 * time.Second
	defaultStorageCleanup       = 1 * time.Hour
	defaultStorageTargetRatio   = 7 // 70%
	defaultWechatTimezone       = "Asia/Shanghai"
	defaultWechatReloginAt      = "03:00"
)

type ArchiveChat struct {
	Vendor string `yaml:"vendor"`
	UID    string `yaml:"uid"`
	ChatID int64  `yaml:"chat_id"`
}

type Configure struct {
	Master struct {
		APIURL    string        `yaml:"api_url"`
		LocalMode bool          `yaml:"local_mode"`
		AdminID   int64         `yaml:"admin_id"`
		Token     string        `yaml:"token"`
		Proxy     string        `yaml:"proxy"`
		PageSize  int           `yaml:"page_size"`
		Archive   []ArchiveChat `yaml:"archive"`

		Telegraph struct {
			Enable bool     `yaml:"enable"`
			Proxy  string   `yaml:"proxy"`
			Tokens []string `yaml:"tokens"`
		} `yaml:"telegraph"`
	} `yaml:"master"`

	Service struct {
		Addr   string `yaml:"addr"`
		Secret string `yaml:"secret"`

		// SendTiemout keeps backward compatibility with historical typo usage.
		SendTiemout time.Duration `yaml:"send_tiemout"`
		SendTimeout time.Duration `yaml:"send_timeout"`

		Queue struct {
			MaxEvents      int    `yaml:"max_events"`
			OverflowPolicy string `yaml:"overflow_policy"`
		} `yaml:"queue"`

		Worker struct {
			MaxConcurrency int `yaml:"max_concurrency"`
			QueueSize      int `yaml:"queue_size"`
		} `yaml:"worker"`

		Media struct {
			MaxBytes        int64         `yaml:"max_bytes"`
			DownloadTimeout time.Duration `yaml:"download_timeout"`
		} `yaml:"media"`

		Storage struct {
			DataDir          string        `yaml:"data_dir"`
			MaxTotalBytes    int64         `yaml:"max_total_bytes"`
			TargetTotalBytes int64         `yaml:"target_total_bytes"`
			CleanupInterval  time.Duration `yaml:"cleanup_interval"`
			MessageTTLDays   int           `yaml:"message_ttl_days"`
			BatchDelete      int           `yaml:"batch_delete"`
		} `yaml:"storage"`
	} `yaml:"service"`

	WechatLogin struct {
		Enable    bool   `yaml:"enable"`
		Trigger   string `yaml:"trigger"`
		Timezone  string `yaml:"timezone"`
		ReloginAt string `yaml:"relogin_at"`

		Hooks struct {
			CheckLoggedIn string        `yaml:"check_logged_in"`
			ResumeLogin   string        `yaml:"resume_login"`
			RequireScan   string        `yaml:"require_scan"`
			Timeout       time.Duration `yaml:"timeout"`
			Retry         int           `yaml:"retry"`
			RetryDelay    time.Duration `yaml:"retry_delay"`
		} `yaml:"hooks"`

		QRCode struct {
			ForwardToTG bool   `yaml:"forward_to_tg"`
			CaptureCmd  string `yaml:"capture_cmd"`
		} `yaml:"qrcode"`
	} `yaml:"wechat_login"`

	Log struct {
		Level string `yaml:"level"`
	} `yaml:"log"`
}

func LoadConfig(path string) (*Configure, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Configure{}
	config.Master.APIURL = "https://api.telegram.org"
	config.Master.PageSize = defaultPageSize
	config.Service.SendTimeout = defaultSendTimeout
	config.Service.SendTiemout = defaultSendTimeout
	config.Service.Queue.MaxEvents = defaultQueueMaxEvents
	config.Service.Queue.OverflowPolicy = defaultQueueOverflowPolicy
	config.Service.Worker.MaxConcurrency = defaultWorkerMaxConcurrency
	config.Service.Worker.QueueSize = defaultWorkerQueueSize
	config.Service.Media.DownloadTimeout = defaultMediaDownloadTimeout
	config.Service.Storage.DataDir = "."
	config.Service.Storage.CleanupInterval = defaultStorageCleanup
	config.WechatLogin.Timezone = defaultWechatTimezone
	config.WechatLogin.ReloginAt = defaultWechatReloginAt
	config.WechatLogin.Hooks.Timeout = 30 * time.Second
	config.WechatLogin.Hooks.Retry = 2
	config.WechatLogin.Hooks.RetryDelay = 10 * time.Second
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	if config.Service.SendTimeout <= 0 && config.Service.SendTiemout > 0 {
		config.Service.SendTimeout = config.Service.SendTiemout
	}
	if config.Service.SendTimeout <= 0 {
		config.Service.SendTimeout = defaultSendTimeout
	}
	config.Service.SendTiemout = config.Service.SendTimeout

	if config.Service.Queue.MaxEvents <= 0 {
		config.Service.Queue.MaxEvents = defaultQueueMaxEvents
	}
	switch config.Service.Queue.OverflowPolicy {
	case "", "block", "drop_oldest":
		if config.Service.Queue.OverflowPolicy == "" {
			config.Service.Queue.OverflowPolicy = defaultQueueOverflowPolicy
		}
	default:
		config.Service.Queue.OverflowPolicy = defaultQueueOverflowPolicy
	}

	if config.Service.Worker.MaxConcurrency <= 0 {
		config.Service.Worker.MaxConcurrency = defaultWorkerMaxConcurrency
	}
	if config.Service.Worker.QueueSize <= 0 {
		config.Service.Worker.QueueSize = config.Service.Queue.MaxEvents
	}

	if config.Service.Media.DownloadTimeout <= 0 {
		config.Service.Media.DownloadTimeout = defaultMediaDownloadTimeout
	}
	if config.Service.Storage.CleanupInterval <= 0 {
		config.Service.Storage.CleanupInterval = defaultStorageCleanup
	}
	if config.Service.Storage.TargetTotalBytes <= 0 && config.Service.Storage.MaxTotalBytes > 0 {
		config.Service.Storage.TargetTotalBytes = config.Service.Storage.MaxTotalBytes * defaultStorageTargetRatio / 10
	}
	if config.Service.Storage.BatchDelete <= 0 {
		config.Service.Storage.BatchDelete = 500
	}
	if config.WechatLogin.Timezone == "" {
		config.WechatLogin.Timezone = defaultWechatTimezone
	}
	if config.WechatLogin.ReloginAt == "" {
		config.WechatLogin.ReloginAt = defaultWechatReloginAt
	}
	if config.WechatLogin.Hooks.Timeout <= 0 {
		config.WechatLogin.Hooks.Timeout = 30 * time.Second
	}
	if config.WechatLogin.Hooks.Retry < 0 {
		config.WechatLogin.Hooks.Retry = 0
	}
	if config.WechatLogin.Hooks.RetryDelay <= 0 {
		config.WechatLogin.Hooks.RetryDelay = 10 * time.Second
	}

	return config, nil
}
