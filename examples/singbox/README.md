# Embedded sing-box examples

These files are readable templates for the `gust-with-singbox` release. Replace
every `REPLACE_*` value before deployment. They use only documentation
addresses and example domains.

Validate a CLI, JSON or mixed configuration without opening listeners or
system resources:

```bash
gust-with-singbox -singboxcheck \
  -L 'socks5://127.0.0.1:1080?udp=true' \
  -F 'singbox://?json=./examples/singbox/shadowsocks-uot-mux-outbound.json'
```

Use `json=` for one native inbound/outbound object and `config=` plus an exact
tag selector for a complete sing-box configuration. `-L` and `-F` accept the
same file, inline JSON and mixed path-override forms described by
`gust-with-singbox -singboxmanual`.

| File | Direction | Example invocation |
|---|---|---|
| `shadowsocks-uot-mux-outbound.json` | `-F` object | `-F 'singbox://?json=./examples/singbox/shadowsocks-uot-mux-outbound.json'` |
| `reality-server.json` | `-L` full config | `-L 'singbox://?config=./examples/singbox/reality-server.json&inbound=reality-in'` |
| `reality-client.json` | `-F` object | `-F 'singbox://?json=./examples/singbox/reality-client.json'` |
| `shadowtls-server.json` | `-L` full config | `-L 'singbox://?config=./examples/singbox/shadowtls-server.json&inbound=shadowtls-in&activate_inbounds:=["shadowtls-in","ss-inner"]'` |
| `remote-dns-outbound.json` | `-F` full config | `-F 'singbox://?config=./examples/singbox/remote-dns-outbound.json&outbound=proxy'` |
| `tun-inbound.json` | `-L` object | `-L 'singbox://?json=./examples/singbox/tun-inbound.json'` |
| `tproxy-inbound.json` | `-L` object | `-L 'singbox://?json=./examples/singbox/tproxy-inbound.json'` |
| `protocol-templates/` | paired `-L` / `-F` catalog | CLI-only, CLI+JSON and Gust JSON-only examples for nine live-validated protocol shapes |

Reality requires a genuinely paired X25519 private/public key, matching SNI,
short ID and UUID. TUN and TProxy require an isolated Linux network namespace,
the necessary capabilities, routing and firewall policy. Static validation
does not prove those runtime prerequisites.

The expanded [protocol template catalog](protocol-templates/README.md) is
sanitized from maintainer-private live configurations. It includes paired
server/client objects and shows three equivalent configuration surfaces without
publishing any real endpoint or credential.
