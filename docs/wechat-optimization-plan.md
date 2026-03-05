# WeChat Hook Optimization (Cost vs Benefit)

Last updated: **2026-02-26**

## Goal
Optimize reliability and operating cost for WeChat auto-login and hook integration, while keeping risk low and preserving existing anti-ban behavior.

## Constraints
- Daily relogin behavior around 03:00 (Asia/Shanghai).
- Two real-world login branches:
  1. no-scan click resume
  2. scan-required fallback
- Visual templates/screenshots will be provided later.

## Option Matrix

### Option A: API-first (recommended baseline)
- Use hook API only (`type=0`, `type=41`) for state and QR acquisition.
- Keep UI operations manual or external script placeholders.
- Dev cost: Low
- Runtime risk: Low
- Benefit: High initial stability, fast delivery

### Option B: VNC click automation only
- Use fixed-coordinate clicks/keystrokes, minimal state detection.
- Dev cost: Medium
- Runtime risk: Medium-High (resolution/theme/window drift)
- Benefit: Medium

### Option C: Full VNC image matching pipeline
- Add template matching + region detection + action policy.
- Dev cost: High
- Runtime risk: Medium (better than fixed coordinates)
- Benefit: High automation coverage

### Option D: Hybrid (API-first + selective visual fallback)
- Use API for hard signals, visual only for ambiguous UI branches.
- Dev cost: Medium
- Runtime risk: Low-Medium
- Benefit: High

## Recommendation
Adopt **Option D** in stages:

1. Stage 1 (now): API-first
- `check_logged_in`: query `type=0`
- `qrcode.capture_cmd`: fetch QR via `type=41`
- keep `resume_login`/`require_scan` as externally injected commands

2. Stage 2 (after screenshots): selective visual fallback
- add template matching only for:
  - login button detection
  - scan-required state detection
  - common popup dismissal
- avoid full-screen brute-force matching

3. Stage 3: optimize reliability/ops
- add retry jitter and cooldown to avoid frequent repeated UI actions
- add per-step metrics and screenshot audit sampling
- wire QR image forwarding to Telegram

## Practical Cost Control
- Keep `octopus` and `octopus-wechat` state in mounted volumes; avoid image-layer growth.
- Enforce media/storage limits in octopus config.
- Keep automation logic mostly script-driven to reduce Go code churn and speed iteration.

## Immediate Action Items
1. Use helper scripts in `scripts/wechat_login/` for API-first stage.
2. Configure `wechat_login.hooks.*` to call those scripts.
3. Wait for your screenshot pack to implement stage-2 visual branching.
