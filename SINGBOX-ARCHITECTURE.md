# Embedded sing-box architecture and maintenance contract

This document describes the durable integration boundary. User-facing syntax
and examples are in `cmd/gost/SINGBOX_MANUAL.md`; data-plane evidence is in
`SINGBOX-VALIDATION.md`; the completed goal, retained code, dependency forks,
discarded experiments and upstream replay map are summarized in
`SINGBOX-INTEGRATION-FINAL-REPORT.md`.

## Product and dependency boundary

- `master` contains only general Gust and gust-x capabilities.
- `singbox-backend` is a downstream extension of `master` and is never merged
  back into it.
- gust-x `master` owns the generic top-level service factory. It has no
  sing-box types, tags, manifests or imports.
- Native parsing, resource extraction, runtime ownership and
  `__gust_egress` remain under `backend/singbox` on gust-x
  `singbox-backend`. Gust's runner contains no protocol-specific dispatch.
- Shared fixes are implemented and tested on both `master` branches first,
  then merged forward into both extension branches.

## Data and ownership paths

Native outbound:

```text
GOST service -> GOST chain/hops -> sing-box Transport -> selected native outbound
```

Native inbound:

```text
selected native inbound -> __gust_egress -> GOST Router -> ordinary GOST chain
```

There is no child process, IPC, loopback SOCKS bridge or hidden adapter
listener. The inbound is a top-level `service.Service`; the outbound is a
self-dialing Transport. Those abstractions are intentionally separate.

An inbound Box is owned by exactly one `-L` service and closes with that
service. Outbound Boxes use the runtime pool and connection leases. A retained
lease keeps an old runtime alive during reload until the active TCP/UDP
connection closes.

## Repeated listeners and chains

For:

```text
-L A -L B -L C -F W -F X -F Y -F Z
```

Gust creates three independent inbound services/Boxes. Each service refers to
the same generated GOST chain definition containing `W -> X -> Y -> Z`.
Runtime ownership is independent; chain configuration is reused. One-Box-per-
listener is the certified correctness baseline because it provides simple
socket ownership, reload and failure isolation.

Automatic inbound grouping is deliberately not a release requirement. The
measured per-Box resource profile is linear and releases every goroutine and
file descriptor on close. Grouping would require cross-service router-scope
dispatch and ambiguous handling of native background traffic; it should be
added only if fixed-runner evidence shows a material benefit that justifies
that complexity.

## Configuration contract

`-L` and `-F` use the same source and merge model:

```text
selected full-config object
  < json object overlay
  < URI authority and userinfo
  < ordinary query assignments
```

Both directions support CLI paths, exact `path:=JSON`, readable inline JSON,
JSON files, full configs and mixed overrides. Direction-specific native
schemas remain separate so an outbound-only field cannot silently become an
inbound field.

`inbound`, `outbound`, `endpoint` and `activate_inbounds` are Gust adapter
controls. They are removed before native decoding. Full-config `-L` starts
only the selected inbound and its exact declared detour closure. A foreign
`route.final` is rejected: unmatched/final traffic belongs to the GOST chain,
while explicit native route rules retain priority.

`-singboxcheck` parses and validates this effective static configuration
without startup. It prints structure only. `-O` prints the complete effective
configuration and must be treated as secret-bearing. Parsing still reads the
declared sources, including fetching an explicitly supplied remote config URL.

## TCP, UDP and failure behavior

- TCP returns the native connection directly; the adapter does not copy the
  payload through an internal proxy.
- `__gust_egress` implements connection and packet-association handlers. One
  UDP association can preserve multiple IP or domain destinations.
- Direct connected UDP reads into the caller buffer. Proxy packet reads use a
  bounded headroom pool so native address headers cannot truncate payloads.
- Prefix-chain, router, DNS, tag and resource errors fail closed. There is no
  system-direct retry.
- Native background activity without an unambiguous router scope may not guess
  a route or silently use system direct.
- Registry presence is not activation permission. Naive outbound, Tailscale
  endpoint, DHCP DNS and the resolved service are globally rejected; a
  WireGuard endpoint is supported only without a preceding GOST prefix.

## Performance boundary

Parsing, canonicalization and Box construction are startup/reload operations,
not per-packet work. Ordinary route copies retain a selected handle on the
started runtime. Distinct non-Direct GOST prefix scopes are cached; the first
connection creates the scope and later connections retain a lightweight
handle. Cache anchors close with their owning Transport.

The costs that remain are native protocol cryptography/handshakes, ordinary
GOST service and chain work, a larger release binary, one inbound Box per
listener, and one cached runtime per distinct embedded outbound route scope.
Fixed-runner gates compare the integration boundary with the same pinned
sing-box implementation's native direct path; they do not claim a universal
WAN percentage.

## Upgrade and release rules

The checked-in available/removed protocol manifest is generated from the
pinned module with build-tag awareness. Unknown resource-owning native types
fail closed until they have a certified extractor. A sing-box upgrade must
regenerate the manifest and repeat schema, protocol, privileged Linux,
security, performance, native-platform and package smoke matrices.

Only `singbox-v*` tags contained in `origin/singbox-backend` may publish this
flavor. The release does not update the standard latest channel. Source
branches must be clean, pushed and green before tagging.
