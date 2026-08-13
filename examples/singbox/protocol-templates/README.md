# Sanitized native protocol templates

These templates are derived from configurations that were exercised against
real sing-box services, but contain no production hostname, address, UUID,
password, Reality key or certificate. They cover:

- VLESS Reality with Vision;
- Hysteria2;
- TUIC;
- ShadowTLS v3 over Shadowsocks 2022;
- Shadowsocks 2022;
- Trojan TLS;
- VLESS gRPC Reality;
- AnyTLS;
- VMess over WebSocket and TLS.

The paired `*-server.json` and `*-client.json` files use the same port,
identity, transport and authentication shape. Replace every `REPLACE_*` value,
the documentation domain, example UUIDs, the example Shadowsocks key and TLS
paths before deployment. A TLS certificate must cover the client-side
`server_name`. Reality private/public keys must be a genuine pair.

The two unavailable Cloudflare Tunnel configurations are intentionally absent.
Naive inbound remains documented elsewhere, but no Naive outbound template is
included because the embedded activation policy rejects its native IPC/
loopback bridge on every platform.

## 1. CLI only

Simple native objects can be written entirely in `-L` and `-F`. The commands
below show the shape; use shell-safe percent encoding for credentials in URI
userinfo.

```bash
# VLESS Reality Vision server
gust-with-singbox \
  -L 'vless+singbox://0.0.0.0:8443?users:=[{"uuid":"00000000-0000-4000-8000-000000000001","flow":"xtls-rprx-vision"}]&tls.enabled=true&tls.server_name=server.example.com&tls.reality.enabled=true&tls.reality.private_key=REPLACE_WITH_REALITY_PRIVATE_KEY&tls.reality.short_id:=["01234567"]&tls.reality.handshake.server=origin.example.com&tls.reality.handshake.server_port:=443' \
  -F 'direct://'

# VLESS Reality Vision client
gust-with-singbox -L 'socks5://127.0.0.1:1080?udp=true' \
  -F 'vless+singbox://00000000-0000-4000-8000-000000000001@server.example.com:8443?flow=xtls-rprx-vision&tls.enabled=true&tls.server_name=server.example.com&tls.utls.enabled=true&tls.utls.fingerprint=chrome&tls.reality.enabled=true&tls.reality.public_key=REPLACE_WITH_REALITY_PUBLIC_KEY&tls.reality.short_id=01234567'
```

```bash
# Hysteria2 server / client
gust-with-singbox \
  -L 'hysteria2+singbox://REPLACE_WITH_HYSTERIA2_PASSWORD@0.0.0.0:8444?tls.enabled=true&tls.alpn=h3&tls.certificate_path=/etc/gust/tls/server.crt&tls.key_path=/etc/gust/tls/server.key' \
  -F 'direct://'
gust-with-singbox -L 'socks5://127.0.0.1:1080?udp=true' \
  -F 'hysteria2+singbox://REPLACE_WITH_HYSTERIA2_PASSWORD@server.example.com:8444?up_mbps:=100&down_mbps:=300&tls.enabled=true&tls.server_name=server.example.com&tls.alpn=h3'
```

```bash
# TUIC server / client
gust-with-singbox \
  -L 'tuic+singbox://00000000-0000-4000-8000-000000000002:REPLACE_WITH_TUIC_PASSWORD@0.0.0.0:8445?congestion_control=bbr&tls.enabled=true&tls.alpn=h3&tls.certificate_path=/etc/gust/tls/server.crt&tls.key_path=/etc/gust/tls/server.key' \
  -F 'direct://'
gust-with-singbox -L 'socks5://127.0.0.1:1080?udp=true' \
  -F 'tuic+singbox://00000000-0000-4000-8000-000000000002:REPLACE_WITH_TUIC_PASSWORD@server.example.com:8445?congestion_control=bbr&udp_relay_mode=native&tls.enabled=true&tls.server_name=server.example.com&tls.alpn=h3'
```

```bash
# Shadowsocks 2022 server / client. Replace the example 16-byte Base64 key.
gust-with-singbox \
  -L 'ss+singbox://2022-blake3-aes-128-gcm:MDEyMzQ1Njc4OWFiY2RlZg%3D%3D@0.0.0.0:8388' \
  -F 'direct://'
gust-with-singbox -L 'socks5://127.0.0.1:1080?udp=true' \
  -F 'ss+singbox://2022-blake3-aes-128-gcm:MDEyMzQ1Njc4OWFiY2RlZg%3D%3D@server.example.com:8388'
```

```bash
# Trojan TLS server / client
gust-with-singbox \
  -L 'trojan+singbox://REPLACE_WITH_TROJAN_PASSWORD@0.0.0.0:8447?tls.enabled=true&tls.certificate_path=/etc/gust/tls/server.crt&tls.key_path=/etc/gust/tls/server.key' \
  -F 'direct://'
gust-with-singbox -L 'socks5://127.0.0.1:1080?udp=true' \
  -F 'trojan+singbox://REPLACE_WITH_TROJAN_PASSWORD@server.example.com:8447?tls.enabled=true&tls.server_name=server.example.com&tls.utls.enabled=true&tls.utls.fingerprint=chrome'
```

```bash
# VLESS gRPC Reality server / client
gust-with-singbox \
  -L 'vless+singbox://00000000-0000-4000-8000-000000000003@0.0.0.0:8448?tls.enabled=true&tls.server_name=server.example.com&tls.reality.enabled=true&tls.reality.private_key=REPLACE_WITH_REALITY_PRIVATE_KEY&tls.reality.short_id:=["89abcdef"]&tls.reality.handshake.server=origin.example.com&tls.reality.handshake.server_port:=443&transport.type=grpc&transport.service_name=example-grpc' \
  -F 'direct://'
gust-with-singbox -L 'socks5://127.0.0.1:1080?udp=true' \
  -F 'vless+singbox://00000000-0000-4000-8000-000000000003@server.example.com:8448?tls.enabled=true&tls.server_name=server.example.com&tls.utls.enabled=true&tls.utls.fingerprint=chrome&tls.reality.enabled=true&tls.reality.public_key=REPLACE_WITH_REALITY_PUBLIC_KEY&tls.reality.short_id=89abcdef&transport.type=grpc&transport.service_name=example-grpc'
```

```bash
# AnyTLS server / client
gust-with-singbox \
  -L 'anytls+singbox://REPLACE_WITH_ANYTLS_PASSWORD@0.0.0.0:8449?tls.enabled=true&tls.certificate_path=/etc/gust/tls/server.crt&tls.key_path=/etc/gust/tls/server.key' \
  -F 'direct://'
gust-with-singbox -L 'socks5://127.0.0.1:1080?udp=true' \
  -F 'anytls+singbox://REPLACE_WITH_ANYTLS_PASSWORD@server.example.com:8449?tls.enabled=true&tls.server_name=server.example.com&tls.utls.enabled=true&tls.utls.fingerprint=chrome'
```

```bash
# VMess WebSocket TLS server / client. Default VMess UDP works; XUDP is not required.
gust-with-singbox \
  -L 'vmess+singbox://00000000-0000-4000-8000-000000000004@0.0.0.0:8450?tls.enabled=true&tls.certificate_path=/etc/gust/tls/server.crt&tls.key_path=/etc/gust/tls/server.key&transport.type=ws&transport.path=%2Fvmess-example&transport.max_early_data:=2560&transport.early_data_header_name=Sec-WebSocket-Protocol' \
  -F 'direct://'
gust-with-singbox -L 'socks5://127.0.0.1:1080?udp=true' \
  -F 'vmess+singbox://00000000-0000-4000-8000-000000000004@server.example.com:8450?security=auto&tls.enabled=true&tls.server_name=server.example.com&transport.type=ws&transport.path=%2Fvmess-example&transport.headers.Host=server.example.com&transport.max_early_data:=2560&transport.early_data_header_name=Sec-WebSocket-Protocol'
```

ShadowTLS is deliberately not flattened into a single scalar URI. Its
Shadowsocks inner connection and ShadowTLS outer connection are two tagged
native objects with an explicit detour relationship; use the JSON form below.

## 2. CLI plus native JSON

For every ordinary pair, select the one-object file with `json=`:

```bash
gust-with-singbox \
  -L 'singbox://?json=./examples/singbox/protocol-templates/vmess-ws-tls-server.json' \
  -F 'direct://'

gust-with-singbox -L 'socks5://127.0.0.1:1080?udp=true' \
  -F 'singbox://?json=./examples/singbox/protocol-templates/vmess-ws-tls-client.json'
```

Replace `vmess-ws-tls` with `vless-reality`, `hysteria2`, `tuic`,
`shadowsocks-2022`, `trojan`, `vless-grpc-reality` or `anytls` to select the
other pairs.

ShadowTLS uses complete native configs because both tagged objects must live
in the same Box:

```bash
gust-with-singbox \
  -L 'singbox://?config=./examples/singbox/protocol-templates/shadowtls-server.json&inbound=shadowtls-in&activate_inbounds:=["shadowtls-in","shadowtls-inner"]' \
  -F 'direct://'

gust-with-singbox -L 'socks5://127.0.0.1:1080?udp=true' \
  -F 'singbox://?config=./examples/singbox/protocol-templates/shadowtls-client.json&outbound=proxy'
```

## 3. Gust JSON only

The two `gost-json-only-vmess-*.json` files contain the complete Gust service,
chain and embedded native options. They require no `-L` or `-F` flags:

```bash
gust-with-singbox -C ./examples/singbox/protocol-templates/gost-json-only-vmess-server.json
gust-with-singbox -C ./examples/singbox/protocol-templates/gost-json-only-vmess-client.json
```

The server and client JSON intentionally repeat the native options. This makes
the deployment self-contained and avoids a hidden dependency on a second JSON
file. Use the same `listener.metadata.protocol/options` and
`connector.metadata.protocol/options` shape to move any other pair into a
JSON-only Gust deployment.

## Validation and security notes

- `server.example.com` and `origin.example.com` are documentation domains, not
  runnable public endpoints.
- Generate new UUIDs, passwords, Shadowsocks keys, Reality keys and TLS keys;
  never deploy the example identities.
- Do not set `tls.insecure=true` in a server template. Install a certificate
  whose SAN matches `server_name`, use a private CA explicitly on clients, or
  add a verified SPKI pin.
- The source service configurations were inspected from the maintainer's
  private sing-box deployment and the matching final `-F` objects were checked
  against the live services from two independent VPS hosts. No private endpoint
  or credential is retained in these files.
