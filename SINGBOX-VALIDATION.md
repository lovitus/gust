# Embedded sing-box acceptance record

- Status: PASS — implementation, local matrix, fixed-runner performance,
  six-host fleet validation, required GitHub workflows and the six-platform
  build-only release rehearsal pass.
- Validation date: 2026-08-08
- Gust branch: `singbox-backend`
- Certified release: `singbox-v3.2.12`
- Certified Gust revision: `f7f88dbba0679536e6103b467ade134c0c6882c9`
- gust-x revision: `1b367c99607ec788540518ab8d9abfd5b2307b44`
- Pinned sing-box: `v1.13.16`
- Certification toolchain: `go1.26.5`

This record covers the complete embedded direction pair: native sing-box
inbounds selected by `-L`, native outbounds/endpoints selected by `-F`, and
their composition through an ordinary GOST chain. A port opening or successful
parse is never counted as a protocol data-plane pass.

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
| Naive | PASS | n/a | real HTTP/2 CONNECT and Naive framing |
| WireGuard endpoint | PASS | PASS | selected endpoint path |
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

Raw five-sample results and the release thresholds are checked in at
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
| TCP throughput, 64 KiB echo | 673.82 MB/s | 674.31 MB/s | 100.07% | >=90% | PASS |
| UDP PPS, 1200-byte echo | 10,964 PPS | 11,106 PPS | 101.29% | >=90% | PASS |
| TCP round-trip p95 / p99 | 106.947 / 112.768 us | 106.905 / 114.333 us | 99.96% / 101.39% | <=110% | PASS |
| UDP round-trip p95 / p99 | 104.588 / 115.931 us | 105.283 / 116.773 us | 100.66% / 100.73% | <=110% | PASS |
| Internal egress UDP write | n/a | 150.6 ns/op median | 96 B, 2 allocs | <=128 B, <=2 allocs | PASS |
| Retained runtime handle | 190.89 us decode/construct | 221.5 ns/op | 0.12%, 80 B, 1 alloc | <=1%, <=96 B, <=1 alloc | PASS |
| Route-scope cache hit | n/a | 262.1 ns/op | 80 B, 1 alloc | <=96 B, <=1 alloc | PASS |
| Direct / proxy packet read | n/a | 0 B/0 allocs; 24 B/1 alloc | no payload-scaled allocation | <=0/0; <=64 B/1 | PASS |
| Fixed-port reload pause | n/a | 3.57 ms median, 3.59 ms p95 | n/a | p95 <=5 ms | PASS |

Resource medians for the one-Box-per-`-L` baseline:

| Boxes | Startup | Live heap delta | Goroutine delta | FD delta | Median max RSS |
|---:|---:|---:|---:|---:|---:|
| 1 | 5.17 ms | 312,704 B | 8 | 4 | 20,004,864 B |
| 2 | 8.18 ms | 612,232 B | 16 | 8 | 20,566,016 B |
| 10 | 32.90 ms | 3,008,056 B | 80 | 40 | 23,990,272 B |
| 50 | 207.22 ms | 14,983,336 B | 400 | 200 | 41,844,736 B |

Every one of the 20 fresh-process resource samples returned goroutine and FD
counts exactly to its pre-start baseline after Close. These numbers prove the
direct integration boundary meets the release thresholds on the fixed runner;
they are not WAN or encrypted-protocol performance claims.

The checker also compares this candidate with the previous accepted baseline,
whose source revision, medians and baseline-file SHA-256 are retained in the
new record. TCP throughput changed by -0.65%, UDP operation time by -2.11%,
and the worst p95/p99 round-trip increase was +0.54%. The worst startup,
live-heap or RSS median increase was +0.85%; all remain inside their checked 5%
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
and Darwin amd64/arm64. Linux and Windows packages include the matching Cronet
library for Naive outbound. Darwin records `naive_outbound` and CCM as
unavailable; Naive inbound remains available.

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

- Required Go CI: [run 31264723529](https://github.com/lovitus/gust/actions/runs/31264723529)
- Pinned/latest compatibility: [run 31264438740](https://github.com/lovitus/gust/actions/runs/31264438740)
- Six-platform build-only release rehearsal: [run 31265090521](https://github.com/lovitus/gust/actions/runs/31265090521)

The acceptance matrix is PASS. Publishing still requires a clean pushed source
branch and an `singbox-v*` tag contained in `origin/singbox-backend`; the tag
workflow independently repeats all six package builds before creating the
GitHub Release.
