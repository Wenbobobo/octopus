# Upstream Docker + WeChat Notes

Last verified: **2026-02-26**

## Sources
- https://hub.docker.com/r/lxduo/octopus
- https://github.com/duo/octopus
- https://github.com/duo/octopus-wechat
- https://github.com/tom-snow/docker-ComWechat
- https://hub.docker.com/r/tomsnow1999/docker-com_wechat_robot
- https://github.com/ljc545w/ComWeChatRobot

## `lxduo/octopus` image facts
- Docker Hub `latest` currently points to version `0.0.17` lineage.
- Last pushed on Docker Hub timeline: **2024-06-12**.
- Upstream Dockerfile builds Go binary, then runs on Alpine with `ffmpeg` + `tzdata`.
- Container working dir is `/data`, binary entrypoint is `/usr/bin/octopus`.
- Practical implication: persistent files (`master.db`, `configure.yaml`) should be mounted from volume/bind to avoid data loss and uncontrolled layer growth.

## `octopus-wechat` upstream container facts
- Dockerfile base image: `zixia/wechat:3.3.0.115`.
- Starts a VNC server on display `:5`, exposes VNC port `5905`.
- Runtime env includes `VNCPASS` for VNC password.
- Runtime manager (`scripts/run.py`) does:
  1. init agent files (`/usr/bin/init-agent.sh`)
  2. start VNC
  3. run `wine cmd /k /home/user/octopus-wechat/octopus-wechat-x86.exe`
- `init-agent.sh` downloads missing runtime assets at startup:
  - latest `octopus-wechat-x86.exe`
  - DLL bundle from `ComWeChatRobot` release
- Upstream bot login path is API-driven:
  - switch to QR login (`type=41`)
  - poll login status (`type=0`) until success

## WeChat login API signals (from upstream `octopus-wechat` code)
- Hook API endpoint pattern: `http://127.0.0.1:<hook_port>/api/?type=<N>`
- Known types used by login flow:
  - `type=0`: login status (`is_login`)
  - `type=41`: QR code image fetch
- Upstream bot login sequence includes:
  - switch to QR login (`LoginWtihQRCode`)
  - poll login state until success (`ensureLogin`)

## `docker-ComWechat` and `ComWeChatRobot` facts
- `docker-ComWechat` image runs Wine + WeChat + VNC + hook injector in one container.
- Typical run mode uses `--privileged`, `--network host`, and persistent mounts for WeChat data.
- Docker Hub latest size is large (about **2.37GB**), much heavier than `lxduo/octopus`.
- In `ComWeChatRobot` source (`DWeChatRobot/wxsocket.cpp`), `type=41` returns `Content-Type: image/png` chunked image bytes directly when available.
- `type=41` fallback returns JSON message when QR is unavailable/already logged in.
- API enum confirms login and QR endpoints are stable IDs (`0`, `41`) in current branch.

## Integration implications for current repo
- Current auto-login manager can already call shell hooks.
- Recommended first-stage hooks:
  - `check_logged_in`: query `type=0`
  - `resume_login`: no-scan click automation command (placeholder until you provide flow screenshots)
  - `require_scan`: force scan-ready UI command
  - `qrcode.capture_cmd`: query `type=41` and save PNG (used in later TG forwarding phase)
- Cost/performance implication:
  - Prefer API-driven checks (`type=0/41`) before VNC image matching, because they are lower complexity and more stable.
  - Use VNC image matching only for UI branching that API cannot expose (button state, popups).
- Reference scripts added under:
  - `scripts/wechat_login/check_logged_in.sh`
  - `scripts/wechat_login/resume_login.sh`
  - `scripts/wechat_login/require_scan.sh`
  - `scripts/wechat_login/capture_qrcode.sh`
