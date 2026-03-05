# Agent Runbook

## First 10 Minutes
1. Read `README.md` and `docs/config-reference.md`.
2. Read deployment guide: `docs/docker-wechat-stack.md`.
3. Read upstream notes: `docs/upstream-docker-wechat-notes.md`.
4. Read strategy notes: `docs/wechat-optimization-plan.md`.
5. Inspect runtime logs for queue overflow, cleaner, and login flow messages.
6. Confirm current config values for queue/worker/media/storage.

## Common Tasks
- Performance tuning: adjust `service.queue` + `service.worker`.
- Storage tuning: adjust `service.storage.max_total_bytes` and interval.
- WeChat login tuning: refine hook scripts and timeout/retry values.
- WeChat optimization decisions: follow staged plan in `docs/wechat-optimization-plan.md`.

## Safety Rules
- Preserve anti-ban behavior equivalence when refactoring.
- Avoid changing message routing semantics without regression tests.
- Keep new behavior behind config toggles where possible.
