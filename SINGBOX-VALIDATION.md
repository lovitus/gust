# Embedded sing-box acceptance record

- Status: PASS — implementation, local matrix, fixed-runner performance,
  six-host fleet validation, required GitHub workflows and the six-platform
  build-only release rehearsal pass.
- Validation date: 2026-08-08
- Gust branch: `singbox-backend`
- Certified release: `singbox-v3.2.12`
- Certified Gust revision: `f7f88dbba0679536e6103b467ade134c0c6882c9`
- gust-x revision: `1b367c99607ec788540518ab8d9abfd5b2307b44`
- Certified release pin: `v1.13.16`
- Current branch pin: `v1.13.17-0.20260811025254-2926c94dd073`
  (maintainer-fork pseudo-version based on upstream `v1.13.16`)
- Certification toolchain: `go1.26.5`

This record covers the complete embedded direction pair: native sing-box
inbounds selected by `-L`, native outbounds/endpoints selected by `-F`, and
their composition through an ordinary GOST chain. A port opening or successful
parse is never counted as a protocol data-plane pass.

The protocol and policy matrices below describe the current branch unless a
paragraph explicitly labels historical `singbox-v3.2.12` evidence. That release
allowed a Linux Naive outbound; the current embedded resource contract rejects
Naive outbound on every platform while retaining Naive inbound.

## How to read this record

Evidence is deliberately separated into levels. A higher level includes the
lower-level checks but cannot be inferred from them.

| Level | What it proves | What it does not prove |
|---|---|---|
| Build | The selected platform and feature tags compile | The protocol starts or carries traffic |
| Decode | Native schema and adapter controls accept the configuration | Credentials, routes or peers work |
| Start | The native object initializes and a declared resource is acquired | A handshake or application payload succeeds |
| Data plane | A controlled native client/server exchanges and verifies application bytes | Every public provider or WAN path works |
| Fleet | The packaged artifact repeats the data-plane check on another OS/kernel | Fixed-runner performance equivalence |
| Performance gate | Five raw samples on the fixed runner meet the declared ratios and budgets | Encrypted-protocol or WAN speed |

`PASS` in a protocol table means data-plane evidence, not merely Build, Decode
or Start. TCP checks compare a complete payload, UDP checks compare datagrams
and destination identity, and DNS checks parse a real query/response. Protocols
with authentication, encryption or TLS use a real native handshake. Negative
checks must fail before system-direct fallback can occur.

## Architecture gates

| Gate | Result |
|---|---|
| Native inbound is an independent top-level `service.Service` | PASS |
| Native outbound remains a self-dialing GOST Transport | PASS |
| No child process, IPC, hidden listener or loopback proxy bridge | PASS |
| `__gust_egress` connects inbound traffic directly to a GOST Router | PASS |
| TCP destination and cancellation preserved | PASS |
| One UDP association retains multiple IP/domain destinations | PASS |
| Chain or route-scope failure remains fail-closed | PASS |
| Inbound Box is service-owned; outbound Box remains pooled | PASS |
| One `-L` per Box correctness baseline | PASS |
| Multiple `-L` services reuse the same named GOST chain configuration | PASS |
| Standard binary does not link or start sing-box | PASS |

Three native `-L` services followed by one four-node `-F` sequence were
started as three independent Boxes, and all three passed application payloads
through the generated `chain-0`. Automatic inbound grouping is not required
for correctness or release.

## Configuration surface

The same parser core, source loader, merge order and typed path engine is used
in both directions. Direction-specific native schemas remain separate.

| Surface | Inbound `-L` | Outbound/endpoint `-F` |
|---|---:|---:|
| `protocol+singbox://` CLI | PASS | PASS |
| Authority and direction-aware userinfo | PASS | PASS |
| Nested `path=value` | PASS | PASS |
| Exact `path:=JSON` | PASS | PASS |
| Inline readable `json=` | PASS | PASS |
| JSON file through `json=` | PASS | PASS |
| Inline/file full config | PASS | PASS |
| Selected tag | `inbound=` PASS | `outbound=` / `endpoint=` PASS |
| Gust JSON/YAML metadata | PASS | PASS |
| Mixed config + CLI override | PASS | PASS |
| Standard flavor `-O json/yaml` | PASS | PASS |
| Native canonical validation | PASS | PASS |
| Safe `-singboxcheck` without startup or value output | PASS | PASS |
| Secret-redacted errors | PASS | PASS |

Readable inline JSON and files are the documented inputs. Base64 controls are
compatibility-only and are not recommended. No heredoc-specific feature was
added.

Merge priority is fixed:

```text
selected full-config object
  < json object overlay
  < URI authority and userinfo
  < ordinary query assignments
```

Adapter controls (`inbound`, `outbound`, `endpoint`,
`activate_inbounds`) are removed before native assignment and decoding.
`activate_inbounds` requires exact JSON array syntax and must equal the
selected inbound's complete native detour closure.

## Full-config safety

| Behavior | Result |
|---|---|
| Only the selected inbound starts by default | PASS |
| Exact inbound detour activation set | PASS |
| Missing, unrelated, cyclic or duplicate activation tag rejected | PASS |
| Unselected listener does not bind | PASS |
| Foreign `route.final` rejected instead of overwritten | PASS |
| Explicit native route rules retain priority | PASS |
| Unmatched/default traffic enters GOST final egress | PASS |
| Fixed non-zero socket port required | PASS |
| TUN reports a stable synthetic address | PASS |
| Removed ShadowsocksR stub rejected | PASS |
| Unknown native resource behavior rejected fail-closed | PASS |
| Native services/APIs/DNS/NTP/rule-set resources without extractor rejected | PASS |
| Unsupported generic GOST service wrappers rejected explicitly | PASS |
| Naive outbound IPC/loopback bridge rejected on every platform | PASS |
| Tailscale endpoint, DHCP DNS and resolved service rejected | PASS |
| WireGuard endpoint accepted only without a preceding GOST prefix | PASS |

The checked-in manifest is generated from the pinned sing-box registration
source with build constraints. CI regenerates it and fails on drift. The full
feature registry contains 18 inbound registrations: 17 available types and the
removed ShadowsocksR compatibility stub.

## Native inbound protocol matrix

All available types have a real controlled client/server data path into the
GOST Router.

| Inbound | TCP | UDP | Additional evidence |
|---|---:|---:|---|
| Shadowsocks | PASS | PASS | AEAD authentication |
| SOCKS | PASS | PASS | authenticated SOCKS5 |
| HTTP | PASS | n/a | authenticated CONNECT |
| Mixed | PASS | PASS | SOCKS path |
| VMess | PASS | PASS | native client |
| VLESS | PASS | PASS | native client |
| Trojan | PASS | PASS | TLS |
| AnyTLS | PASS | PASS | TLS and UoT packet path |
| Hysteria | PASS | PASS | native QUIC |
| Hysteria2 | PASS | PASS | native QUIC |
| TUIC | PASS | PASS | native QUIC |
| Naive | PASS | n/a | real HTTP/2 CONNECT and Naive framing |
| Direct | PASS | PASS | override destination |
| VLESS Reality | PASS | n/a | basic and `xtls-rprx-vision` |
| ShadowTLS | PASS | dependency path | v3 entry to explicitly activated Shadowsocks detour |
| TUN | PASS | PASS | real isolated Linux TUN |
| REDIRECT | PASS | n/a | real iptables REDIRECT and original destination |
| TProxy | PASS | PASS | real policy route, TPROXY target and original destination |
| ShadowsocksR | unavailable | unavailable | removed native stub; rejected |

The ordinary outbound matrix uses the same evidence rule:

| Outbound or endpoint | TCP | UDP | Handshake/dependency evidence |
|---|---:|---:|---|
| Shadowsocks | PASS | PASS | AEAD authentication |
| SOCKS4 / SOCKS4a | PASS | n/a | IPv4 and domain target forms |
| SOCKS5 | PASS | PASS | authentication and UDP association |
| HTTP | PASS | n/a | authenticated CONNECT |
| VMess | PASS | PASS | native peer and transport options |
| VLESS | PASS | PASS | native peer and domain destination |
| VLESS Reality / Vision | PASS | n/a in certified profiles | Reality key/SNI; basic and `xtls-rprx-vision` |
| Trojan | PASS | PASS | TLS and authentication |
| AnyTLS | PASS | PASS | TLS and UoT packet path |
| Hysteria2 | PASS | PASS | native QUIC peer |
| TUIC | PASS | PASS | native QUIC peer |
| SSH | PASS | n/a | authenticated SSH channel |
| Naive | POLICY REJECTED | POLICY REJECTED | fixed fail-closed IPC/loopback error |
| WireGuard endpoint | PASS | PASS | selected endpoint path, unscoped only |
| Direct | PASS | PASS | control and prefix-route baseline |

Outbound chaining additionally retains domain destination, remote DNS,
selector, native detour and endpoint coverage. A protocol marked `n/a` does not
gain UDP support from the adapter.

## Chain, DNS and failure matrix

| Scenario | Result |
|---|---|
| Native `-L ->` system direct | PASS |
| Native `-L ->` ordinary GOST `-F` | PASS |
| Native `-L ->` sing-box `-F` | PASS |
| Native `-L ->` ordinary `-F ->` sing-box `-F` registry isolation | PASS |
| Ordinary and embedded nodes mixed in one chain | PASS |
| Multiple non-contiguous embedded nodes | PASS |
| Two embedded outbound nodes | PASS |
| TCP application payload | PASS |
| UDP application datagram | PASS |
| One UDP association with multiple targets | PASS |
| Domain destination retained without premature local resolution | PASS |
| Actual DNS query through the configured UDP path | PASS |
| Remote DNS automatically follows a preceding GOST route | PASS |
| Explicit native DNS detour remains native | PASS |
| Authentication/TLS/route/DNS failure does not bypass the chain | PASS |

## Security and abuse-resistance matrix

| Threat or operator error | Enforced behavior | Evidence |
|---|---|---|
| Wrong password, UUID, key, SNI or ALPN | Native handshake fails; no direct retry | Negative data-plane tests |
| Prefix proxy, selected route or DNS transport fails | Request fails closed | Prefix/router/DNS failure tests |
| User full config owns `route.final` | Startup rejects the conflict | Config transform tests |
| Unselected inbound or unrelated detour is present | It is filtered and never binds | Activation and listener tests |
| Detour closure is missing, cyclic, duplicated or excessive | Startup rejects it | Activation-set tests |
| Unknown native field or wrong direction field | Canonical decode rejects the exact path | Native validation tests |
| Error contains password, token, private key or complete JSON | Sensitive value is redacted | Redaction tests |
| Removed or unknown registry type | No fallback to another registry or direct | Registry-scope tests |
| Native type may own an unknown socket/system resource | Activation manifest rejects it | Generated-manifest tests |
| Port is zero, occupied or replacement cannot bind | Failure is reported; the old service is retained when possible | Lifecycle tests |
| Caller cancels or closes an active route | Dial/copy stops and leases are released | Cancellation and forced-close tests |
| Static preflight output exposes a config value | Output is structure-only and startup is not attempted | CLI example/preflight tests |
| Published docs contain a real IP/domain/likely credential | CI rejects the documentation | Documentation privacy checker |

Configuration output from `-O`, debug logs and copied provider JSON can still
contain operational secrets. Store source files with least-privilege filesystem
permissions, avoid putting credentials in shell history, and redact output
before sharing it. The documentation checker is a publication guard, not a
general secret manager.

## Lifecycle, reload and robustness

| Gate | Result |
|---|---|
| Synchronous startup reports bind failure | PASS |
| Replacement starts before old service retirement | PASS |
| Failed replacement preserves running service | PASS |
| Active outbound TCP/UDP lease outlives old config handle | PASS |
| Inbound close stops listener and releases Router | PASS |
| Multiple service lifecycle isolation | PASS |
| Fixed port conflict | PASS |
| Cancel, deadline and forced close | PASS |
| Runtime-pool singleflight/reference ownership | PASS |
| Retained selected-tag handle avoids per-connection decode/Box construction | PASS |
| Route-scope cache concurrent acquire/close under race detector | PASS |
| Direct/proxy UDP read allocation budgets | PASS |
| Gust-owned lifecycle/config surface under race detector | PASS |
| URI lexer fuzzing | PASS |
| CI test, cross-build and native-smoke matrices schedule independently | PASS |
| Privileged runner resources and native smoke children have bounded cleanup | PASS |

Reality/Vision data paths run normally but are excluded from Go's checkptr race
mode because pinned upstream `sing-vmess` Vision code faults under that
instrumentation. The rest of Gust's lifecycle and configuration integration
remains under `-race`; Reality is still exercised as a real data path on all
ordinary platform runs.

## Fixed-runner performance

Raw five-sample benchmark/latency results, two complete five-sample resource
rounds, and the release thresholds are checked in at
`SINGBOX-PERFORMANCE-BASELINE.json`. The runner is identified generically;
private infrastructure addresses are intentionally excluded. The values below
qualify the current fork-pin candidate; they do not replace the certified
`singbox-v3.2.12` release record at the top of this document until the candidate
branch completes its required workflows and is promoted.

Environment: Linux amd64, Intel Xeon E5-2620 v4, GOMAXPROCS=2, Go 1.26.5,
sing-box `v1.13.17-0.20260811025254-2926c94dd073`, full feature tags, CGO
disabled, loopback MTU 65536. The exact gust-x source revision and test-binary
SHA-256 are recorded in the baseline file.

| Comparison | Official median | Gust median | Ratio | Gate | Result |
|---|---:|---:|---:|---:|---:|
| TCP throughput, 64 KiB echo | 677.39 MB/s | 674.93 MB/s | 99.64% | >=90% | PASS |
| UDP PPS, 1200-byte echo | 10,905 PPS | 10,807 PPS | 99.11% | >=90% | PASS |
| TCP round-trip p95 / p99 | 106.871 / 115.781 us | 106.579 / 115.530 us | 99.73% / 99.78% | <=110% | PASS |
| UDP round-trip p95 / p99 | 106.515 / 121.742 us | 106.972 / 123.102 us | 100.43% / 101.12% | <=110% | PASS |
| Internal egress UDP write | n/a | 150.3 ns/op median | 96 B, 2 allocs | <=128 B, <=2 allocs | PASS |
| Retained runtime handle | 191.50 us decode/construct | 220.1 ns/op | 0.11%, 80 B, 1 alloc | <=1%, <=96 B, <=1 alloc | PASS |
| Route-scope cache hit | n/a | 261.3 ns/op | 80 B, 1 alloc | <=96 B, <=1 alloc | PASS |
| Direct / proxy packet read | n/a | 0 B/0 allocs; 24 B/1 alloc | no payload-scaled allocation | <=0/0; <=64 B/1 | PASS |
| Fixed-port reload pause | n/a | 3.59 ms median, 3.59 ms p95 | n/a | p95 <=5 ms | PASS |

Resource medians for the one-Box-per-`-L` baseline:

| Boxes | Startup | Live heap delta | Goroutine delta | FD delta | Median max RSS |
|---:|---:|---:|---:|---:|---:|
| 1 | 5.30 ms | 312,496 B | 8 | 4 | 19,955,712 B |
| 2 | 8.21 ms | 612,176 B | 16 | 8 | 20,488,192 B |
| 10 | 32.85 ms | 3,025,432 B | 80 | 40 | 24,207,360 B |
| 50 | 207.49 ms | 14,982,904 B | 400 | 200 | 41,553,920 B |

Every one of the 40 fresh-process resource samples returned goroutine and FD
counts exactly to its pre-start baseline after Close. These numbers prove the
direct integration boundary meets the release thresholds on the fixed runner;
they are not WAN or encrypted-protocol performance claims.

The checker also compares this candidate with the previous accepted baseline,
whose source revision, medians and baseline-file SHA-256 are retained in the
new record. TCP throughput changed by +0.09%, UDP operation time by +2.76%,
and the worst p95/p99 round-trip increase was +5.42%. The worst startup,
live-heap or RSS median increase was +2.52%; all remain inside their checked 5%
or 10% limits.

### Supplemental non-gating performance health check

On 2026-08-08 the certified source was rebuilt with Go 1.26.3 and sampled on
the internal Linux fleet. This is intentionally not substituted for the fixed
Go 1.26.5 release gate:

| Check | Supplemental result | Interpretation |
|---|---:|---|
| TCP throughput median, Gust / official | 95.42% | Above the 90% release floor |
| UDP PPS median, Gust / official | 107.34% | Above the 90% release floor |
| TCP round-trip p95 / p99 | 100.55% / 108.99% | Within the 110% ratio ceiling |
| UDP round-trip p95 / p99 | 101.00% / 96.20% | Within the 110% ratio ceiling |
| Internal egress UDP write | 96 B, 2 allocs/op | Allocation budget retained |
| Reload median on three shared hosts | 12.81 / 3.23 / 2.58 ms | One noisy host exceeds the fixed 5 ms gate |

The identical reload binary met the gate on two hosts and missed it on one.
That variance is recorded rather than averaged away: a shared validation host
is useful for drift detection but is not a controlled performance runner. A
single resource sample at 1/2/10/50 Boxes again returned goroutines and FDs
exactly to baseline after Close; live deltas remained 8 goroutines and 4 FDs
per Box.

## Platform and validation fleet

The same Linux amd64 test binary was hash-verified across five Linux machines;
the same Darwin arm64 binary was used on the macOS machine.

| Environment | Native protocol matrix | Privileged matrix |
|---|---:|---:|
| Three internal Linux hosts | PASS on all | TUN/REDIRECT/TProxy PASS on all |
| Overseas Linux VPS A | PASS | TUN/REDIRECT/TProxy PASS |
| Overseas Linux VPS B | PASS | TUN/REDIRECT PASS; nested-netns TProxy unavailable |
| Internal Darwin arm64 | PASS | Linux-only, n/a |

The final CLI fleet pass used self-started services only. All six machines ran
three independent native `-L` services sharing one embedded direct chain and
passed TCP, UDP and an actual DNS query. A three-machine internal chain and a
two-VPS overseas chain additionally ran native SOCKS inbound -> ordinary GOST
SOCKS prefix -> embedded Shadowsocks outbound -> native Shadowsocks inbound;
TCP, UDP and DNS passed end to end. This run exposed and then regression-tested
the inbound/outbound registry-scope isolation boundary.

On VPS B, the kernel accepted the TPROXY target, incremented firewall counters
and loaded the required modules, but did not deliver marked traffic inside the
nested isolated namespace. That host is recorded as an environment capability
exception, not a product pass. Four independent Linux kernels completed the
same real TProxy TCP/UDP data path.

CI cross-builds and native smoke cover Linux amd64/arm64, Windows amd64/arm64
and Darwin amd64/arm64. Linux and Windows full-feature packages still include
the matching Cronet library required by the compiled native graph, but every
platform records `naive_outbound` as unavailable because embedded activation
policy rejects its IPC/loopback bridge. Darwin additionally records CCM as
unavailable and omits Cronet. Naive inbound remains available everywhere.

### Supplemental fleet run on 2026-08-08

One hash-identical Linux amd64 full-tag test binary was copied to the three
internal Linux hosts (binary SHA-256
`84a9908ca39cd3117c2b1baf439d3f75c91f1dbc8cf4ee195b6780bbeed724fb`;
companion library SHA-256
`dc7293a929dffa695aae1a89555e7366158fa0a3f40bbe3012d445bc05c99672`).
All three completed the full embedded protocol,
configuration, failure and security suite. All three then completed the real
host-isolated TUN, REDIRECT and TProxy TCP/UDP matrix. The source-layout-only
pin test was run in the source checkout; it was explicitly skipped for the
artifact-only remote runs because its compiled absolute `go.mod` path does not
exist there.

An unprivileged Docker repetition passed. A privileged nested Docker run with
a different firewall backend passed TUN and TProxy but its REDIRECT connection
was refused; the same binary passed REDIRECT in a host-created isolated network
namespace. This is retained as an environment prerequisite finding: Docker
privilege alone does not prove that its firewall backend matches the host. A
container result is not promoted to PASS until the target, rules, counters and
original-destination path are all observed.

The Darwin arm64 release candidate archive was checksum-verified on the
internal macOS host (SHA-256
`3de3c3aeb985ad268c69e97fa45d7781322cdf4c18e58433be6dd2a034f93997`).
Its offline manual and feature identity loaded, then a SOCKS native
inbound delivered a byte-compared TCP file through an embedded Direct
outbound. A Direct native UDP inbound separately returned an exact datagram and
carried a parsed DNS query/response. Both URI fuzz targets also completed more
than 1.9 million combined executions without a crash on local Darwin arm64.

## Reproduction commands

```bash
# Standard regression
(cd ../gust-x && go test ./...)

# Full embedded feature set
tags="$(bash .github/scripts/singbox-tags.sh)"
(cd ../gust-x && go test -p 1 -tags "$tags" ./backend/singbox ./config/...)

# Privileged Linux matrix (run from a full-tag test binary as root)
GUST_SINGBOX_PRIVILEGED_TESTS=1 ./singbox.test \
  -test.v -test.run '^TestEmbeddedLinuxPrivilegedInboundMatrix$'

# Official differential and internal egress performance
GOMAXPROCS=2 ./singbox.test -test.run '^$' \
  -test.bench '^(BenchmarkInboundTCPThroughput|BenchmarkInboundUDPPPS|BenchmarkEgressUDPPacket|BenchmarkInboundReloadSamePort)$' \
  -test.benchmem -test.benchtime=2s -test.count=5

# Fresh process for each resource count
GOMAXPROCS=2 GUST_SINGBOX_RESOURCE_BOXES=50 ./singbox.test \
  -test.v -test.run '^TestInboundBoxResourceProfile$'

# Validate checked-in raw evidence and release gates
python3 .github/scripts/check_singbox_performance.py

# Parser robustness and publication privacy
(cd ../gust-x && go test ./config/cmd/singboxuri -run '^$' \
  -fuzz '^FuzzLexDoesNotPanic$' -fuzztime=10s)
python3 .github/scripts/check_doc_privacy.py \
  SINGBOX-ARCHITECTURE.md SINGBOX-VALIDATION.md \
  cmd/gost/SINGBOX_MANUAL.md examples/singbox/README.md
```

## Certification evidence

- Functional code baseline (19/19): [run 31646252480](https://github.com/lovitus/gust/actions/runs/31646252480)
- Reporting/policy baseline (19/19): [run 31661558459](https://github.com/lovitus/gust/actions/runs/31661558459)
- Pinned/latest compatibility on that reporting/policy baseline:
  [run 31661558468](https://github.com/lovitus/gust/actions/runs/31661558468)
- Matching gust-x reporting/policy branch check:
  [run 31661536717](https://github.com/lovitus/gust-x/actions/runs/31661536717)
- Earlier six-platform build-only release rehearsal: [run 31265090521](https://github.com/lovitus/gust/actions/runs/31265090521)

These exact runs certify the named immutable subjects. The moving branch HEAD
must still pass its own required workflows before a release; this document does
not recursively embed its own future commit SHA. The local recovery package
records exact permanent branch refs, fixed-runner performance evidence and
later documentation-only branch CI without embedding maintainer-private runner
metadata in this repository.

The acceptance matrix is PASS. Publishing still requires a clean pushed source
branch and an `singbox-v*` tag contained in `origin/singbox-backend`; the tag
workflow independently repeats all six package builds before creating the
GitHub Release.

## Classic VMess UDP association regression (2026-08-13)

A maintainer-provided VMess WebSocket/TLS service was checked from two overseas
Linux VPS hosts against one pure-IPv4 TCP/UDP echo target. The service address,
credentials and raw host logs remain in the private validation archive.

The control matrix isolated the adapter boundary:

| Client path | TCP | Default VMess UDP |
|---|---:|---:|
| official sing-box 1.13.18 | 10/10 per VPS | 10/10 per VPS |
| Mihomo 1.19.29 | 10/10 per VPS | 10/10 per VPS |
| Gust before the fix | 10/10 per VPS | 0/10 per VPS |

GOST opens SOCKS UDP as an unconnected association and supplies the target on
each `WriteTo`; classic VMess instead fixes one destination in its session
handshake. The old adapter opened the native packet connection with an empty
destination, logged `:0`, and timed out. XUDP happened to carry each packet's
address and therefore masked the adapter error; it is not required for normal
VMess UDP and is not enabled implicitly by the fix.

The final gust-x subject is
`318b6c61cb2f85c8c84e439371ead4b063ae75db`. It lazily creates and reuses one
bounded classic VMess packet session per actual association target. On both VPS
hosts the exact final binary passed TCP 10/10, 100 newly-created default UDP
associations out of 100, and 100 packets alternating between two targets in one
association. Each log contained 102 real target entries and zero `:0` entries.
The tested binary SHA-256 was
`646f761466da4c426951ccce318516b6f5d76fb33ec987cfcafa361441196914`.

## Sanitized protocol-template acceptance (2026-08-13)

`examples/singbox/protocol-templates` contains paired server/client objects for
VLESS Reality Vision, Hysteria2, TUIC, ShadowTLS v3, Shadowsocks 2022, Trojan,
VLESS gRPC Reality, AnyTLS and VMess WebSocket TLS. The source shapes were
compared with the maintainer's private sing-box server configuration; all
endpoints, identities, credentials, keys, pins and service-specific paths were
replaced before the files entered Git.

The final `-F` objects were replayed against the corresponding live services
from two independent VPS hosts. All nine protocol shapes passed on both hosts;
one first-attempt Trojan check was transient and the bounded replay passed 3/3.
The two unavailable Cloudflare Tunnel nodes are not presented as working
templates, and Naive outbound remains excluded by the embedded IPC/loopback
activation policy.

CI parses every template and runs `-singboxcheck` over all paired object files,
the two-object ShadowTLS detour closure, representative CLI-only forms and the
two complete Gust JSON-only VMess configurations. Release packaging recursively
includes the catalog, and native asset smoke verifies its presence. Static
checks never open a listener, certificate or system resource.

## Mixed 11-hop chain and SOCKS BIND acceptance (2026-08-14)

Two private Linux VPS hosts exercised the release pair without reusing any
production credential. VPS A ran one Gust process with nine simultaneous
native sing-box inbounds (the same protocols listed above), one authenticated
SOCKS5 inbound with BIND enabled and one SINGS inbound with UoT enabled. VPS B
constructed one ordered `-F` chain containing all nine native outbounds,
SOCKS5 and SINGS. Temporary ports, identities, Reality keys, passwords and a
one-day test certificate were generated only for this run and removed from
both hosts afterward.

The 11-hop forward chain passed:

| Data plane | Result |
|---|---:|
| SOCKS TCP, concurrent short requests | 200/200 |
| TCP forward, concurrent short requests | 200/200 |
| SOCKS TCP, 8 MiB response | 10/10 |
| TCP forward, 8 MiB response | 10/10 |
| UDP echo, 1200-byte payload | 500/500 |
| DNS A query over UDP | 500/500 |

The chain was then reordered so authenticated SOCKS5 was the final `-F` hop.
Its BIND command created the negotiated remote listener; RTCP passed 50/50
concurrent HTTP round trips. An LTCP consumer traversed the complete chain to
that remote listener and returned over the RTCP provider path, passing 50/50
short requests and 10/10 responses of 8 MiB. All participating processes
remained alive and their final logs contained no fatal, panic or error record.

This runtime check also found that the ShadowTLS helper Shadowsocks inbound
must have an explicit loopback `listen_port` when Gust activates both objects
in the detour closure. The checked-in template now supplies that private helper
port, and CI asserts that it remains loopback-only and non-zero. The exact
Linux binary SHA-256 used by this run was
`1f04e683713ec58f0f580ffb9bb9e8f8dde1e7736a9ee7945010adbdc092dcd9`;
raw logs and private host metadata remain outside the repository.

## Native IPv6 mixed-chain acceptance (2026-08-14)

The mixed-chain matrix was repeated between two private Linux hosts that each
had a native global IPv6 address and IPv6 default route. Both directions first
passed 10/10 ICMPv6 probes. Every native server object listened on `::` (the
ShadowTLS helper used `::1`), all 11 client hop addresses used IPv6 literals,
and the HTTP, UDP echo and DNS targets also used the second host's global IPv6
literal. No tunnel or address translation was used in the accepted run.

Socket and log audits prevented an IPv4 fallback: the service ports exposed 10
TCPv6 and three UDPv6 listeners with zero corresponding IPv4 listeners; the
SOCKS BIND listener and LTCP consumer were TCPv6-only. The final client log
contained 4,051 native outbound connection records with bracketed IPv6 targets
and zero non-bracketed target records.

The complete 11-hop chain passed the same data-plane gates over IPv6:

| Data plane | Result |
|---|---:|
| SOCKS TCP, concurrent short requests | 200/200 |
| TCP forward, concurrent short requests | 200/200 |
| SOCKS TCP, 8 MiB response | 10/10 |
| TCP forward, 8 MiB response | 10/10 |
| UDP echo, 1200-byte payload | 500/500 |
| DNS request over IPv6 UDP transport | 500/500 |
| SOCKS-final RTCP | 50/50 |
| SOCKS-final LTCP | 50/50 |
| LTCP/RTCP return path, 8 MiB response | 10/10 |

An incremental topology check also documented the packet-carrier boundary.
TCP passed at every depth from one through 11 hops. UDP passed with one through
three hops; using ShadowTLS/Shadowsocks as the outermost layer at depths four
and five did not provide the required packet carrier. Trojan restored packet
transport at depth six, depths six through nine passed, and SOCKS as the
outermost tenth hop again lacked it. Adding SINGS/UoT as hop 11 restored UDP,
and the complete target topology passed 500/500. This is an ordering property,
not an IPv6-specific protocol failure.

The first private host proposed for this check was not used because inspection
showed only a link-local address and no IPv6 default route. It was replaced by
another maintainer-provided host with native global IPv6. The replacement
host's default-drop IPv6 firewall was opened only for the peer's source IPv6
and the temporary test port range; those rules, all processes, configurations,
credentials and listeners were removed after evidence capture. Raw host data
remains in the maintainer-private archive.
