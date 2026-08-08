# End-to-end coverage ledger

This is a maintained coverage ledger, not an implementation promise based only
on registered protocol names. A capability moves to **certified** only when a
repeatable test proves its data plane.

## Automated on every `master` change

The Docker suite currently covers these functional areas:

- HTTP proxying, HTTP/2, TLS and mutual TLS.
- Shadowsocks TCP/UDP behavior.
- DNS and resolver behavior over TCP/UDP.
- Forwarding, ingress, routing, chained routing and node selection.
- Authentication, admission, bypass, hosts and probe policies.
- HTTP cache, file service, sniffing, uTLS and PHT.
- Negative paths and stable failure codes.

The CI unit/configuration job and Docker E2E job are independent and each has a
hard time limit. Known upstream-inherited suite exclusions apply only to the
unit/configuration job; Docker E2E remains separately visible.

## Protocol expansion queue

Each row requires TCP and UDP where the protocol supports both, authentication
where applicable, a negative case, bounded shutdown and leak-free cleanup.

| Class | Examples | Required evidence | Status |
|---|---|---|---|
| SOCKS4/4a | TCP proxy | Returned payload | certified |
| SOCKS5 | TCP, authentication, UDP association | Returned payload, auth rejection and repeated datagrams | certified |
| Relay | TCP and authentication | Returned payload and auth rejection | certified; UDP queued |
| Stream transports | WebSocket/WSS, gRPC, h2/h2c | Payload through each negotiated transport | certified |
| UDP-backed stream transports | QUIC, KCP | TCP payload over each UDP transport | certified; native datagram boundary queued |
| Multiplexing | mws/mwss/mtls/mtcp | Six concurrent streams per transport | certified |
| Obfuscation | ohttp/otls | Payload | certified; invalid-peer failure queued |
| SSH | ssh/sshd | Generated test-only host key and payload | queued |
| HTTP/3 family | HTTP/3, WebTransport | QUIC data plane and shutdown | queued |
| Privileged networking | TUN/TAP, redirect, ICMP, MASQUE | Isolated privileged runner evidence | runner required |

## External compatibility

Compatibility tests must pin or record the peer implementation version and
must not depend on public DNS or a third-party echo service. Mihomo UoT v2 DNS
and arbitrary UDP are certified by the independent `Mihomo UoT v2
compatibility` CI job. It uses locally controlled responders and verifies the
downloaded peer binary checksum. Evidence published to the repository must use
sanitized role names and example endpoints.

## Completion rules

1. Configuration parsing and a listening socket do not count as protocol
   support.
2. Tests must be self-contained or clearly identify the required runner
   capability.
3. Failure paths must terminate within the job timeout and print container
   state for diagnosis.
4. New documentation and fixtures must pass the privacy checker.
5. Benchmarks are kept separate from correctness tests and run on a fixed
   environment before a performance claim is made.
