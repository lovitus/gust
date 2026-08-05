#!/usr/bin/env bash
set -euo pipefail

asset_root="${1:-}"
if [[ -z "${asset_root}" || ! -d "${asset_root}" ]]; then
  echo "usage: $0 ASSET_ROOT" >&2
  exit 2
fi
asset_root="$(cd "${asset_root}" && pwd)"

binary="${asset_root}/gust-with-singbox"
if [[ "$(go env GOOS)" == windows ]]; then
  binary="${binary}.exe"
fi
if [[ ! -f "${binary}" ]]; then
  echo "missing asset binary: ${binary}" >&2
  exit 1
fi

"${binary}" -V | tee "${asset_root}/smoke-version.txt"
grep -F 'flavor=singbox' "${asset_root}/smoke-version.txt"

echo_log="${RUNNER_TEMP:-/tmp}/singbox-echo.log"
gust_log="${RUNNER_TEMP:-/tmp}/singbox-direct.log"
python3 tests/e2e/scripts/tcp_echo.py >"${echo_log}" 2>&1 &
echo_pid=$!
(
  cd "${asset_root}"
  "${binary}" -L 'http://127.0.0.1:18888' -F 'direct+singbox://' >"${gust_log}" 2>&1
) &
gust_pid=$!
cleanup() {
  kill "${gust_pid}" "${echo_pid}" 2>/dev/null || true
  wait "${gust_pid}" "${echo_pid}" 2>/dev/null || true
}
trap cleanup EXIT

passed=false
for _ in {1..40}; do
  if curl --fail --silent --show-error --max-time 2 \
    --proxy http://127.0.0.1:18888 http://127.0.0.1:5678 \
    | grep -Fx 'hello-gost'; then
    passed=true
    break
  fi
  sleep 0.25
done
if [[ "${passed}" != true ]]; then
  cat "${gust_log}" >&2
  exit 1
fi

if python3 -c 'import json,sys; sys.exit("naive" in json.load(open(sys.argv[1], encoding="utf-8"))["unavailableFeatures"])' \
  "${asset_root}/feature-manifest.json"; then
  naive_log="${RUNNER_TEMP:-/tmp}/singbox-naive.log"
  (
    cd "${asset_root}"
    "${binary}" -L 'http://127.0.0.1:18993' \
      -F 'naive+singbox://user:password@127.0.0.1:443?tls.enabled=true' \
      >"${naive_log}" 2>&1
  ) &
  naive_pid=$!
  sleep 2
  if ! kill -0 "${naive_pid}" 2>/dev/null; then
    cat "${naive_log}" >&2
    echo "Naive runtime exited during native asset smoke" >&2
    exit 1
  fi
  kill "${naive_pid}" 2>/dev/null || true
  wait "${naive_pid}" 2>/dev/null || true
fi
