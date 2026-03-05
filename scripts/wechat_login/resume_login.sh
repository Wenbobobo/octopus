#!/usr/bin/env sh
set -eu

# Apply version compatibility patch first.
set_version_cmd="$(dirname "$0")/set_version.sh"
set_version_ok=1
if [ -x "$set_version_cmd" ]; then
  if "$set_version_cmd"; then
    set_version_ok=0
  fi
fi

# Optional no-scan login click flow.
# You can inject a concrete VNC/UI automation command via RESUME_LOGIN_CMD.
if [ -n "${RESUME_LOGIN_CMD:-}" ]; then
  /bin/sh -lc "$RESUME_LOGIN_CMD"
  exit $?
fi

# If no custom command, return success when version patch was applied.
if [ "$set_version_ok" -eq 0 ]; then
  exit 0
fi

echo "RESUME_LOGIN_CMD is not configured and version patch failed" >&2
exit 1
