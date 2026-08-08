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
when using a Naive **outbound**. Darwin does not include the Cronet-based Naive
outbound or CCM; a Naive inbound is still available because its server path
does not use Cronet.

## 2. Configuration model

Gust uses the same adapter syntax in both directions:

- `-L '<protocol>+singbox://...'` creates a native sing-box inbound and hands
  its unmatched traffic to the ordinary GOST chain.
- `-F '<protocol>+singbox://...'` adds a native sing-box outbound/endpoint to
  that chain.
- A plain `-L` or `-F` remains the standard GOST implementation. Adding
  `+singbox` is always explicit and never changes the other flavor silently.

The usual local HTTP proxy routed through a sing-box outbound is:

```bash
gost \
  -L 'http://127.0.0.1:8080' \
  -F 'vless+singbox://UUID@proxy.example.com:443?tls.enabled=true&tls.server_name=proxy.example.com'
```

Use single quotes around node URIs in POSIX shells so `&`, `?`, brackets and
JSON are not interpreted by the shell. In PowerShell, single quotes work for
the same examples.

## Architecture and request path

The integration embeds sing-box as a Go library in the Gust process. It does
not execute a `sing-box` child process and does not create a hidden localhost
SOCKS/HTTP bridge. The two normal data paths are:

```text
application
  -> Gust -L listener and handler
  -> preceding Gust -F nodes, if any
  -> selected embedded sing-box outbound or endpoint
  -> destination

native client
  -> embedded sing-box -L inbound
  -> internal __gust_egress adapter
  -> Gust -F chain, if any
  -> destination
```

For `-F`, Gust owns the listener, service routing and chain selection while
sing-box owns the selected outbound protocol and its native dependencies. For
`-L`, sing-box owns the listening socket and protocol handshake while Gust owns
the final chain. Explicit native `route.rules` are preserved and take priority;
unmatched/default traffic enters `__gust_egress`. A full `-L` config whose
`route.final` names a user outbound is rejected instead of being silently
overwritten. Convert that choice to an explicit rule when it must remain native.

At configuration time, `-L` and `-F` share the URI lexer, readable JSON/file
loader, merge rules and typed path assignments. Direction-specific native
schemas prevent an outbound-only field from being accepted on an inbound (and
vice versa). Outbound Boxes use the canonical runtime pool and reference-counted
leases. Each inbound service owns one Box because it owns sockets and system
resources; reload starts its replacement first and closes the retired service
after handoff.

## CLI configuration

### URI syntax

```text
<protocol>+singbox://[userinfo@][server][:port][?options][#node-name]
singbox://?json=<node-object-or-file>
singbox://?config=<full-config-file>&inbound=<tag>      # with -L
singbox://?config=<full-config-file>&outbound=<tag>     # with -F
singbox://?config=<full-config-file>&endpoint=<tag>     # with -F
```

Stable aliases include `ss` for `shadowsocks`, `socks4`, `socks4a`, `socks5`,
`hy2` for `hysteria2`, and `wg` for the WireGuard endpoint. Protocol and field
names follow the sing-box v1.13.16 schema selected by direction.

### Native inbound (`-L`)

The simplest inbound form is intentionally identical to `-F`: the URI
authority supplies the listening address, user information supplies the
protocol credentials, and query paths supply native options.

```bash
# SOCKS5 server; unmatched TCP and UDP use the following four-hop chain
gost \
  -L 'socks5+singbox://user:password@0.0.0.0:1080' \
  -F 'socks5://first.example.com:1080' \
  -F 'ss+singbox://METHOD:PASSWORD@second.example.com:8388' \
  -F 'ssh://user:password@third.example.com:22' \
  -F 'hy2+singbox://password@fourth.example.com:443?tls.enabled=true&tls.server_name=fourth.example.com'

# Shadowsocks server
gost -L 'ss+singbox://chacha20-ietf-poly1305:SECRET@0.0.0.0:8388'

# VLESS Reality server (use its native private key and handshake target)
gost -L 'vless+singbox://UUID@0.0.0.0:443?tls.enabled=true&tls.server_name=example.com&tls.reality.enabled=true&tls.reality.private_key=PRIVATE_KEY&tls.reality.short_id=SHORT_ID&tls.reality.handshake.server=example.com&tls.reality.handshake.server_port:=443'

# TUN, REDIRECT and TProxy are Linux/system-resource listeners
gost -L 'tun+singbox://?interface_name=gust0&address:=["192.0.2.1/30"]&stack=gvisor'
gost -L 'redirect+singbox://0.0.0.0:12345'
gost -L 'tproxy+singbox://0.0.0.0:12345?network:=["tcp","udp"]'
```

The same form applies to `http`, `mixed`, `vmess`, `vless`, `trojan`,
`anytls`, `hysteria`, `hysteria2`, `tuic`, `naive` and `direct`.
ShadowTLS is normally selected from a full config together with its explicit
detour inbound. TLS, Reality, QUIC, transport, multiplex and (where supported
by that direction) UoT fields remain native sing-box fields; use a JSON object
when their CLI form would be difficult to review.

Each repeated `-L` creates an independent service-owned Box. For example,
three `-L` arguments followed by four `-F` arguments mean three independent
listeners, each referring to the same generated `chain-0` containing those
four nodes. The chain configuration is reused; the listener Boxes are not
automatically combined. This is the correctness and lifecycle-isolation
baseline, not an accidental duplicate runtime.

The native inbound must use a fixed non-zero port. TUN has a stable synthetic
address. Unknown resource-owning types, implicit API/service listeners and
resource behavior that cannot be certified are rejected fail-closed before
startup.

### Native outbound (`-F`)

Common examples:

```bash
# Shadowsocks
gost -L 'socks5://127.0.0.1:1080' \
  -F 'ss+singbox://chacha20-ietf-poly1305:SECRET@proxy.example.com:8388'

# SOCKS5 upstream, TCP and UDP
gost -L 'socks5://127.0.0.1:1080?udp=true' \
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

### Certified protocol quick reference

`PASS` below means that a real native peer exchanged and verified application
payloads. It never means only that a URI parsed or a port opened.

| Native `-L` inbound | TCP | UDP | Important requirement |
|---|---:|---:|---|
| Shadowsocks, SOCKS, Mixed | PASS | PASS | Match method/users; enable UDP in the client path |
| HTTP | PASS | n/a | CONNECT authentication is native |
| VMess, VLESS | PASS | PASS | Match UUID and transport options |
| Trojan | PASS | PASS | Valid certificate/SNI and password |
| AnyTLS | PASS | PASS | TLS; packet path uses UoT |
| Hysteria, Hysteria2, TUIC | PASS | PASS | QUIC/UDP reachability and valid TLS |
| Naive | PASS | n/a | HTTP/2 CONNECT and Naive framing |
| Direct | PASS | PASS | Explicit destination override |
| VLESS Reality / Vision | PASS | n/a in certified profiles | Public/private key, SNI, short ID and flow must match |
| ShadowTLS | PASS | detour-dependent | Exact inbound detour activation set |
| TUN | PASS | PASS | Linux network privilege, interface and routes |
| REDIRECT | PASS | n/a | Linux firewall original-destination support |
| TProxy | PASS | PASS | Linux policy routing and TPROXY target |

| Native `-F` outbound/endpoint | TCP | UDP | Important requirement |
|---|---:|---:|---|
| Shadowsocks | PASS | PASS | Matching AEAD method/key |
| SOCKS4 / SOCKS4a | PASS | n/a | IPv4/domain target form |
| SOCKS5 | PASS | PASS | Upstream UDP association support |
| HTTP | PASS | n/a | CONNECT upstream |
| VMess, VLESS, Trojan | PASS | PASS | Matching peer, transport and security fields |
| Reality / Vision | PASS | n/a in certified profiles | Key/SNI/short ID/flow must match |
| AnyTLS | PASS | PASS | TLS and UoT support |
| Hysteria2, TUIC | PASS | PASS | QUIC/UDP reachability |
| SSH, Naive | PASS | n/a | Native protocol handshake |
| WireGuard endpoint | PASS | PASS | Select with `endpoint=` |
| Direct | PASS | PASS | Baseline and explicit native route |

ShadowsocksR is a removed sing-box compatibility stub and is rejected. A
protocol marked `n/a` does not gain UDP support from Gust. Platform packaging
also matters: Darwin lacks the Cronet Naive outbound and CCM, while the Naive
inbound remains available. Inspect `feature-manifest.json` instead of assuming
that a type compiled for one platform exists on another.

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

`json=` accepts either an inline native object or a JSON file path. Its
direction comes from `-L` or `-F`; users do not need a second syntax for
inbounds. Readable JSON is recommended. Base64 source controls remain only for
backward compatibility and are deliberately not used in this manual.

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

An inbound object works the same way. For example, a Shadowsocks server with
native inbound multiplexing can remain readable:

```json
{
  "type": "shadowsocks",
  "listen": "0.0.0.0",
  "listen_port": 8388,
  "method": "2022-blake3-aes-128-gcm",
  "password": "REPLACE_WITH_KEY",
  "multiplex": {
    "enabled": true
  }
}
```

```bash
gost -L 'singbox://?json=./ss-inbound.json' -F 'direct://'
```

Direction still matters. In sing-box v1.13.16, Shadowsocks `plugin`,
`plugin_opts` and `udp_over_tcp` are outbound fields, while inbound
`multiplex` is a different server schema. A client node containing plugin,
UoT v2 and multiplex settings should therefore be used unchanged with `-F`:

```json
{
  "type": "shadowsocks",
  "server": "proxy.example.com",
  "server_port": 8388,
  "method": "2022-blake3-aes-128-gcm",
  "password": "REPLACE_WITH_KEY",
  "plugin": "obfs-local",
  "plugin_opts": "obfs=http;obfs-host=cdn.example.com",
  "udp_over_tcp": {"enabled": true, "version": 2},
  "multiplex": {"enabled": true, "protocol": "h2mux"}
}
```

```bash
gost -L 'socks5://127.0.0.1:1080?udp=true' \
  -F 'singbox://?json=./ss-outbound.json'
```

If those outbound-only keys are supplied to `-L`, native schema validation
reports the exact invalid path; Gust does not guess, drop or reinterpret them.

Inline JSON is portable and readable when shell quoting remains manageable:

```bash
gost -L 'singbox://?json={"type":"socks","listen":"127.0.0.1","listen_port":1080}' \
  -F 'direct+singbox://'
```

An inline object is also valid:

```bash
gost -L 'http://127.0.0.1:8080' \
  -F 'singbox://?json={"type":"socks","server":"127.0.0.1","server_port":1081,"version":"5"}'
```

## Complete sing-box configuration

Use a complete config when the selected object depends on DNS,
`selector`/`urltest`, `detour`, services, endpoints or other tagged objects.
The selector name follows direction:

```bash
gost -L 'socks5://127.0.0.1:1080' \
  -F 'singbox://?config=/etc/sing-box/config.json&outbound=proxy'
```

```bash
gost -L 'singbox://?config=/etc/sing-box/server.json&inbound=entry' \
  -F 'direct://'
```

For a WireGuard or Tailscale endpoint, select it with `endpoint=`:

```bash
gost -L 'socks5://127.0.0.1:1080' \
  -F 'singbox://?config=/etc/sing-box/config.json&endpoint=tailnet'
```

With `-F`, only one of `outbound=` and `endpoint=` may be present; native
inbounds are not activated. With `-L`, `inbound=` is required. Only that
inbound is activated unless it depends on another inbound through the native
top-level `detour` / `ListenOptions.detour` field.

Such a dependency must be explicit and exact:

```bash
gost -L 'singbox://?config=./shadowtls-server.json&inbound=entry&activate_inbounds:=["entry","inner"]'
```

`activate_inbounds` is a Gust adapter control, not a sing-box JSON field. It
requires exact `:=JSON` array syntax, is removed before native decoding, and
must equal the selected inbound's complete detour closure. Missing, unrelated,
duplicate, cyclic or unknown tags fail before startup; Gust never starts every
inbound in a supplied file by surprise.

## Gust JSON configuration

Do not confuse a Gust config passed with `-C` with the sing-box node JSON
passed by `json=`. A native inbound is a top-level Gust service whose
listener type is `singbox`; the handler's chain is the ordinary GOST chain:

```json
{
  "services": [
    {
      "name": "native-socks",
      "addr": "127.0.0.1:1080",
      "listener": {
        "type": "singbox",
        "metadata": {
          "protocol": "socks",
          "options": {
            "users": [{"username": "user", "password": "password"}]
          }
        }
      },
      "handler": {"type": "auto", "chain": "singbox-chain"}
    }
  ],
  "chains": [{"name": "singbox-chain", "hops": []}]
}
```

For a full native server config, listener metadata uses the same controls as
the CLI:

```json
{
  "config": "/etc/sing-box/server.json",
  "inbound": "entry",
  "activate_inbounds": ["entry", "inner"]
}
```

`addr`, when present, overrides the selected native inbound's `listen` and
`listen_port`. Put authentication, TLS, admission-like protocol controls and
other native listener behavior inside the native options. Generic GOST listener
authentication/TLS, handler metadata, limiters, observers, rewrites and similar
wrappers are rejected for this service instead of being accepted but bypassed.

An outbound-only complete Gust JSON configuration looks like this:

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
selected node from config
  < json node overlay
  < URI authority and userinfo
  < ordinary query assignments
```

Use `-O json` or `-O yaml` to inspect the merged Gust configuration before
running it. The output can contain credentials, so do not publish it:

```bash
gost -L 'socks5://127.0.0.1:1080' -F 'singbox://?json=./node.json' -O json
```

### What configuration output preserves

`-O` renders the effective Gust configuration, not the original sing-box text.
For a singbox flavor, native validation canonicalizes the node and complete
config first. JSON/YAML whitespace, object key order and source-file layout are
therefore not preserved; `json=` and `config=` require strict JSON, so source
comments are not accepted or preserved. Aliases are expanded to canonical
protocol names, native field types are retained and the result is deterministic
for the same effective input. In a standard flavor, native validation is
deferred, so use `path:=JSON` when booleans, numbers, arrays or null must retain
an exact type across both flavors.

Relative `json=` and `config=` paths are resolved against the process working
directory when the command/config is parsed, and each file source is limited
to 16 MiB. The rendered metadata contains the merged object, not a byte-for-byte
reference round trip. Treat `-O` output as a secret-bearing generated config,
not as a formatter for the source file.

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

Internally, a sing-box node is a self-dialing GOST transport. When chain
construction reaches one, Gust moves all preceding nodes into that transport's
prefix route and starts a new route segment. The sing-box node then opens the
next hop or final destination itself. If another GOST node follows it, sing-box
dials that node's address and the following node performs its normal handshake.
Two sing-box nodes work the same way recursively: the second receives a prefix
whose first self-dialing node already contains the earlier prefix. This avoids
an extra local proxy socket while preserving command-line `-F` order.

A complete sing-box config may contain its own `detour`, selector and DNS
dependencies. Outbound and endpoint leaves without a user detour observe the
preceding GOST route. Networked DNS transport leaves without a user detour are
injected in the same way, so DNS used to resolve the proxy server also follows
the request-scoped GOST prefix. An explicit DNS `detour` is preserved.

For a native `-L`, the direction is reversed at the internal boundary:
sing-box accepts and decrypts the connection, then `__gust_egress` asks the
service's stable GOST Router for `tcp` or an unconnected `udp` association.
That Router holds the chain named by the service handler. TCP is streamed
directly; UDP keeps each packet's destination address and domain name, so one
SOCKS association may carry multiple targets. Missing routing context, prefix
failure or an unknown target fails closed and never falls back to a system
direct connection.

## TCP, UDP and DNS usage

TCP works through ordinary HTTP/SOCKS local services. For applications that
need UDP, use a UDP-capable local handler such as SOCKS5 and an outbound that
supports UDP. Hysteria2 and TUIC are validated for both TCP and UDP data paths.

GOST SOCKS5 listeners do not enable UDP relay by default. Add `udp=true` to
the local listener when clients use SOCKS5 UDP ASSOCIATE:

```bash
gost -L 'socks5://127.0.0.1:1080?udp=true' \
  -F 'socks5+singbox://user:password@proxy.example.com:1080?network=tcp,udp'
```

On a server with a restrictive firewall, reserve and allow an explicit relay
range, for example
`?udp=true&udp.minPort=20000&udp.maxPort=20100`. The TCP listener port and the
chosen UDP relay range must both be reachable by the client. An embedded
sing-box outbound accepts both connected UDP destinations and the unconnected
packet connection used by GOST's SOCKS5 UDP ASSOCIATE handler.

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
      {"type": "udp", "tag": "remote-dns", "server": "dns.example.com"}
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

## User acceptance recipe

Test the exact packaged binary and configuration that will be deployed. Keep a
direct control result from the same client and target so a target outage is not
misreported as a proxy failure.

1. **Identity:** require `flavor=singbox`, the expected sing-box version and the
   expected platform features from `-V` and `feature-manifest.json`.
2. **Configuration:** render with `-O json`, inspect direction, selected tag,
   types and merge overrides locally, then securely delete or protect the
   secret-bearing output.
3. **TCP:** fetch or echo a controlled payload through the actual `-L`; compare
   its bytes or digest, not just the HTTP status or open port.
4. **UDP:** send distinct datagrams and compare payloads. For a SOCKS UDP
   association, use at least two target ports and one domain destination to
   catch accidental fixed-target or local-resolution behavior.
5. **DNS:** send a fresh A or AAAA query through the configured UDP route and
   parse the response. Repeat against a second controlled resolver before
   diagnosing a transport from one resolver's failure.
6. **Security:** repeat one request with a deliberately wrong password, UUID,
   key, SNI or route. It must fail and must not appear at the final target.
7. **Shutdown/reload:** close the client, stop Gust, verify the listener is
   released, then reload on the fixed port while an existing outbound lease is
   active.

A small local TCP control can be run without an Internet target:

```bash
python3 -m http.server 18080 --bind 127.0.0.1

# In another terminal, start the configuration under test, then force curl to
# use the local SOCKS listener even for a loopback target.
curl --fail --noproxy '' --socks5-hostname 127.0.0.1:1080 \
  http://127.0.0.1:18080/
```

This proves only TCP for that listener and chain. It does not prove UDP, DNS,
Reality or a provider node. A Reality check must pass application bytes after
the native handshake; a TCP connect to its port is not sufficient.

For TUN, REDIRECT and TProxy, use an isolated Linux network namespace or a
dedicated test host. Record the selected firewall backend, rules and counters,
policy routes, original destination and target receipt. A privileged Docker
container can still differ from the host's firewall backend: a nested REDIRECT
failure must be reproduced in a host-created isolated namespace before it is
classified as a product defect. Never weaken production firewall policy merely
to make a test pass.

## Security hardening

- Prefer JSON files with owner-only permissions for long credentials. URI
  credentials may be retained in shell history, process listings or service
  manager metadata.
- Treat `-O json`, `-O yaml`, debug logs and provider configuration as
  secret-bearing output. Do not paste them into issues without manual review.
- Use strict certificate verification and the actual SNI. Do not disable TLS
  verification to hide a name, clock or trust-chain error.
- Keep `route.final` under Gust ownership for native `-L`. Convert an intended
  native exception to an explicit rule; do not expect Gust to overwrite a
  foreign final route.
- Activate only the selected inbound and its exact declared detour closure.
  Unknown, duplicated, cyclic or unrelated activation tags must remain errors.
- Bind administrative or local proxy listeners to loopback unless remote
  access is intentional. Apply native authentication and firewall policy to
  every remotely reachable listener.
- Restrict TUN/TProxy capabilities, routes and firewall changes to the smallest
  required scope. Do not run the entire service privileged when only a helper
  or isolated environment needs network administration.
- Keep the release binary, feature manifest and Cronet library from the same
  archive and verify the published checksum. Do not mix libraries between
  versions or architectures.
- Pin and review the complete native config. Unknown fields and newer schema
  fields are rejected rather than silently ignored by the validated version.
- Verify a negative request never reaches the target. Authentication, prefix
  chain, router and DNS errors are designed to fail closed with no
  system-direct retry.

## Performance and trade-offs

There is no single meaningful "sing-box overhead percentage": Internet RTT,
the chosen protocol, cryptography, TLS/QUIC handshakes, multiplex settings and
packet sizes normally dominate. The integration-specific costs are narrower:

- TCP payloads are not copied through an internal localhost proxy. The backend
  returns the native sing-box connection with lifecycle and runtime-lease
  wrappers; Gust then performs its usual service copy. After connection setup,
  the wrappers do not copy each TCP payload.
- An outbound without a preceding GOST prefix route uses the runtime acquired
  when its Transport starts. A non-Direct embedded outbound after a GOST prefix
  route acquires a route-scoped selected-tag handle for the request. Canonical
  identity lets matching requests share the expensive Box, but that scoped
  acquisition still decodes/canonicalizes the stored config, hashes it and
  performs pool/reference accounting. It is connection work, not packet work.
- The connected UDP adapter currently allocates a temporary
  `len(application buffer) + 512` byte buffer on each `Read`. The extra headroom
  prevents SOCKS-style address headers from truncating the datagram before
  sing-box removes them. This is correct but can increase allocation and GC
  pressure for high packet-rate UDP workloads.
- Protocol costs are the native sing-box implementation's costs: for example,
  AEAD encryption, Reality/TLS, QUIC congestion control, padding and protocol
  multiplexing. Gust adds listener/handler and chain work around that data
  path; it does not remove those costs.
- Native inbound configuration, canonicalization and Box construction happen
  at service start/reload, never once per connection or packet.
- One native `-L` owns one Box. More listeners therefore add native Box
  memory, goroutines, descriptors and startup time linearly. They reuse the
  same GOST chain objects but intentionally do not share listening Boxes.
- `__gust_egress` performs no JSON work and does not open a loopback proxy.
  Its UDP association retains packet boundaries and target addresses; the
  implementation is required not to allocate a large object for every packet.

Practical advantages are one process, no IPC or hidden listener, full selected
node dependency graphs, composition with existing Gust nodes, fail-closed
prefix chaining, remote UDP domain preservation and safe reload leases. The
main costs and limits are a larger singbox binary and runtime memory footprint,
route-scoped outbound handle work, current outbound UDP read allocation, a
pinned schema/feature set, platform-specific features, and one Box per native
listener. Use the standard flavor when none of the embedded protocols is
needed. Put a frequently used embedded node before unnecessary GOST prefix
nodes when both chain orders are semantically valid; never reorder hops merely
for a microbenchmark.

The `singbox-v3.2.11` release has fixed-runner evidence against the same
official sing-box v1.13.16 direct inbound. Five raw samples are included in
`SINGBOX-PERFORMANCE-BASELINE.json`:

| Controlled comparison | Gust / official or measured result | Release gate |
|---|---:|---:|
| TCP throughput median | 100.01% | at least 90% |
| UDP PPS median | 99.78% | at least 90% |
| TCP round-trip p95 / p99 | 100.50% / 100.79% | at most 110% |
| UDP round-trip p95 / p99 | 99.44% / 99.73% | at most 110% |
| `__gust_egress` UDP write | 151.4 ns/op, 96 B, 2 allocs | at most 128 B and 2 allocs |
| Fixed-port reload p95 | 3.79 ms | at most 5 ms |

The one-Box-per-`-L` resource baseline measured 8 goroutines and 4 file
descriptors per live Box. Median startup was 5.04 ms for one Box, 32.70 ms for
10 and 220.15 ms for 50; median live heap delta was about 0.31, 3.04 and
15.20 MB respectively. All 20 fresh-process samples returned goroutine and FD
counts exactly to baseline after Close.

These results prove the direct inbound-to-GOST integration boundary meets its
declared fixed-runner gates. They do not claim that every encrypted protocol,
provider, WAN route or arbitrary shared host has the same ratio. A 2026-08-08
health sample illustrates the distinction: the identical binary reloaded in
2.58–3.23 ms on two shared hosts and 12.81 ms on another. Keep fixed hardware,
Go version, feature tags, GOMAXPROCS and workload for a release comparison;
record environmental outliers instead of averaging them into a product claim.

For reproducible local diagnostics, limit the TCP dial iteration count on
systems with a small ephemeral-port range:

```bash
tags="$(bash ../gust/.github/scripts/singbox-tags.sh)"
go test -tags "$tags" ./backend/singbox -run '^$' \
  -bench 'Benchmark(TCP|RuntimePool)' -benchmem -benchtime=500x

# The differential boundary test must use the same binary and host for both
# official and Gust sub-benchmarks.
GOMAXPROCS=2 go test -tags "$tags" ./backend/singbox -run '^$' \
  -bench '^(BenchmarkInboundTCPThroughput|BenchmarkInboundUDPPPS|BenchmarkEgressUDPPacket|BenchmarkInboundReloadSamePort)$' \
  -benchmem -benchtime=2s -count=5
```

Loopback throughput results are kernel-buffer microbenchmarks. They can detect
regressions and allocations but must not be presented as WAN or encrypted
protocol throughput.

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

For native `-L`, also check:

- `listen_port` is fixed and non-zero, and the address is not already in use.
- Linux TUN/REDIRECT/TProxy privileges, routes and firewall rules exist.
- A full config has no foreign `route.final`.
- `activate_inbounds:=[]` exactly matches the selected inbound's native
  detour closure.
- An application payload reaches the expected final target through the GOST
  chain. A listening port alone proves only startup.

Identical canonical outbound configurations share a runtime. Native inbound
Boxes remain service-owned. Reload starts the new runtime/service before
retiring the old one, and active outbound TCP/UDP leases keep an old pooled
runtime alive until the connection closes.

The validated schema is pinned to sing-box v1.13.16. Fields introduced by a
newer sing-box release are not automatically accepted. Native `-L` startup
is explicit: a full config activates only the selected inbound and its exact
declared dependency closure.

The repository's detailed acceptance evidence is in
`SINGBOX-VALIDATION.md`. Release archives include this manual, the validation
record, `feature-manifest.json`, license notices and exact source revisions.
