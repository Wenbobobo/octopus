#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH='' cd -- "$(dirname -- "$0")/../.." && pwd)"
RUNTIME_DIR="${ROOT_DIR}/deploy/wechat/runtime"
TMP_DIR="$(mktemp -d)"
OCTOPUS_RELEASE_TAG="${OCTOPUS_WECHAT_RELEASE_TAG:-}"
COMWECHAT_RELEASE_TAG="${COMWECHAT_RELEASE_TAG:-}"

cleanup() {
  rm -rf "${TMP_DIR}"
}
trap cleanup EXIT INT TERM

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing command: $1" >&2
    exit 1
  fi
}

need_cmd curl

extract_zip() {
  zip_file="$1"
  out_dir="$2"
  if command -v unzip >/dev/null 2>&1; then
    unzip -qq "${zip_file}" -d "${out_dir}"
    return 0
  fi
  need_cmd python3
  python3 - "$zip_file" "$out_dir" <<'PY'
import os
import sys
import zipfile

zip_path = sys.argv[1]
out_path = sys.argv[2]
os.makedirs(out_path, exist_ok=True)
with zipfile.ZipFile(zip_path, "r") as zf:
    zf.extractall(out_path)
PY
}

mkdir -p "${RUNTIME_DIR}"

if [ ! -f "${RUNTIME_DIR}/octopus-wechat-x86.exe" ]; then
  octopus_release_api="https://api.github.com/repos/duo/octopus-wechat/releases/latest"
  if [ -n "${OCTOPUS_RELEASE_TAG}" ]; then
    octopus_release_api="https://api.github.com/repos/duo/octopus-wechat/releases/tags/${OCTOPUS_RELEASE_TAG}"
  fi
  exe_url="$(curl -fsSL "${octopus_release_api}" \
    | sed -n 's/.*"browser_download_url":[[:space:]]*"\(.*x86\.exe\)".*/\1/p' \
    | head -n1)"
  if [ -z "${exe_url}" ]; then
    echo "failed to resolve octopus-wechat exe url" >&2
    exit 1
  fi
  echo "downloading octopus-wechat-x86.exe"
  curl -fL "${exe_url}" -o "${RUNTIME_DIR}/octopus-wechat-x86.exe"
fi

if [ ! -f "${RUNTIME_DIR}/wxDriver.dll" ] || [ ! -f "${RUNTIME_DIR}/SWeChatRobot.dll" ]; then
  comwechat_release_api="https://api.github.com/repos/ljc545w/ComWeChatRobot/releases/latest"
  if [ -n "${COMWECHAT_RELEASE_TAG}" ]; then
    comwechat_release_api="https://api.github.com/repos/ljc545w/ComWeChatRobot/releases/tags/${COMWECHAT_RELEASE_TAG}"
  fi
  zip_url="$(curl -fsSL "${comwechat_release_api}" \
    | sed -n 's/.*"browser_download_url":[[:space:]]*"\(.*\.zip\)".*/\1/p' \
    | head -n1)"
  if [ -z "${zip_url}" ]; then
    echo "failed to resolve ComWeChatRobot zip url" >&2
    exit 1
  fi
  echo "downloading ComWeChatRobot runtime zip"
  curl -fL "${zip_url}" -o "${TMP_DIR}/comwechat.zip"
  extract_zip "${TMP_DIR}/comwechat.zip" "${TMP_DIR}/extract"
  for dll in wxDriver.dll wxDriver64.dll SWeChatRobot.dll; do
    dll_path="$(find "${TMP_DIR}/extract" -type f -name "${dll}" | head -n1 || true)"
    if [ -n "${dll_path}" ]; then
      cp -f "${dll_path}" "${RUNTIME_DIR}/${dll}"
    fi
  done
fi

ls -lh "${RUNTIME_DIR}"/octopus-wechat-x86.exe "${RUNTIME_DIR}"/*.dll
