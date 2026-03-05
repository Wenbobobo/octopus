# Octopus
A Telegram bot bridge other IM (qq, wechat, etc.) conversations together.

## Dependencies
* go
* ffmpeg (optional for qq/wechat audio)

# Docker
* [octopus](https://hub.docker.com/r/lxduo/octopus)
```shell
docker run -d -p 11111:11111 --name=octopus --restart=always -v octopus:/data lxduo/octopus:latest
```

# Limbs
* [octopus-qq](https://github.com/duo/octopus-qq)
* [octopus-wechat](https://github.com/duo/octopus-wechat)
* [octopus-wechat-web](https://github.com/duo/octopus-wechat-web)

# Documentation

## Bot
Create a bot with [@BotFather](https://t.me/botfather), get a token.
Set /setjoingroups Enable and /setprivacy Disable

## Configuration
* configure.yaml
```yaml
master:
  api_url: http://10.0.0.10:8081 # Optional, Telegram local bot api server
  local_mode: true # Optional, local server mode
  admin_id: # Required, Telegram user id (administrator)
  token:  1234567:xxxxxxxx # Required, Telegram bot token
  proxy: http://1.1.1.1:7890 # Optional, proxy for Telegram
  page_size: 10 # Optional, command list result pagination size
  archive: # Optional, archive client chat by topic
    - vendor: wechat # qq, wechat, etc
      uid: wxid_xxxxxxx # client id
      chat_id: 123456789 # topic enabled group id (grant related permissions to bot)
  telegraph: # Optional
    enable: true # Convert some message to telegra.ph article (e.g. QQ forward message)
  	proxy: http://1.1.1.1:7890 # Optional, proxy for telegra.ph
    tokens:
      - abcdefg # telegra.ph tokens

service:
  addr: 0.0.0.0:11111 # Required, listen address
  secret: hello # Required, user defined secret
  send_timeout: 3m # Optional
  queue:
    max_events: 4096 # Optional, bridge queue capacity
    overflow_policy: block # Optional, block | drop_oldest
  worker:
    max_concurrency: 8 # Optional
    queue_size: 4096 # Optional
  media:
    max_bytes: 0 # Optional, max download bytes, 0 means unlimited
    download_timeout: 90s # Optional
  storage:
    data_dir: . # Optional, total size monitor root
    max_total_bytes: 0 # Optional, 0 means disabled
    target_total_bytes: 0 # Optional, default 70% of max_total_bytes
    cleanup_interval: 1h # Optional
    message_ttl_days: 0 # Optional, remove expired message mappings
    batch_delete: 500 # Optional

log:
  level: info

wechat_login:
  enable: false
  trigger: startup_check
  timezone: Asia/Shanghai
  relogin_at: "03:00"
  hooks:
    check_logged_in: "" # exit code 0 means logged-in
    resume_login: "" # try click-login flow
    require_scan: "" # fallback flow when scan is required
    timeout: 30s
    retry: 2
    retry_delay: 10s
  qrcode:
    forward_to_tg: false # reserved for phase-3
    capture_cmd: "" # output QR image path
```

## Command
All messages will be sent to the admin directly by default, you can archive chat by topic or /link specific remote chat to a Telegram group.
```
/help Show command list.
/link Manage remote chat link.
/chat Generate a remote chat head.
```

## Stability Notes
- Event bridge queue is now bounded and supports overflow policy.
- Event handling uses worker pools to avoid goroutine spikes under burst traffic.
- Optional media download limits protect memory and bandwidth usage.
- Optional storage cleaner can keep DB/container footprint bounded.
- WeChat auto-login orchestrator supports startup check and daily relogin scheduling via shell hooks.

## Extra Docs
- `docs/upstream-docker-wechat-notes.md`: upstream Docker/GitHub implementation notes (octopus + octopus-wechat).
- `docs/docker-wechat-stack.md`: ready-to-run Docker compose deployment and operations guide.
- `docs/wechat-autologin.md`: auto-login workflow, hook contract, and config examples.
- `docs/wechat-optimization-plan.md`: cost/benefit analysis and staged optimization roadmap.
- `scripts/wechat_login/*.sh`: helper scripts for hook-based WeChat login integration.
