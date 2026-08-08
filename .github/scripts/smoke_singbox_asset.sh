#!/usr/bin/env bash
set -euo pipefail

asset_root="${1:-}"
if [[ -z "${asset_root}" || ! -d "${asset_root}" ]]; then
  echo "usage: $0 ASSET_ROOT" >&2
  exit 2
fi
asset_root="$(cd "${asset_root}" && pwd)"

binary="${asset_root}/gust-with-singbox"
if [[ ! -f "${binary}" && -f "${binary}.exe" ]]; then
  binary="${binary}.exe"
fi
if [[ ! -f "${binary}" ]]; then
  echo "missing asset binary: ${binary}" >&2
  exit 1
fi

"${binary}" -V | tee "${asset_root}/smoke-version.txt"
grep -F 'flavor=singbox' "${asset_root}/smoke-version.txt"
"${binary}" -singboxmanual > "${asset_root}/smoke-singbox-manual.txt"
grep -F '# Gust embedded sing-box manual' "${asset_root}/smoke-singbox-manual.txt"
grep -F '## CLI configuration' "${asset_root}/smoke-singbox-manual.txt"
grep -F '## Gust JSON configuration' "${asset_root}/smoke-singbox-manual.txt"
grep -F '## Mixed CLI and JSON configuration' "${asset_root}/smoke-singbox-manual.txt"
test -f "${asset_root}/SINGBOX-MANUAL.md"
test -f "${asset_root}/SINGBOX-VALIDATION.md"
test -f "${asset_root}/SINGBOX-PERFORMANCE-BASELINE.json"
test -f "${asset_root}/SINGBOX-ARCHITECTURE.md"
test -f "${asset_root}/examples/singbox/README.md"
python3 -m json.tool "${asset_root}/SINGBOX-PERFORMANCE-BASELINE.json" >/dev/null
preflight="$("${binary}" -singboxcheck \
  -L 'socks5://127.0.0.1:1080?udp=true' \
  -F "singbox://?json=${asset_root}/examples/singbox/shadowsocks-uot-mux-outbound.json")"
grep -F 'sing-box configuration OK' <<<"${preflight}"
grep -F 'startup not attempted' <<<"${preflight}"

echo_log="${RUNNER_TEMP:-/tmp}/singbox-echo.log"
gust_log="${RUNNER_TEMP:-/tmp}/singbox-direct.log"
python3 tests/e2e/scripts/tcp_echo.py >"${echo_log}" 2>&1 &
echo_pid=$!
udp_log="${RUNNER_TEMP:-/tmp}/singbox-udp-echo.log"
dns_log="${RUNNER_TEMP:-/tmp}/singbox-dns.log"
python3 tests/e2e/scripts/udp_echo.py >"${udp_log}" 2>&1 &
udp_pid=$!
python3 tests/e2e/scripts/dns_responder.py --port 15353 >"${dns_log}" 2>&1 &
dns_pid=$!
(
  cd "${asset_root}"
  "${binary}" \
    -L 'http://127.0.0.1:18888' \
    -L 'socks5+singbox://user:password@127.0.0.1:18889' \
    -F 'direct+singbox://' >"${gust_log}" 2>&1
) &
gust_pid=$!
stop_processes() {
  if [[ $# -eq 0 ]]; then
    return
  fi
  kill "$@" 2>/dev/null || true
  # Git Bash cannot reliably deliver a graceful signal to every Windows
  # process. Bound shutdown before waiting so a Gust/Cronet child cannot keep
  # the native-smoke job alive after all assertions have passed.
  sleep 1
  kill -KILL "$@" 2>/dev/null || true
  wait "$@" 2>/dev/null || true
}
cleanup() {
  stop_processes "${gust_pid}" "${echo_pid}" "${udp_pid}" "${dns_pid}"
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
curl --fail --silent --show-error --max-time 2 \
  --proxy 'socks5h://user:password@127.0.0.1:18889' \
  http://127.0.0.1:5678 | grep -Fx 'hello-gost'
python3 tests/e2e/scripts/socks5_udp_client.py \
  --proxy 127.0.0.1:18889 --username user --password password \
  --target 127.0.0.1:5679 --payload hello-gost-udp \
  | grep -Fx 'hello-gost-udp'
python3 tests/e2e/scripts/socks5_udp_client.py \
  --proxy 127.0.0.1:18889 --username user --password password \
  --target 127.0.0.1:15353 --dns-name test.example.com \
  | grep -F 'DNS PASS:'

if python3 -c 'import json,sys; sys.exit("naive_outbound" in json.load(open(sys.argv[1], encoding="utf-8"))["unavailableFeatures"])' \
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
  stop_processes "${naive_pid}"
fi
