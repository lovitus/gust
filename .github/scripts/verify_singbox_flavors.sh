#!/usr/bin/env bash
set -euo pipefail

if [[ ! -d ../gust-x ]]; then
  echo 'gust-x must be checked out next to gust' >&2
  exit 1
fi

runner_temp="${RUNNER_TEMP:-$(mktemp -d)}"
tags="$(bash .github/scripts/singbox-tags.sh)"
standard_bin="${runner_temp}/gust-standard"
singbox_bin="${runner_temp}/gust-with-singbox"

go build -trimpath -o "${standard_bin}" ./cmd/gost
go build -trimpath -tags "${tags}" -o "${singbox_bin}" ./cmd/gost

"${standard_bin}" -V | tee "${runner_temp}/standard-version.txt"
"${singbox_bin}" -V | tee "${runner_temp}/singbox-version.txt"
grep -F 'flavor=standard' "${runner_temp}/standard-version.txt"
grep -F 'flavor=singbox' "${runner_temp}/singbox-version.txt"

for binary in "${standard_bin}" "${singbox_bin}"; do
  "${binary}" -singboxmanual > "${runner_temp}/singbox-manual.txt"
  grep -F '# Gust embedded sing-box manual' "${runner_temp}/singbox-manual.txt"
  grep -F '## CLI configuration' "${runner_temp}/singbox-manual.txt"
  grep -F '## Gust JSON configuration' "${runner_temp}/singbox-manual.txt"
  grep -F '## Mixed CLI and JSON configuration' "${runner_temp}/singbox-manual.txt"
done

if go version -m "${standard_bin}" | grep -F $'github.com/sagernet/sing-box\t'; then
  echo 'standard binary unexpectedly links sing-box' >&2
  exit 1
fi
go version -m "${singbox_bin}" | grep -F $'github.com/sagernet/sing-box\t'

python3 tests/e2e/scripts/tcp_echo.py >"${runner_temp}/echo.log" 2>&1 &
echo_pid=$!
(
  cd "${runner_temp}"
  "${singbox_bin}" \
    -L 'http://127.0.0.1:18888' \
    -L 'socks5+singbox://user:password@127.0.0.1:18889' \
    -L 'http+singbox://user:password@127.0.0.1:18890' \
    -L 'singbox://?json={"type":"mixed","listen":"127.0.0.1","listen_port":18891}' \
    -F 'direct+singbox://' >"${runner_temp}/singbox-e2e.log" 2>&1
) &
gust_pid=$!
# shellcheck disable=SC2329 # invoked by trap
cleanup() {
  kill "${gust_pid}" "${echo_pid}" 2>/dev/null || true
  wait "${gust_pid}" "${echo_pid}" 2>/dev/null || true
}
trap cleanup EXIT

ready=false
for _ in {1..30}; do
  if curl --fail --silent --show-error --max-time 2 \
    --proxy http://127.0.0.1:18888 http://127.0.0.1:5678 \
    | grep -Fx 'hello-gost'; then
    ready=true
    break
  fi
  sleep 0.2
done
if [[ "${ready}" != true ]]; then
  cat "${runner_temp}/singbox-e2e.log" >&2
  exit 1
fi
curl --fail --silent --show-error --max-time 2 \
  --proxy 'socks5h://user:password@127.0.0.1:18889' \
  http://127.0.0.1:5678 | grep -Fx 'hello-gost'
curl --fail --silent --show-error --max-time 2 \
  --proxy 'http://user:password@127.0.0.1:18890' \
  http://127.0.0.1:5678 | grep -Fx 'hello-gost'
curl --fail --silent --show-error --max-time 2 \
  --proxy 'socks5h://127.0.0.1:18891' \
  http://127.0.0.1:5678 | grep -Fx 'hello-gost'
