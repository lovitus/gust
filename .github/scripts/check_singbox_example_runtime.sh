#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 /path/to/gust-with-singbox" >&2
  exit 2
fi

binary="$1"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
examples="${root}/examples/singbox"

check() {
  local output
  output="$("${binary}" -singboxcheck "$@")"
  grep -q '^sing-box configuration OK$' <<<"${output}"
  grep -q '^startup not attempted; sockets and system resources were not opened$' <<<"${output}"
  if grep -Eq 'REPLACE_|example-user|proxy\.example\.com|resolver\.example\.com' <<<"${output}"; then
    echo "sing-box safe validation output exposed a configuration value" >&2
    exit 1
  fi
}

check -L 'socks5://127.0.0.1:1080?udp=true' \
  -F "singbox://?json=${examples}/shadowsocks-uot-mux-outbound.json"
check -L "singbox://?config=${examples}/reality-server.json&inbound=reality-in" \
  -F 'direct://'
check -L 'socks5://127.0.0.1:1080' \
  -F "singbox://?json=${examples}/reality-client.json"
check -L "singbox://?config=${examples}/shadowtls-server.json&inbound=shadowtls-in&activate_inbounds:=[\"shadowtls-in\",\"ss-inner\"]" \
  -F 'direct://'
check -L 'socks5://127.0.0.1:1080' \
  -F "singbox://?config=${examples}/remote-dns-outbound.json&outbound=proxy"
check -L "singbox://?json=${examples}/tun-inbound.json" -F 'direct://'
check -L "singbox://?json=${examples}/tproxy-inbound.json" -F 'direct://'

if "${binary}" -singboxcheck \
  -L 'socks5://127.0.0.1:1080' \
  -F 'ss+singbox://chacha20-ietf-poly1305:secret@proxy.example.com:8388?definitely_unknown_field=true' \
  >/dev/null 2>&1; then
  echo "sing-box static validation accepted an unknown native field" >&2
  exit 1
fi

echo "sing-box example static runtime validation PASS"
