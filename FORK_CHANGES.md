# Fork Changes (gust)

Based on upstream: go-gost/gost v3.2.6 + go-gost/x v0.8.1

## New Features

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

Full protocol and deployment documentation:
https://github.com/lovitus/gust-x/blob/main/docs/porty.md

### SSH Relay Fallback
When SSH server disables TCP forwarding (AllowTcpForwarding=no), automatically falls back through:
1. direct-tcpip (original SSH, always tried first)
2. Multiplexed relay (single session, unlimited connections)
3. Embedded relay binary (auto-uploaded, hash-cached)
4. exec fallbacks: nc, socat, perl, python, bash

### Escape-Based Password Parsing
Supports backslash escapes and quotes in inline passwords, backward compatible with URL encoding.

## Modified Files in go-gost/x
- dialer/connector/listener/handler/porty - Porty protocol registration and integration
- internal/util/porty - Porty core protocol, session routing, P2P peering, and tests
- internal/util/ws/coalesce.go - Optional WebSocket write coalescing
- connector/sshd/connector.go - Use DialOrExec instead of direct Dial
- config/cmd/cmd.go - Preprocess userinfo for escape/quote parsing
- internal/util/ssh/session.go - Cleanup relay state on close

## New Files in go-gost/x
- cmd/portyd/main.go - Standalone Porty server
- internal/util/porty/peer - P2P peering payloads and bounded direct probes
- internal/util/porty/session/p2p.go - Provider-side direct listener and knock handling
- internal/util/ssh/relay.go - Core fallback orchestration
- internal/util/ssh/relay_embed.go - Embedded relay binary management
- internal/util/ssh/mux.go - Mux dialer for multiplexed relay
- internal/util/ssh/muxproto/ - Mux protocol framing
- internal/util/ssh/relaybin/ - Embedded relay binaries (linux/darwin x amd64/arm64)
- cmd/relay/main.go - Relay binary source

## Upstream Merge Notes
When merging future upstream updates, only 3 existing files were modified.
The changes are minimal and isolated to the sshd connector path.
