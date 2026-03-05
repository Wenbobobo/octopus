# Docker WeChat Stack

## Purpose
Provide a reproducible local deployment for:
- `octopus` bridge service
- `octopus-wechat` limb service (VNC + hook API)

This setup keeps data persistent and bounded, and keeps WeChat login automation hook-based.

## Files
- `docker-compose.wechat.yml`
- `deploy/octopus/configure.yaml`
- `deploy/wechat/runtime/configure.yaml`
- `deploy/octopus/data/` (runtime DB and data)
- `deploy/wechat/runtime/` (wechat limb runtime files)
- `deploy/wechat/wechat_files/` and `deploy/wechat/application_data/` (WeChat persistence)

## One-Time Edit
1. Edit `deploy/octopus/configure.yaml`:
- `master.admin_id`
- `master.token`
- `service.secret`
2. Edit `deploy/wechat/runtime/configure.yaml`:
- keep `service.secret` equal to octopus `service.secret`
3. Copy env template and edit optional runtime commands:
```bash
cp .env.wechat.example .env.wechat
```
4. Recommended defaults in `.env.wechat`:
- `WECHAT_VERSION=3.9.12.16`
- `WECHAT_API_PORTS_CSV=22223,18888`

## Pull and Build
```bash
docker pull lxduo/octopus-wechat:latest
docker compose -f docker-compose.wechat.yml build octopus
./scripts/deploy/prepare_wechat_runtime.sh
docker compose -f docker-compose.wechat.yml config >/tmp/octopus-wechat-compose.rendered.yaml
```

Optional pin download tags for runtime assets:
```bash
OCTOPUS_WECHAT_RELEASE_TAG=master_202305181416 \
COMWECHAT_RELEASE_TAG=3.7.0.30-0.0.9 \
./scripts/deploy/prepare_wechat_runtime.sh
```

## Start
```bash
docker compose --env-file .env.wechat -f docker-compose.wechat.yml up -d
```

## Useful Commands
```bash
docker compose -f docker-compose.wechat.yml ps
docker compose -f docker-compose.wechat.yml logs -f octopus
docker compose -f docker-compose.wechat.yml logs -f octopus-wechat
docker compose -f docker-compose.wechat.yml restart octopus
docker compose -f docker-compose.wechat.yml down
```

## VNC and Login Debug
- VNC address: `<host-ip>:5905`
- VNC password: `VNCPASS` environment value
- Hook login status check (inside octopus):
```bash
docker compose -f docker-compose.wechat.yml exec octopus \
  /data/scripts/wechat_login/check_logged_in.sh
```
- Capture QR image (inside octopus):
```bash
docker compose -f docker-compose.wechat.yml exec octopus \
  /data/scripts/wechat_login/capture_qrcode.sh /tmp/wechat_qrcode.png
```

## Notes
- Compose uses `network_mode: host`, aligned with common ComWechat lineage runtime behavior.
- `octopus-wechat` now runs through `scripts/deploy/wechat_supervisor.sh`:
  - cleans stale X locks before each agent cycle
  - restarts VNC and agent in-process (container stays alive)
  - auto-calls `type=35` version patch with `WECHAT_VERSION`
- `scripts/deploy/prepare_wechat_runtime.sh` preloads `octopus-wechat-x86.exe` and required DLL files into `deploy/wechat/runtime/`.
- `deploy/octopus/configure.yaml` already enables:
  - bounded queue/worker
  - media download cap
  - periodic storage cleanup
  - WeChat login manager hooks
- `resume_login` / `require_scan` remain command placeholders via:
  - `RESUME_LOGIN_CMD`
  - `REQUIRE_SCAN_CMD`
  These will be finalized after your login screenshots are provided.
- If `octopus-wechat` logs `Failed to switch to QR login` during first-time bring-up, collect VNC screenshots and log tail for state-specific tuning.
- If WeChat UI shows network anomaly but logs show repeated agent exits, first check whether `type=35` patch is actually applied in logs (`version set via port ...` from supervisor).
