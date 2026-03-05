#!/usr/bin/env sh
set -eu

DISPLAY_NUM="${DISPLAY_NUM:-5}"
DISPLAY=":${DISPLAY_NUM}"
AGENT_BIN="${AGENT_BIN:-/home/user/octopus-wechat/octopus-wechat-x86.exe}"
RESTART_DELAY="${WECHAT_AGENT_RESTART_DELAY:-3}"
HOOK_HOST="${HOOK_HOST:-127.0.0.1}"
WECHAT_VERSION="${WECHAT_VERSION:-3.9.12.16}"
WECHAT_API_PORTS_CSV="${WECHAT_API_PORTS_CSV:-22223,18888}"

log() {
  printf '[wechat-supervisor] %s\n' "$*"
}

cleanup_display() {
  /usr/bin/vncserver -kill "$DISPLAY" >/dev/null 2>&1 || true
  pkill -f "Xtigervnc.*${DISPLAY}" >/dev/null 2>&1 || true
  sudo rm -f "/tmp/.X${DISPLAY_NUM}-lock" "/tmp/.X11-unix/X${DISPLAY_NUM}" >/dev/null 2>&1 || true
}

start_vnc() {
  mkdir -p /home/user/.vnc
  /usr/bin/vncpasswd -f <<EOF >/home/user/.vnc/passwd
${VNCPASS:-YourSafeVNCPassword}
EOF
  chmod 700 /home/user/.vnc/passwd
  /usr/bin/vncserver -localhost no -xstartup /usr/bin/openbox "$DISPLAY"
}

set_version_once() {
  old_ifs="$IFS"
  IFS=','
  for port in $WECHAT_API_PORTS_CSV; do
    IFS="$old_ifs"
    port="$(printf '%s' "$port" | tr -d '[:space:]')"
    if [ -z "$port" ]; then
      IFS=','
      continue
    fi
    resp="$(curl -fsS -m 3 -X POST "http://${HOOK_HOST}:${port}/api/?type=35" \
      -d "{\"version\":\"${WECHAT_VERSION}\"}" 2>/dev/null || true)"
    if printf '%s' "$resp" | grep -q '"result":"OK"'; then
      log "version set via port ${port}: ${WECHAT_VERSION}"
      IFS="$old_ifs"
      return 0
    fi
    IFS=','
  done
  IFS="$old_ifs"
  return 1
}

version_guard() {
  agent_pid="$1"
  retries=0
  while kill -0 "$agent_pid" >/dev/null 2>&1; do
    if set_version_once; then
      return 0
    fi
    retries=$((retries + 1))
    if [ "$retries" -ge 30 ]; then
      log "version patch not ready yet after ${retries} retries"
      retries=0
    fi
    sleep 2
  done
  return 1
}

if [ -x /usr/bin/init-agent.sh ]; then
  /usr/bin/init-agent.sh || log "init-agent failed, continue with mounted runtime files"
fi

while true; do
  if [ ! -f "$AGENT_BIN" ]; then
    log "agent binary missing: $AGENT_BIN"
    sleep "$RESTART_DELAY"
    continue
  fi

  cleanup_display
  log "starting vnc on ${DISPLAY}"
  start_vnc

  log "starting agent: $AGENT_BIN"
  wine cmd /k "$AGENT_BIN" &
  agent_pid="$!"

  version_guard "$agent_pid" >/dev/null 2>&1 || true

  wait "$agent_pid" || true
  log "agent exited, restart in ${RESTART_DELAY}s"
  sleep "$RESTART_DELAY"
done
