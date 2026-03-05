#!/usr/bin/env sh
set -eu

HOOK_HOST="${HOOK_HOST:-127.0.0.1}"
HOOK_PORT="${HOOK_PORT:-22223}"
TIMEOUT="${TIMEOUT:-5}"

resp="$(curl -fsS -m "$TIMEOUT" -X POST "http://${HOOK_HOST}:${HOOK_PORT}/api/?type=0" -d '{}')"

# Return 0 only when both result and login flag are positive.
printf '%s' "$resp" | grep -q '"result":"OK"'
printf '%s' "$resp" | grep -Eq '"is_login"[[:space:]]*:[[:space:]]*1'
