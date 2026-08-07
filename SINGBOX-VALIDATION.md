# Embedded sing-box acceptance record

- Status: PASS — implementation, local matrix, fixed-runner performance,
  six-host fleet validation, required GitHub workflows and the six-platform
  build-only release rehearsal pass.
- Validation date: 2026-08-08
- Gust branch: `singbox-backend`
- Certified Gust revision: `424b092bb8a31277e4ea94f4ee8970e5dbd2bd84`
- gust-x revision: `713e879c97dba1ea0976e63b709030954ccacd74`
- Pinned sing-box: `v1.13.16`
- Certification toolchain: `go1.26.5`

This record covers the complete embedded direction pair: native sing-box
inbounds selected by `-L`, native outbounds/endpoints selected by `-F`, and
their composition through an ordinary GOST chain. A port opening or successful
parse is never counted as a protocol data-plane pass.

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

The ordinary outbound matrix remains passing for Shadowsocks, SOCKS4/4a/5,
HTTP, VMess, VLESS, Reality/Vision, Trojan, AnyTLS, Hysteria2, TUIC, SSH,
Naive, WireGuard endpoint and Direct. Outbound chaining retains TCP, UDP,
domain destination, remote DNS, selector, detour and endpoint coverage.

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
private infrastructure addresses are intentionally excluded.

Environment: Linux amd64, Intel Xeon E5-2620 v4, GOMAXPROCS=2, Go 1.26.5,
sing-box v1.13.16, full feature tags, CGO disabled, loopback MTU 65536.

| Comparison | Official median | Gust median | Ratio | Gate | Result |
|---|---:|---:|---:|---:|---:|
| TCP throughput, 64 KiB echo | 695.00 MB/s | 695.09 MB/s | 100.01% | >=90% | PASS |
| UDP PPS, 1200-byte echo | 10,873 PPS | 10,849 PPS | 99.78% | >=90% | PASS |
| TCP round-trip p95 / p99 | 106.236 / 116.068 us | 106.765 / 116.983 us | 100.50% / 100.79% | <=110% | PASS |
| UDP round-trip p95 / p99 | 106.675 / 119.219 us | 106.076 / 118.899 us | 99.44% / 99.73% | <=110% | PASS |
| Internal egress UDP write | n/a | 151.4 ns/op median | 96 B, 2 allocs | <=128 B, <=2 allocs | PASS |
| Fixed-port reload pause | n/a | 3.76 ms median, 3.79 ms p95 | n/a | p95 <=5 ms | PASS |

Resource medians for the one-Box-per-`-L` baseline:

| Boxes | Startup | Live heap delta | Goroutine delta | FD delta | Median max RSS |
|---:|---:|---:|---:|---:|---:|
| 1 | 5.04 ms | 305,416 B | 8 | 4 | 19,546,112 B |
| 2 | 8.11 ms | 646,128 B | 16 | 8 | 20,262,912 B |
| 10 | 32.70 ms | 3,041,936 B | 80 | 40 | 23,941,120 B |
| 50 | 220.15 ms | 15,203,392 B | 400 | 200 | 41,582,592 B |

Every one of the 20 fresh-process resource samples returned goroutine and FD
counts exactly to its pre-start baseline after Close. These numbers prove the
direct integration boundary meets the release thresholds on the fixed runner;
they are not WAN or encrypted-protocol performance claims.

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
```

## Certification evidence

- Required Go CI: [run 31196849104](https://github.com/lovitus/gust/actions/runs/31196849104)
- Pinned/latest compatibility: [run 31195877341](https://github.com/lovitus/gust/actions/runs/31195877341)
- Six-platform build-only release rehearsal: [run 31197466032](https://github.com/lovitus/gust/actions/runs/31197466032)

The acceptance matrix is PASS. Publishing still requires a clean pushed source
branch and an `singbox-v*` tag contained in `origin/singbox-backend`; the tag
workflow independently repeats all six package builds before creating the
GitHub Release.
