# Fork Changes (gust)

Based on upstream: go-gost/gost `c7d793c619e3690f091dd98ac9b4085e212db26f` + go-gost/x v0.15.2

## New Features

### Optional embedded sing-box flavor
The permanent `singbox-backend` branch adds separate `standard` and `singbox` build
flavors. Standard builds report `flavor=standard` and are checked to contain no
`github.com/sagernet/sing-box` module; tagged builds report the pinned sing-box
version and feature set. CI smoke-tests both flavors, and the compatibility
workflow compiles against the pinned and latest stable sing-box releases when
the pinned inputs change or a maintainer dispatches it.

The CLI can parse `*+singbox://` node URIs into structured connector
metadata for `-O`, including inline/file/base64 JSON, full-config tag selection,
nested paths, exact JSON values and URL-safe secret handling. Runtime chain
integration supports request-scoped GOST prefix routes, multi-level self-dialing
nodes, native typed option validation, full config dependency graphs and a
content-hash/singleflight/refcount runtime pool. Prefix failures are fail-closed;
every potentially active background network leaf must be proven to reach the
same prefix scope or the scoped configuration is rejected.
Configuration-file metadata supports nested options and json/config references.

`gust-with-singbox` release assets cover Linux amd64/arm64, Windows
amd64/arm64 and Darwin amd64/arm64. Linux and Windows full-feature builds use
the upstream pure-Go Cronet loader and bundle its matching shared library, but
embedded policy rejects Naive outbound on every platform; Naive inbound remains
available. Darwin additionally omits the Naive outbound tag and CCM in its
reproducible `CGO_ENABLED=0` flavor. Each archive contains exact source refs,
GPLv3 text and upstream license notices. Full usage and lifecycle documentation is in
https://github.com/lovitus/gust-x/blob/singbox-backend/docs/singbox.md.

Gust and gust-x maintain this flavor on matching `singbox-backend` branches.
Its `singbox-v*` releases contain only the six embedded assets and never update
the standard `master` release, latest marker, or package-manager channels.

### Porty Stealth Port Forwarding
`gust` registers the `porty` listener/handler/dialer/connector from `gust-x` and
the release workflow builds both `gost` and the standalone `portyd` server.

Porty provides:
1. HTTPS/WebSocket-disguised TCP port forwarding with HMAC authentication and
   chacha20poly1305 AEAD records.
2. smux session caching: one TCP/WS/AEAD session carries unlimited logical
   ltcp/rtcp streams.
3. Multi-port and mixed-role reuse: one cached session can register many
   provider ports and also consume other ports.
4. Optional P2P direct mode: after relay succeeds, peers exchange direct
   candidates over an encrypted peering stream and cache per-port direct
   sessions. Relay remains the fallback path.
5. Optional WebSocket write coalescing for concurrent small AEAD records.
6. Dynamic `bindpath` bridge in standalone `portyd`: an authenticated provider
   can register a temporary WebSocket path, allowing Mihomo/iOS WS clients to
   reach a provider-side local backend without adding Trojan/VLESS/SS decoding
   to `portyd`.

### sings + Mihomo smux
`sings` integrates sing-shadowsocks for TCP, UoT (UDP-over-TCP), and AEAD 2022
ciphers. It now also accepts sing-box/Mihomo common `smux` TCP multiplexing on
the standard `sp.mux.sing-box.arpa:444` destination. This lets Mihomo reuse a
small number of TCP/WSS connections when `sings` is exposed through Porty
dynamic `bindpath`.

Operational notes:
1. Use Mihomo top-level `smux` with `protocol: smux`, `only-tcp: true`, and
   `padding: false`.
2. Keep `gost-plugin` `plugin-opts.mux=false`; that setting is not sing-mux.
3. UDP/UoT still uses separate SS TCP connections and is not controlled by
   `smux.max-connections`.

Full protocol and deployment documentation:
https://github.com/lovitus/gust-x/blob/master/docs/porty.md

### SSH Relay Fallback
When SSH server disables TCP forwarding (AllowTcpForwarding=no), automatically falls back through:
1. direct-tcpip (original SSH, always tried first)
2. Multiplexed relay (single session, unlimited connections)
3. Embedded relay binary (auto-uploaded, hash-cached)
4. exec fallbacks: nc, socat, perl, python, bash

### Escape-Based Password Parsing
Supports backslash escapes and quotes in inline passwords, backward compatible with URL encoding.

### SOCKS5 UDP First-Packet Source Pinning
`udpSourceCheck=first-packet` supports deployments where the TCP control
connection and UDP datagrams arrive from different source IPs, while pinning
the first UDP source IP and port for the rest of the association.

## Modified Files in go-gost/x
- dialer/connector/listener/handler/porty - Porty protocol registration and integration
- internal/util/porty - Porty core protocol, session routing, P2P peering, and tests
- internal/util/ws/coalesce.go - Optional WebSocket write coalescing
- cmd/portyd/main.go - Dynamic bindpath bridge before disguise/proxy handling
- handler/sings - sing-shadowsocks TCP/UoT handler with sing-mux smux support
- internal/util/singmux - Minimal sing-mux wire codec for session/stream framing
- connector/sshd/connector.go - Use DialOrExec instead of direct Dial
- config/cmd/cmd.go - Preprocess userinfo for escape/quote parsing
- internal/util/ssh/session.go - Cleanup relay state on close

## New Files in go-gost/x
- cmd/portyd/main.go - Standalone Porty server
- internal/util/porty/peer - P2P peering payloads and bounded direct probes
- internal/util/porty/session/p2p.go - Provider-side direct listener and knock handling
- internal/util/porty/proto - REGISTER_PATH control frame for dynamic bindpath
- internal/util/ssh/relay.go - Core fallback orchestration
- internal/util/ssh/relay_embed.go - Embedded relay binary management
- internal/util/ssh/mux.go - Mux dialer for multiplexed relay
- internal/util/ssh/muxproto/ - Mux protocol framing
- internal/util/ssh/relaybin/ - Embedded relay binaries (linux/darwin x amd64/arm64)
- cmd/relay/main.go - Relay binary source

## Upstream Merge Notes
When merging future upstream updates, the SSH relay fallback remains isolated to
the sshd connector/session path. Porty, dynamic bindpath, sings, and relay helper
code live in gust-x protocol packages and command binaries, with `gust` mainly
registering and publishing those components.
