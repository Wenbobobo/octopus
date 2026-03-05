# WeChat Auto Login

## Current Scope
- Startup login check.
- Daily relogin trigger (`Asia/Shanghai` + `relogin_at`).
- Hook-based flow for resume/scan fallback.
- Designed for Linux container + VNC workflows.
- Upstream-aligned with `octopus-wechat` hook API (`type=0` login status, `type=41` QR fetch).
- Recommended strategy is hybrid: API-first + selective visual fallback (see `docs/wechat-optimization-plan.md`).

## Hook Contract
All hooks are shell commands.

1. `check_logged_in`
- Exit code `0`: logged-in.
- Non-zero: not ready.

2. `resume_login`
- Should attempt the no-scan click path when possible.

3. `require_scan`
- Runs after resume retries fail.
- Expected to switch UI to scan-ready state.

## Suggested Hook Scripts
Use the helper scripts added in this repo:
- `scripts/wechat_login/check_logged_in.sh`
- `scripts/wechat_login/set_version.sh`
- `scripts/wechat_login/resume_login.sh`
- `scripts/wechat_login/require_scan.sh`
- `scripts/wechat_login/capture_qrcode.sh`

`set_version.sh` applies ComWeChat `type=35` patch automatically (default target version `3.9.12.16`).

Example config:
```yaml
wechat_login:
  enable: true
  trigger: startup_check
  timezone: Asia/Shanghai
  relogin_at: "03:00"
  hooks:
    check_logged_in: "/data/scripts/wechat_login/check_logged_in.sh"
    resume_login: "/data/scripts/wechat_login/resume_login.sh"
    require_scan: "/data/scripts/wechat_login/require_scan.sh"
    timeout: 30s
    retry: 2
    retry_delay: 10s
  qrcode:
    forward_to_tg: false
    capture_cmd: "/data/scripts/wechat_login/capture_qrcode.sh /tmp/wechat_qrcode.png"
```

## Collaboration Inputs Needed
For better matching/automation scripts, provide samples for:
- logged-in state
- click-login state
- scan-required state
- common error popup state

## After You Provide Screenshots
1. We add template matching for state detection (no-scan vs scan-required).
2. We replace placeholder `resume_login` / `require_scan` commands with stable UI actions.
3. We enable TG QR forwarding flow using `capture_qrcode.sh` output.

## Phase-3 Placeholder
- `qrcode.capture_cmd` + TG forwarding will be wired in a later phase.
