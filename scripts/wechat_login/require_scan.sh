#!/usr/bin/env sh
set -eu

# Placeholder entrypoint for scan-required flow.
# You can inject a concrete VNC/UI automation command via REQUIRE_SCAN_CMD.
if [ -n "${REQUIRE_SCAN_CMD:-}" ]; then
  /bin/sh -lc "$REQUIRE_SCAN_CMD"
  exit $?
fi

echo "REQUIRE_SCAN_CMD is not configured" >&2
exit 1
