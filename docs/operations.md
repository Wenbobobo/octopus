# Operations

## Startup Checks
1. Verify Telegram bot token/admin id.
2. Verify websocket secret and limb connections.
3. If enabled, WeChat login manager runs startup check.
4. For upstream runtime details, read `docs/upstream-docker-wechat-notes.md`.

## Capacity Controls
- `service.queue.max_events`: queue capacity per bridge direction.
- `service.worker.max_concurrency`: max parallel event workers.
- `service.media.max_bytes`: max download size per remote file.

## Storage Cleanup
- Enable with `service.storage.max_total_bytes > 0`.
- Cleaner runs every `service.storage.cleanup_interval`.
- Cleanup order:
  1. TTL cleanup (`message_ttl_days`).
  2. If still over limit, prune oldest message mappings in batches.
  3. `VACUUM` to reclaim sqlite file space.
- Mount `/data` as persistent volume so DB growth is visible and controllable from host.
- Avoid writing bulky transient files into image layers; place runtime temp files on mounted paths.

## Docker Footprint Tips
- `octopus` image itself is small (around tens of MB), but runtime volume can grow from DB/media temp files.
- Keep `service.media.max_bytes` enabled to prevent oversized payload persistence paths.
- Use `service.storage.max_total_bytes` + `target_total_bytes` to enforce hard ceiling.
- If experimenting frequently, periodically clean stopped containers/images on host side (`docker system df` first, then prune with caution).
- If you use `docker-com_wechat_robot` lineage images, note they are much heavier (GB-level) and often require `--privileged` + `network_mode: host`; isolate them from the bridge service container when possible.

## Troubleshooting
- Repeated `LimbClient not found`: check vendor id and limb websocket status.
- High memory under burst: lower `max_concurrency` and increase backpressure (`overflow_policy=block`).
- Oversized storage: decrease `target_total_bytes` and/or increase cleanup frequency.
