#!/usr/bin/env sh
set -eu

HOOK_HOST="${HOOK_HOST:-127.0.0.1}"
HOOK_PORT="${HOOK_PORT:-22223}"
TIMEOUT="${TIMEOUT:-10}"
OUT="${1:-/tmp/wechat_qrcode.png}"
TMP="${OUT}.tmp"

curl -fsS -m "$TIMEOUT" -X POST "http://${HOOK_HOST}:${HOOK_PORT}/api/?type=41" -d '{}' -o "$TMP"

# If response is JSON, it's usually an error payload rather than image bytes.
if head -c 1 "$TMP" 2>/dev/null | grep -q '{'; then
  cat "$TMP" >&2 || true
  rm -f "$TMP"
  exit 1
fi

mv "$TMP" "$OUT"
printf '%s\n' "$OUT"
