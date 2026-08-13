#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 /path/to/gust-with-singbox" >&2
  exit 2
fi

binary="$1"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
examples="${root}/examples/singbox"
templates="${examples}/protocol-templates"

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

for protocol in \
  vless-reality hysteria2 tuic shadowsocks-2022 trojan \
  vless-grpc-reality anytls vmess-ws-tls; do
  check -L "singbox://?json=${templates}/${protocol}-server.json" -F 'direct://'
  check -L 'socks5://127.0.0.1:1080?udp=true' \
    -F "singbox://?json=${templates}/${protocol}-client.json"
done

check \
  -L "singbox://?config=${templates}/shadowtls-server.json&inbound=shadowtls-in&activate_inbounds:=[\"shadowtls-in\",\"shadowtls-inner\"]" \
  -F 'direct://'
jq -e '
  [.inbounds[] | select(.tag == "shadowtls-inner")][0]
  | .listen == "127.0.0.1" and (.listen_port > 0)
' "${templates}/shadowtls-server.json" >/dev/null
check -L 'socks5://127.0.0.1:1080?udp=true' \
  -F "singbox://?config=${templates}/shadowtls-client.json&outbound=proxy"
check -C "${templates}/gost-json-only-vmess-server.json"
check -C "${templates}/gost-json-only-vmess-client.json"
check -C "${templates}/gost-json-only-nine-server.json"
check -C "${templates}/gost-json-only-nine-client.json"
jq -e '.services | length == 9' "${templates}/gost-json-only-nine-server.json" >/dev/null
jq -e '
  (.services | length == 9)
  and (.chains | length == 9)
' "${templates}/gost-json-only-nine-client.json" >/dev/null

# Keep representative CLI-only forms executable. In particular, Vision flow
# belongs to a VLESS inbound user, not the top-level inbound options.
check \
  -L 'vless+singbox://0.0.0.0:8443?users:=[{"uuid":"00000000-0000-4000-8000-000000000001","flow":"xtls-rprx-vision"}]&tls.enabled=true&tls.server_name=server.example.com&tls.reality.enabled=true&tls.reality.private_key=REPLACE_WITH_REALITY_PRIVATE_KEY&tls.reality.short_id:=["01234567"]&tls.reality.handshake.server=origin.example.com&tls.reality.handshake.server_port:=443' \
  -F 'direct://'
check -L 'socks5://127.0.0.1:1080?udp=true' \
  -F 'vmess+singbox://00000000-0000-4000-8000-000000000004@server.example.com:8450?security=auto&tls.enabled=true&tls.server_name=server.example.com&transport.type=ws&transport.path=%2Fvmess-example&transport.headers.Host=server.example.com&transport.max_early_data:=2560&transport.early_data_header_name=Sec-WebSocket-Protocol'

if "${binary}" -singboxcheck \
  -L 'socks5://127.0.0.1:1080' \
  -F 'ss+singbox://chacha20-ietf-poly1305:secret@proxy.example.com:8388?definitely_unknown_field=true' \
  >/dev/null 2>&1; then
  echo "sing-box static validation accepted an unknown native field" >&2
  exit 1
fi

echo "sing-box example static runtime validation PASS"
