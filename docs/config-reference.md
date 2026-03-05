# Config Reference (Additions)

## service.queue
- `max_events` (int, default `4096`)
- `overflow_policy` (`block` | `drop_oldest`, default `block`)

## service.worker
- `max_concurrency` (int, default `8`)
- `queue_size` (int, default `max_events`)

## service.media
- `max_bytes` (int64, default `0`, unlimited)
- `download_timeout` (duration, default `90s`)

## service.storage
- `data_dir` (string, default `.`)
- `max_total_bytes` (int64, default `0`, disabled)
- `target_total_bytes` (int64, default `70%` of max)
- `cleanup_interval` (duration, default `1h`)
- `message_ttl_days` (int, default `0`, disabled)
- `batch_delete` (int, default `500`)

## wechat_login
- `enable` (bool, default `false`)
- `trigger` (string, default `startup_check`)
- `timezone` (string, default `Asia/Shanghai`)
- `relogin_at` (string `HH:MM`, default `03:00`)
- `hooks.check_logged_in` (shell command)
- `hooks.resume_login` (shell command)
- `hooks.require_scan` (shell command)
- `hooks.timeout` (duration, default `30s`)
- `hooks.retry` (int, default `2`)
- `hooks.retry_delay` (duration, default `10s`)
- `qrcode.forward_to_tg` (bool, reserved for next phase)
- `qrcode.capture_cmd` (shell command, reserved for next phase)

## Suggested Script Paths
- `scripts/wechat_login/check_logged_in.sh`
- `scripts/wechat_login/resume_login.sh`
- `scripts/wechat_login/require_scan.sh`
- `scripts/wechat_login/capture_qrcode.sh`
