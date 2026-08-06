# Gust embedded sing-box manual

This manual applies to the `gost` command distributed by the Gust project and
to the `gust-with-singbox` release binary. If the binary was renamed to `gust`,
replace `gost` with `gust` in the examples.

Print this manual without network access:

```bash
gost -singboxmanual
```

## 1. Check the build before configuring it

```bash
gost -V
```

`flavor=singbox` means the embedded backend can run. `flavor=standard` can
parse a `+singbox` URI and render it with `-O`, but it cannot start the
embedded runtime. Download an asset whose name starts with
`gust-with-singbox` when runtime support is required.

The sing-box release assets contain a `feature-manifest.json`. On Linux and
Windows, keep the bundled `libcronet.so` or `libcronet.dll` beside the binary
when using Naive. The official Darwin build does not include Naive or CCM.

## 2. Configuration model

Gust uses two separate layers:

- `-L` creates a local GOST listener/service, such as an HTTP or SOCKS proxy.
- `-F` adds an outbound chain node. A URI whose scheme is `singbox` or ends in
  `+singbox` is handled by the embedded sing-box backend.

The embedded integration is an outbound/endpoint backend. It deliberately
does not turn sing-box `inbounds` into `-L` listeners. A full sing-box config
passed to `-F` must therefore contain no `inbounds`.

The usual local HTTP proxy routed through a sing-box outbound is:

```bash
gost \
  -L 'http://127.0.0.1:8080' \
  -F 'vless+singbox://UUID@proxy.example.com:443?tls.enabled=true&tls.server_name=proxy.example.com'
```

Use single quotes around node URIs in POSIX shells so `&`, `?`, brackets and
JSON are not interpreted by the shell. In PowerShell, single quotes work for
the same examples.

## CLI configuration

### URI syntax

```text
<protocol>+singbox://[userinfo@][server][:port][?options][#node-name]
singbox://?json=<node-object-or-file>
singbox://?config=<full-config-file>&outbound=<tag>
singbox://?config=<full-config-file>&endpoint=<tag>
```

Stable aliases include `ss` for `shadowsocks`, `socks4`, `socks4a`, `socks5`,
`hy2` for `hysteria2`, and `wg` for the WireGuard endpoint. Protocol and field
names follow the sing-box v1.13.16 outbound/endpoint schema.

Common examples:

```bash
# Shadowsocks
gost -L 'socks5://127.0.0.1:1080' \
  -F 'ss+singbox://chacha20-ietf-poly1305:SECRET@proxy.example.com:8388'

# SOCKS5 upstream, TCP and UDP
gost -L 'socks5://127.0.0.1:1080' \
  -F 'socks5+singbox://USER:PASSWORD@proxy.example.com:1080?network=tcp,udp'

# HTTP upstream
gost -L 'http://127.0.0.1:8080' \
  -F 'http+singbox://USER:PASSWORD@proxy.example.com:3128'

# VMess over WebSocket
gost -L 'socks5://127.0.0.1:1080' \
  -F 'vmess+singbox://UUID@proxy.example.com:443?security=auto&tls.enabled=true&tls.server_name=proxy.example.com&transport.type=ws&transport.path=%2Fvmess'

# VLESS over TLS and WebSocket
gost -L 'socks5://127.0.0.1:1080' \
  -F 'vless+singbox://UUID@proxy.example.com:443?tls.enabled=true&tls.server_name=proxy.example.com&transport.type=ws&transport.path=%2Fvless'

# Trojan TLS
gost -L 'http://127.0.0.1:8080' \
  -F 'trojan+singbox://PASSWORD@proxy.example.com:443?tls.enabled=true&tls.server_name=proxy.example.com'

# Hysteria2
gost -L 'socks5://127.0.0.1:1080' \
  -F 'hysteria2+singbox://PASSWORD@proxy.example.com:443?up_mbps=100&down_mbps=300&tls.enabled=true&tls.server_name=proxy.example.com'

# TUIC
gost -L 'socks5://127.0.0.1:1080' \
  -F 'tuic+singbox://UUID:PASSWORD@proxy.example.com:443?congestion_control=bbr&tls.enabled=true&tls.server_name=proxy.example.com&tls.alpn=h3'

# AnyTLS
gost -L 'socks5://127.0.0.1:1080' \
  -F 'anytls+singbox://PASSWORD@proxy.example.com:443?tls.enabled=true&tls.server_name=proxy.example.com'

# SSH
gost -L 'http://127.0.0.1:8080' \
  -F 'ssh+singbox://USER:PASSWORD@proxy.example.com:22'
```

Percent-encode reserved characters in user information and paths. For
example, `@` in a password becomes `%40`, `/ws` can be written `%2Fws`, and a
literal comma becomes `%2C`.

### Nested values and exact JSON values

Nested fields use dot paths. Object arrays use indexes:

```text
tls.enabled=true
tls.server_name=proxy.example.com
transport.type=grpc
transport.service_name=proxy
peers[0].port=51820
peers[0].allowed_ips=0.0.0.0/0,::/0
```

In a singbox flavor, plain `path=value` is converted with the pinned native
schema. Use `path:=JSON` when an exact JSON type is important or when producing
identical output with a standard flavor:

```text
tls.enabled:=true
tls.alpn:=["h3","h2"]
transport.headers:={"Host":["cdn.example.com"]}
tls.server_name:=null
```

Arrays may be comma-separated or repeated. The first CLI assignment replaces
the lower-priority array; later repeated assignments append in URI order.

### Reality

Both ordinary VLESS Reality and `xtls-rprx-vision` are supported. The client
must use the server's actual Reality public key, short ID and SNI:

```bash
gost -L 'socks5://127.0.0.1:1080' \
  -F 'vless+singbox://UUID@proxy.example.com:443?flow=xtls-rprx-vision&tls.enabled=true&tls.server_name=reality.example.com&tls.utls.enabled=true&tls.utls.fingerprint=chrome&tls.reality.enabled=true&tls.reality.public_key=PUBLIC_KEY&tls.reality.short_id=SHORT_ID'
```

Omit `flow=xtls-rprx-vision` for basic Reality. A TCP connect to the server
port is not a Reality test; verify traffic through the local `-L` proxy.

## Node JSON configuration

`json=` accepts either an inline sing-box outbound/endpoint object or a file
path. `json64=` accepts an unpadded URL-safe base64 object.

`vless-node.json`:

```json
{
  "type": "vless",
  "server": "proxy.example.com",
  "server_port": 443,
  "uuid": "00000000-0000-4000-8000-000000000001",
  "tls": {
    "enabled": true,
    "server_name": "proxy.example.com"
  },
  "transport": {
    "type": "ws",
    "path": "/vless"
  }
}
```

Run it with:

```bash
gost -L 'socks5://127.0.0.1:1080' \
  -F 'singbox://?json=./vless-node.json#edge'
```

An inline object is also valid:

```bash
gost -L 'http://127.0.0.1:8080' \
  -F 'singbox://?json={"type":"socks","server":"127.0.0.1","server_port":1081,"version":"5"}'
```

## Complete sing-box configuration

Use a complete config when the selected outbound depends on DNS, route rules,
`selector`/`urltest`, `detour`, services, or endpoints:

```bash
gost -L 'socks5://127.0.0.1:1080' \
  -F 'singbox://?config=/etc/sing-box/config.json&outbound=proxy'
```

For a WireGuard or Tailscale endpoint, select it with `endpoint=`:

```bash
gost -L 'socks5://127.0.0.1:1080' \
  -F 'singbox://?config=/etc/sing-box/config.json&endpoint=tailnet'
```

`config64=` is the base64url equivalent. Only one of `outbound=` and
`endpoint=` may be present. The chosen node's complete dependency graph is
kept. Configs containing `inbounds` are rejected.

## Gust JSON configuration

Do not confuse a Gust config passed with `-C` with the sing-box node JSON
passed by `json=`. A complete Gust JSON configuration looks like this:

```json
{
  "services": [
    {
      "name": "local-socks",
      "addr": "127.0.0.1:1080",
      "handler": {"type": "socks5", "chain": "singbox-chain"},
      "listener": {"type": "tcp"}
    }
  ],
  "chains": [
    {
      "name": "singbox-chain",
      "hops": [
        {
          "name": "edge-hop",
          "nodes": [
            {
              "name": "edge",
              "addr": "proxy.example.com:443",
              "connector": {
                "type": "singbox",
                "metadata": {
                  "protocol": "vless",
                  "options": {
                    "uuid": "00000000-0000-4000-8000-000000000001",
                    "tls": {
                      "enabled": true,
                      "server_name": "proxy.example.com"
                    }
                  }
                }
              },
              "dialer": {"type": "direct"}
            }
          ]
        }
      ]
    }
  ]
}
```

Start it with:

```bash
gost -C ./gost.json
```

Inside `connector.metadata`, use one of these forms:

```json
{"protocol":"hysteria2","options":{"server":"proxy.example.com","server_port":443,"password":"SECRET","tls":{"enabled":true,"server_name":"proxy.example.com"}}}
```

```json
{"json":"/etc/gost/nodes/proxy.json"}
```

```json
{"config":"/etc/sing-box/config.json","outbound":"proxy"}
```

The same structure is accepted in YAML; JSON is shown here to make every
object, array, number and boolean type explicit.

## Mixed CLI and JSON configuration

There are two supported kinds of mixing.

First, combine a Gust config source with CLI services and nodes. Config sources
are loaded left-to-right, then CLI objects are appended:

```bash
gost \
  -C '{"log":{"level":"info"}}' \
  -L 'socks5://127.0.0.1:1080' \
  -F 'vless+singbox://UUID@proxy.example.com:443?tls.enabled=true&tls.server_name=proxy.example.com'
```

If a service is already in a `-C` file and the chain is supplied only by
`-F`, point that service's handler at `chain-0`, the chain name generated for
CLI nodes.

Second, use node JSON or a complete sing-box config as a base and override it
from the URI:

```bash
gost -L 'socks5://127.0.0.1:1080' \
  -F 'vless+singbox://NEW_UUID@new.example.com:443?json=./vless-node.json&tls.server_name=new.example.com'
```

```bash
gost -L 'socks5://127.0.0.1:1080' \
  -F 'singbox://override.example.com:443?config=./sing-box.json&outbound=proxy&json=./node-overlay.json&tls.server_name=override.example.com'
```

Merge priority is fixed and does not depend on query order:

```text
selected node from config/config64
  < json/json64 node overlay
  < URI authority and userinfo
  < ordinary query assignments
```

Use `-O json` or `-O yaml` to inspect the merged Gust configuration before
running it. The output can contain credentials, so do not publish it:

```bash
gost -L 'socks5://127.0.0.1:1080' -F 'singbox://?json=./node.json' -O json
```

## Proxy chains

Repeat `-F` to build a multi-hop chain in command-line order:

```bash
gost -L 'http://127.0.0.1:8080' \
  -F 'socks5://first.example.com:1080' \
  -F 'vless+singbox://UUID@second.example.com:443?tls.enabled=true&tls.server_name=second.example.com' \
  -F 'direct+singbox://'
```

A sing-box node can be first, middle, or last. When it follows another GOST
node, its network leaf uses the request-scoped GOST prefix route. A prefix
failure is fail-closed and does not silently fall back to the system network.

A complete sing-box config may contain its own `detour`, selector and DNS
dependencies. Outbound and endpoint leaves without a user detour observe the
preceding GOST route. Networked DNS transport leaves without a user detour are
injected in the same way, so DNS used to resolve the proxy server also follows
the request-scoped GOST prefix. An explicit DNS `detour` is preserved.

## TCP, UDP and DNS usage

TCP works through ordinary HTTP/SOCKS local services. For applications that
need UDP, use a UDP-capable local handler such as SOCKS5 and an outbound that
supports UDP. Hysteria2 and TUIC are validated for both TCP and UDP data paths.

For proxy outbounds, the UDP `net.Conn` adapter passes a destination hostname
to sing-box as a SOCKS domain address without resolving it locally. This allows
remote-only UDP target names when the selected protocol supports domain
destinations. Direct UDP still uses local DNS because it has no remote resolver.

DNS can be handled in either layer:

- Configure a GOST DNS service/resolver in the Gust `-C` file.
- Load a full sing-box config whose selected outbound depends on its sing-box
  DNS configuration. A networked sing-box DNS server with no explicit detour
  automatically follows a preceding GOST node.

For example, this remote DNS transport is attached directly to the preceding
GOST route; no placeholder outbound is required:

```json
{
  "dns": {
    "servers": [
      {"type": "udp", "tag": "remote-dns", "server": "1.1.1.1"}
    ]
  },
  "outbounds": [
    {
      "type": "socks",
      "tag": "proxy",
      "server": "proxy.example.com",
      "server_port": 1080,
      "domain_resolver": "remote-dns"
    }
  ]
}
```

When testing DNS, send an actual DNS query through the configured path. A TCP
connection to port 53 does not validate UDP DNS, and a single public resolver
failure does not prove that the proxy transport is broken.

## Validation and troubleshooting

1. Run `gost -V` and require `flavor=singbox`.
2. Check `feature-manifest.json` for the current platform.
3. Render configuration with `-O json` and inspect node type, tag and field
   types without publishing secrets.
4. Test through the local `-L` proxy to a controlled TCP/UDP/DNS target.
5. For TLS and Reality, verify SNI, ALPN, public key, short ID, UUID and flow.
6. For Naive, verify the Cronet library is beside the executable.
7. Use `-D` for debug logs. Configuration errors identify field paths but
   intentionally redact passwords, private keys, tokens and complete JSON.

Identical canonical configurations share a runtime. Reload starts the new
runtime before retiring the old one, and active TCP/UDP leases keep the old
runtime alive until the connection closes.

The validated schema is pinned to sing-box v1.13.16. Fields introduced by a
newer sing-box release are not automatically accepted. `Bind` and implicit
sing-box inbound startup are not part of this outbound integration.

The repository's detailed acceptance evidence is in
`SINGBOX-VALIDATION.md`. Release archives include this manual, the validation
record, `feature-manifest.json`, license notices and exact source revisions.
