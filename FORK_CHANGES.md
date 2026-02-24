# Fork Changes (gust)

Based on upstream: go-gost/gost v3.2.6 + go-gost/x v0.8.1

## New Features

### SSH Relay Fallback
When SSH server disables TCP forwarding (AllowTcpForwarding=no), automatically falls back through:
1. direct-tcpip (original SSH, always tried first)
2. Multiplexed relay (single session, unlimited connections)
3. Embedded relay binary (auto-uploaded, hash-cached)
4. exec fallbacks: nc, socat, perl, python, bash

### Escape-Based Password Parsing
Supports backslash escapes and quotes in inline passwords, backward compatible with URL encoding.

## Modified Files in go-gost/x
- connector/sshd/connector.go - Use DialOrExec instead of direct Dial
- config/cmd/cmd.go - Preprocess userinfo for escape/quote parsing
- internal/util/ssh/session.go - Cleanup relay state on close

## New Files in go-gost/x
- internal/util/ssh/relay.go - Core fallback orchestration
- internal/util/ssh/relay_embed.go - Embedded relay binary management
- internal/util/ssh/mux.go - Mux dialer for multiplexed relay
- internal/util/ssh/muxproto/ - Mux protocol framing
- internal/util/ssh/relaybin/ - Embedded relay binaries (linux/darwin x amd64/arm64)
- cmd/relay/main.go - Relay binary source

## Upstream Merge Notes
When merging future upstream updates, only 3 existing files were modified.
The changes are minimal and isolated to the sshd connector path.
