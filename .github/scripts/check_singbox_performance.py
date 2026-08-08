#!/usr/bin/env python3
"""Validate the checked-in fixed-runner sing-box performance baseline."""

from __future__ import annotations

import argparse
import json
import math
import statistics
from pathlib import Path


def percentile(values: list[float], quantile: float) -> float:
    ordered = sorted(values)
    return ordered[max(0, math.ceil(len(ordered) * quantile) - 1)]


def require_samples(values: object, name: str) -> list[float]:
    if not isinstance(values, list) or len(values) < 5:
        raise ValueError(f"{name} must contain at least five samples")
    if any(not isinstance(value, (int, float)) or value < 0 for value in values):
        raise ValueError(f"{name} contains an invalid sample")
    return [float(value) for value in values]


def validate(baseline_path: Path, ref_path: Path) -> dict[str, float]:
    baseline = json.loads(baseline_path.read_text(encoding="utf-8"))
    pinned_ref = ref_path.read_text(encoding="utf-8").strip()
    source_ref = baseline["source"]["gust_x_revision"]
    if source_ref != pinned_ref:
        raise ValueError(
            f"performance baseline gust-x revision {source_ref} does not match pin {pinned_ref}"
        )
    if baseline.get("status") != "accepted":
        raise ValueError("performance baseline status must be accepted")

    gates = baseline["release_gates"]
    tcp = baseline["benchmarks"]["tcp"]
    tcp_official_rate = require_samples(tcp["official"]["mb_per_second"], "TCP official throughput")
    tcp_gust_rate = require_samples(tcp["gust"]["mb_per_second"], "TCP Gust throughput")
    tcp_official_ns = require_samples(tcp["official"]["ns_per_op"], "TCP official latency")
    tcp_gust_ns = require_samples(tcp["gust"]["ns_per_op"], "TCP Gust latency")
    tcp_ratio = statistics.median(tcp_gust_rate) / statistics.median(tcp_official_rate)
    tcp_p95_ratio = percentile(tcp_gust_ns, 0.95) / percentile(tcp_official_ns, 0.95)
    if tcp_ratio < gates["official_tcp_median_ratio_min"]:
        raise ValueError(f"TCP median throughput ratio {tcp_ratio:.4f} failed")
    if tcp_p95_ratio > gates["official_p95_latency_ratio_max"]:
        raise ValueError(f"TCP p95 latency ratio {tcp_p95_ratio:.4f} failed")

    udp = baseline["benchmarks"]["udp"]
    udp_official_ns = require_samples(udp["official"]["ns_per_op"], "UDP official latency")
    udp_gust_ns = require_samples(udp["gust"]["ns_per_op"], "UDP Gust latency")
    udp_pps_ratio = statistics.median(udp_official_ns) / statistics.median(udp_gust_ns)
    udp_p95_ratio = percentile(udp_gust_ns, 0.95) / percentile(udp_official_ns, 0.95)
    if udp_pps_ratio < gates["official_udp_median_pps_ratio_min"]:
        raise ValueError(f"UDP median PPS ratio {udp_pps_ratio:.4f} failed")
    if udp_p95_ratio > gates["official_p95_latency_ratio_max"]:
        raise ValueError(f"UDP p95 latency ratio {udp_p95_ratio:.4f} failed")

    egress = baseline["benchmarks"]["egress_udp_write"]
    egress_allocs = require_samples(egress["allocs_per_op"], "egress UDP allocations")
    egress_bytes = require_samples(egress["bytes_per_op"], "egress UDP bytes")
    if max(egress_allocs) > 2 or max(egress_bytes) > 128:
        raise ValueError("egress UDP allocation budget failed")
    if egress.get("payload_scaled_allocation") is not False:
        raise ValueError("egress UDP must explicitly record no payload-scaled allocation")

    runtime_handle = baseline["benchmarks"]["runtime_handle"]
    decode_ns = require_samples(
        runtime_handle["decode_and_construct"]["ns_per_op"],
        "decoded runtime handle latency",
    )
    retained = runtime_handle["retained"]
    retained_ns = require_samples(retained["ns_per_op"], "retained runtime handle latency")
    retained_bytes = require_samples(retained["bytes_per_op"], "retained runtime handle bytes")
    retained_allocs = require_samples(retained["allocs_per_op"], "retained runtime handle allocations")
    retained_ratio = statistics.median(retained_ns) / statistics.median(decode_ns)
    if retained_ratio > gates["retained_runtime_median_ratio_max"]:
        raise ValueError(f"retained runtime latency ratio {retained_ratio:.4f} failed")
    if max(retained_bytes) > gates["retained_runtime_bytes_max"]:
        raise ValueError("retained runtime byte budget failed")
    if max(retained_allocs) > gates["retained_runtime_allocs_max"]:
        raise ValueError("retained runtime allocation budget failed")

    scoped = runtime_handle["scoped_cache_hit"]
    if max(require_samples(scoped["bytes_per_op"], "scoped cache hit bytes")) > gates["retained_runtime_bytes_max"]:
        raise ValueError("scoped runtime cache byte budget failed")
    if max(require_samples(scoped["allocs_per_op"], "scoped cache hit allocations")) > gates["retained_runtime_allocs_max"]:
        raise ValueError("scoped runtime cache allocation budget failed")

    packet_read = baseline["benchmarks"]["packet_read"]
    direct_read = packet_read["direct"]
    proxy_read = packet_read["proxy_headroom"]
    if max(require_samples(direct_read["bytes_per_op"], "direct packet read bytes")) > gates["packet_read_direct_bytes_max"]:
        raise ValueError("direct packet read byte budget failed")
    if max(require_samples(direct_read["allocs_per_op"], "direct packet read allocations")) > gates["packet_read_direct_allocs_max"]:
        raise ValueError("direct packet read allocation budget failed")
    if max(require_samples(proxy_read["bytes_per_op"], "proxy packet read bytes")) > gates["packet_read_proxy_bytes_max"]:
        raise ValueError("proxy packet read byte budget failed")
    if max(require_samples(proxy_read["allocs_per_op"], "proxy packet read allocations")) > gates["packet_read_proxy_allocs_max"]:
        raise ValueError("proxy packet read allocation budget failed")
    if packet_read.get("payload_scaled_allocation") is not False:
        raise ValueError("packet read must explicitly record no payload-scaled allocation")

    reload_result = baseline["benchmarks"]["fixed_port_reload"]
    reload_ns = require_samples(reload_result["ns_per_op"], "fixed-port reload pause")
    if percentile(reload_ns, 0.95) > gates["fixed_port_reload_p95_ns_max"]:
        raise ValueError("fixed-port reload p95 pause budget failed")

    latency = baseline["latency_round_trip"]
    if latency.get("samples_per_trial", 0) < 1000 or latency.get("trials", 0) < 5:
        raise ValueError("latency profile must contain five trials of at least 1000 samples")
    latency_ratios: dict[str, float] = {}
    for network in ("tcp", "udp"):
        profile = latency[network]
        for quantile in ("p95_ns", "p99_ns"):
            official = require_samples(
                profile["official"][quantile], f"{network.upper()} latency official {quantile}"
            )
            gust = require_samples(
                profile["gust"][quantile], f"{network.upper()} latency Gust {quantile}"
            )
            ratio = statistics.median(gust) / statistics.median(official)
            latency_ratios[f"{network}_{quantile}_ratio"] = ratio
            if ratio > gates["official_p95_p99_latency_ratio_max"]:
                raise ValueError(
                    f"{network.upper()} {quantile} latency ratio {ratio:.4f} failed"
                )

    for count in ("1", "2", "10", "50"):
        resource = baseline["resources"].get(count)
        if not isinstance(resource, dict):
            raise ValueError(f"missing {count}-Box resource profile")
        for field in (
            "startup_ns",
            "heap_live_delta_bytes",
            "goroutine_live_delta",
            "fd_live_delta",
            "max_rss_bytes",
            "goroutine_after_close_delta",
            "fd_after_close_delta",
        ):
            require_samples(resource.get(field), f"{count}-Box {field}")
        if max(resource["goroutine_after_close_delta"]) > gates["goroutine_or_fd_after_close_delta_max"]:
            raise ValueError(f"{count}-Box goroutines did not return to baseline")
        if max(resource["fd_after_close_delta"]) > gates["goroutine_or_fd_after_close_delta_max"]:
            raise ValueError(f"{count}-Box file descriptors did not return to baseline")

    return {
        "tcp_ratio": tcp_ratio,
        "tcp_p95_ratio": tcp_p95_ratio,
        "udp_pps_ratio": udp_pps_ratio,
        "udp_p95_ratio": udp_p95_ratio,
        "retained_runtime_ratio": retained_ratio,
        **latency_ratios,
    }


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--baseline",
        type=Path,
        default=Path("SINGBOX-PERFORMANCE-BASELINE.json"),
    )
    parser.add_argument(
        "--gust-x-ref",
        type=Path,
        default=Path(".github/singbox-gust-x.ref"),
    )
    args = parser.parse_args()
    result = validate(args.baseline, args.gust_x_ref)
    print(
        "sing-box performance baseline PASS: "
        f"TCP median={result['tcp_ratio']:.4f}, TCP p95={result['tcp_p95_ratio']:.4f}, "
        f"UDP PPS={result['udp_pps_ratio']:.4f}, UDP p95={result['udp_p95_ratio']:.4f}, "
        f"retained runtime={result['retained_runtime_ratio']:.4f}, "
        f"round-trip p99 TCP={result['tcp_p99_ns_ratio']:.4f}, "
        f"UDP={result['udp_p99_ns_ratio']:.4f}"
    )


if __name__ == "__main__":
    main()
