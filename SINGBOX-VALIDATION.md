# Embedded sing-box acceptance record

- Status: PASS
- Validation date: 2026-08-06
- Validated Gust implementation revision: `b522126c9e16fb180c736545b1510352ec91dad8`
- Validated CI workflow revision: `dbf7d291d2c99550d61c25311c9dedd571741525`
- Validated gust-x revision: `db85df20eb30b59fab523d652b194b8f895c2642`
- Pinned sing-box: `v1.13.16`
- Go toolchain: `go1.26.5`

This is the acceptance record for the embedded sing-box delivery, not a claim
that every unrelated upstream or historical repository test has no known
failure. It records the configuration, backend, protocol, packaging and
platform gates used to approve this feature.

## Automated acceptance runs

- Go CI: <https://github.com/lovitus/gust/actions/runs/31072461492>
  - Result: PASS, 12 of 12 jobs.
  - Core tests, standard/singbox flavor separation, embedded backend tests,
    Linux Naive runtime check, five cross-build jobs and six native platform
    smoke jobs all completed successfully.
- Compatibility matrix: <https://github.com/lovitus/gust/actions/runs/31072461526>
  - Result: PASS for pinned and `latest`; both resolved to sing-box v1.13.16 at
    validation time, using Go 1.26.5.
- Release dry-run: <https://github.com/lovitus/gust/actions/runs/31072097266>
  - Result: PASS, six build jobs and six retained platform artifacts.
  - The publish job was intentionally skipped because this was a manual
    build-only validation, so no tag or GitHub Release was created.

## Configuration surface

| Area | Evidence | Result |
|---|---|---|
| CLI `protocol+singbox://` parsing | URI lexer/parser, native typed paths, aliases, URL-safe secrets | PASS |
| Node JSON | inline JSON, file, `json64`, source-size/error handling | PASS |
| Full sing-box config | outbound/endpoint selection, dependency graph, inbound rejection | PASS |
| Mixed config | config + JSON + authority + query precedence | PASS |
| Gust JSON/YAML metadata | nested options and JSON/config references | PASS |
| Exact JSON assignments | bool, number, object, array and null through `path:=JSON` | PASS |
| Config rendering | standard and singbox flavors through `-O json/yaml` | PASS |
| Error redaction | password, private key, token, PSK and raw JSON not echoed | PASS |

The fixed overlay order is: selected full-config node, node JSON overlay, URI
authority/userinfo, ordinary query assignments. Tests also cover arrays,
object-array paths, repeated values, percent-encoded commas, type mismatches,
unknown native fields and transport discriminators.

## Real protocol data paths

Tests create controlled local protocol servers and pass application payloads
through the embedded client; a successful port open alone is not counted.

| Protocol/transport | TCP | UDP | Notes |
|---|---:|---:|---|
| Shadowsocks | PASS | not a protocol-specific gate | `chacha20-ietf-poly1305` data plane |
| SOCKS4/4a/5 | PASS | PASS | authentication, GOST-prefix UDP and SOCKS5 UDP ASSOCIATE covered |
| HTTP proxy | PASS | n/a | authenticated upstream |
| VMess | PASS | not a protocol-specific gate | native TCP and WebSocket |
| VLESS | PASS | not a protocol-specific gate | native TCP, WebSocket TLS, HTTP/2 TLS and gRPC TLS |
| VLESS Reality | PASS | n/a | basic and `xtls-rprx-vision` |
| Trojan | PASS | not a protocol-specific gate | TLS data plane |
| AnyTLS | PASS | not a protocol-specific gate | TLS data plane |
| Hysteria2 | PASS | PASS | native QUIC data plane |
| TUIC | PASS | PASS | native QUIC data plane |
| SSH | PASS | n/a | authenticated TCP proxy data plane |
| Naive | PASS | n/a | Linux full client/server data plane using trusted local certificate |
| WireGuard endpoint | lifecycle PASS | lifecycle PASS | native endpoint validation/start/close |
| Direct | PASS | PASS | embedded baseline and release smoke |

Reality was tested with a valid X25519 key pair, short ID, SNI and controlled
TLS handshake target in both basic and Vision modes. Naive was additionally
rebuilt and exercised on an independent Linux validation client with the
matching Cronet runtime.

## Chain, DNS and failure semantics

| Behavior | Result |
|---|---|
| sing-box node as self-dialing chain transport | PASS |
| Two embedded sing-box nodes in one GOST chain | PASS |
| Ordinary GOST prefix followed by sing-box | PASS |
| TCP and UDP through request-scoped prefix route | PASS |
| Prefix route failure remains fail-closed | PASS |
| Full-config selector uses GOST prefix route | PASS |
| Full-config hosts-based DNS dependency retained | PASS |
| Remote DNS transport automatically follows the GOST prefix | PASS, real UDP query and route-dial assertion |
| Unconnected UDP PacketConn used by SOCKS5 UDP ASSOCIATE | PASS, arbitrary `WriteTo` destination |
| Endpoint prefix injection | PASS |
| Copying a route preserves runtime ownership | PASS |
| Probe uses an explicit target for self-dialing node | PASS |

DNS acceptance covers both a full-config `hosts` dependency and a controlled
networked UDP DNS server. The networked test resolves the proxy hostname,
carries an application payload through the selected outbound, and asserts that
both the DNS UDP connection and proxy TCP connection used the preceding GOST
route. No explicit DNS detour is present in that test.
Operational DNS checks use an actual query and do not infer UDP/DNS support
from TCP port 53 or from one external resolver.

## Lifecycle, concurrency and robustness

| Gate | Result |
|---|---|
| Native schema round-trip matrix | PASS |
| Canonical content identity and runtime sharing | PASS |
| Selected tags share an identical Box | PASS |
| Different prefix paths isolate stateful runtimes | PASS |
| Singleflight concurrent creation | PASS |
| Failed replacement preserves the running runtime | PASS |
| TCP connection lease outlives configuration handle | PASS |
| UDP packet lease outlives configuration handle | PASS |
| Repeated replacement drains the runtime pool | PASS |
| Dial-time context propagation and post-dial detachment | PASS |
| Gust-owned lifecycle/config surface under `-race` | PASS |
| URI lexer fuzzing | PASS, approximately 240,000 executions in 3 seconds |

Runtime-pool handle benchmark, five local runs on Apple M4 with Go 1.26.3:
approximately 13.3–13.6 microseconds/op, 22,486–22,487 bytes/op and 341
allocations/op. Benchmarks are diagnostic baselines, not release pass/fail
thresholds and do not measure an encrypted protocol or network RTT.

## Platform and release assets

| Target | Native build/smoke | Package | Naive runtime |
|---|---:|---:|---:|
| Linux amd64 | PASS | PASS | PASS, bundled `libcronet.so` |
| Linux arm64 | PASS | PASS | PASS, bundled `libcronet.so` |
| Windows amd64 | PASS | PASS | PASS, bundled `libcronet.dll` |
| Windows arm64 | PASS | PASS | PASS, bundled `libcronet.dll` |
| Darwin amd64 | PASS | PASS | intentionally unavailable |
| Darwin arm64 | PASS | PASS | intentionally unavailable |

Every package records the exact Gust/gust-x revisions, Go version, build tags,
runtime files and unavailable features in `feature-manifest.json`. Packaging
checks also cover LF-stable GPL checksum validation, Windows ZIP creation and
extraction, runtime library placement, license/notice inclusion and flavor
inspection with `go version -m`.

## Multi-host operational acceptance

The release-dry-run artifacts were also exercised outside GitHub Actions on
two independent Linux amd64 VPS hosts and one Darwin arm64 host on the
maintainer network. Temporary listeners, credentials, firewall rules and test
files were isolated from the repositories and removed after acceptance.

| Scenario | Result |
|---|---|
| Linux VPS A to VPS B controlled HTTP/TCP payload | PASS |
| Linux VPS B to VPS A controlled HTTP/TCP payload | PASS |
| Darwin CLI `protocol+singbox://` | PASS |
| Darwin inline node JSON | PASS |
| Darwin mixed `-C`, `-L` and `-F` | PASS |
| Darwin two-hop standard GOST plus sing-box chain | PASS |
| SOCKS5 UDP ASSOCIATE through sing-box to controlled UDP echo | PASS on Darwin and Linux |
| Actual UDP DNS query through sing-box | PASS on Darwin and Linux |
| Remote DNS resolves the proxy hostname through the preceding GOST hop | PASS |

The Darwin and Linux binaries came from run `31072097266`; their manifests
identified the Gust implementation and gust-x revisions above, plus Go 1.26.5.
The Linux package included and loaded `libcronet.so`; the Darwin package
correctly declared Naive and CCM unavailable. TCP checks required the expected
application response body. UDP checks required an echoed datagram, and DNS
checks parsed a DNS response with the matching transaction ID and at least one
answer.

## Commands used by maintainers

```bash
# Main command package, including the embedded manual
go test ./cmd/gost

# Embedded backend and integration packages
tags="$(bash .github/scripts/singbox-tags.sh)"
(cd ../gust-x && go test -tags "$tags" \
  ./backend/singbox ./config/cmd/singboxuri ./config/cmd \
  ./chain ./config/parsing/node ./hop)

# Gust-owned race surface; protocol data paths are split in CI as noted below
(cd ../gust-x && go test -race -tags "$tags" \
  ./backend/singbox ./config/cmd/singboxuri ./config/cmd \
  ./chain ./config/parsing/node ./hop)

# Fuzz the URI lexer
(cd ../gust-x && go test ./config/cmd/singboxuri \
  -run '^$' -fuzz '^FuzzLexDoesNotPanic$' -fuzztime 3s)

# Build one release asset
bash .github/scripts/build_singbox_asset.sh \
  --version test --tag test --goos linux --goarch amd64 --out-dir /tmp/gust-assets
```

CI excludes real HTTP/2 and gRPC data-plane cases from the race command, then
runs them normally. sing-box v1.13.16 has upstream race-detector findings in
those transport initializers; Gust lifecycle/config integration remains under
`-race`, while HTTP/2 and gRPC application data paths still have passing native
tests.

## Acceptance boundaries

- The embedded backend implements outbound and endpoint use. `-L` does not map
  arbitrary sing-box inbounds, and complete configs containing inbounds are
  rejected.
- Protocol UDP data-path gates include a remote-only domain association that
  is preserved as a SOCKS address without local resolution. Direct UDP still
  uses local DNS by design because no remote resolver exists.
- GOST's empty-destination UDP dial is retained as an unconnected packet
  connection, so a SOCKS5 listener can relay each UDP ASSOCIATE datagram to
  the destination supplied to `WriteTo`.
- Full-config DNS gates cover both the sing-box `hosts` transport and a real
  networked UDP DNS transport automatically attached to the preceding prefix.
- Darwin release assets use a reproducible `CGO_ENABLED=0` limited feature set
  and intentionally omit Naive and CCM.
- Formal publishing was not part of the dry-run: merge approval, release tag
  and GitHub Release remain explicit maintainer gates.
- GitHub currently emits a Node.js 20 deprecation annotation for upstream
  action versions. Jobs are forced onto Node.js 24 by GitHub and pass; the
  annotation is not a product-test failure.
