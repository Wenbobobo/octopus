#!/usr/bin/env sh
set -eu

# Placeholder entrypoint for no-scan login click flow.
# You can inject a concrete VNC/UI automation command via RESUME_LOGIN_CMD.
if [ -n "${RESUME_LOGIN_CMD:-}" ]; then
  /bin/sh -lc "$RESUME_LOGIN_CMD"
  exit $?
fi

echo "RESUME_LOGIN_CMD is not configured" >&2
exit 1
