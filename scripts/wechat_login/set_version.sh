#!/usr/bin/env sh
set -eu

HOOK_HOST="${HOOK_HOST:-127.0.0.1}"
TIMEOUT="${TIMEOUT:-5}"
WECHAT_VERSION="${WECHAT_VERSION:-3.9.12.16}"
WECHAT_API_PORTS_CSV="${WECHAT_API_PORTS_CSV:-${HOOK_PORT:-22223},18888}"

set_ok=1

OLD_IFS="$IFS"
IFS=','
for port in $WECHAT_API_PORTS_CSV; do
  IFS="$OLD_IFS"
  port="$(printf '%s' "$port" | tr -d '[:space:]')"
  if [ -z "$port" ]; then
    IFS=','
    continue
  fi
  resp="$(curl -fsS -m "$TIMEOUT" -X POST "http://${HOOK_HOST}:${port}/api/?type=35" \
    -d "{\"version\":\"${WECHAT_VERSION}\"}" 2>/dev/null || true)"
  if printf '%s' "$resp" | grep -q '"result":"OK"'; then
    printf 'version_set_ok port=%s version=%s\n' "$port" "$WECHAT_VERSION"
    set_ok=0
    break
  fi
  IFS=','
done
IFS="$OLD_IFS"

if [ "$set_ok" -ne 0 ]; then
  echo "failed to set wechat version via type=35 on ports: ${WECHAT_API_PORTS_CSV}" >&2
  exit 1
fi
