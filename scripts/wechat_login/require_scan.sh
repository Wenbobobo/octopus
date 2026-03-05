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

# Optional scan-required flow.
# You can inject a concrete VNC/UI automation command via REQUIRE_SCAN_CMD.
if [ -n "${REQUIRE_SCAN_CMD:-}" ]; then
  /bin/sh -lc "$REQUIRE_SCAN_CMD"
  exit $?
fi

if [ "$set_version_ok" -eq 0 ]; then
  exit 0
fi

echo "REQUIRE_SCAN_CMD is not configured and version patch failed" >&2
exit 1
