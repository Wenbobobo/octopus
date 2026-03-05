# Architecture

## Runtime Flow
1. `main.go` loads config and initializes two bridge queues:
   - `master -> slave`
   - `slave -> master`
2. `MasterService` handles Telegram input/output.
3. `LimbService` handles limb/onebot websocket input/output.
4. Both directions are serialized by chat key and processed by bounded worker pools.

## Key Components
- `internal/master`: Telegram adapter and message routing logic.
- `internal/slave`: limb/onebot websocket adapters.
- `internal/filter`: media/sticker/voice transformations.
- `internal/manager`: SQLite-backed mapping tables (`chat`, `link`, `topic`, `message`).
- `internal/storage`: periodic cleanup and compaction.
- `internal/wechatlogin`: startup and scheduled login orchestration hooks.

## Reliability Controls
- Bounded message queue with overflow policy (`block` / `drop_oldest`).
- Bounded worker pools to prevent goroutine storms.
- Media download size and timeout limits.
- DB prune + vacuum for footprint control.
