#!/usr/bin/env python3
"""Reject private endpoints and likely credentials in published sing-box docs."""

from __future__ import annotations

import argparse
import ipaddress
import re
from pathlib import Path


DEFAULT_DOCS = tuple(sorted(Path(".").glob("SINGBOX*.md"))) + (
    Path("cmd/gost/SINGBOX_MANUAL.md"),
)

IPV4_RE = re.compile(
    r"(?<![0-9.])(?:[0-9]{1,3}\.){3}[0-9]{1,3}(?:/[0-9]{1,2})?(?![0-9.])"
)
IPV6_RE = re.compile(
    r"(?<![0-9A-Fa-f:])(?:[0-9A-Fa-f]{0,4}:){2,7}"
    r"[0-9A-Fa-f]{0,4}(?:/[0-9]{1,3})?(?![0-9A-Fa-f:])"
)
DOMAIN_RE = re.compile(
    r"(?<![A-Za-z0-9_-])(?:[A-Za-z0-9-]+\.)+(?:com|net|org|io|us|cn|dev|app|cloud|xyz|top|me|co|ai|test)(?![A-Za-z0-9_-])",
    re.IGNORECASE,
)
LIKELY_SECRET_RE = re.compile(
    r"(?i)(?:password|passwd|private_key|token|secret)"
    r"(?:\\?[\"']?\s*[:=]\s*\\?[\"']?)"
    r"([A-Za-z0-9_+/=-]{24,})"
)

ALLOWED_DOMAINS = {"github.com"}
ALLOWED_EXAMPLE_SUFFIXES = (
    ".example.com",
    ".example.net",
    ".example.org",
    ".example.test",
)


def location(text: str, offset: int) -> tuple[int, int]:
    line = text.count("\n", 0, offset) + 1
    previous_newline = text.rfind("\n", 0, offset)
    column = offset - previous_newline
    return line, column


def allowed_ip(value: str) -> bool:
    candidate = value
    try:
        network = ipaddress.ip_network(candidate, strict=False)
    except ValueError:
        return False
    address = network.network_address if "/" in candidate else ipaddress.ip_address(candidate)
    if address.version == 4:
        if address.is_loopback:
            return True
        if candidate in {"0.0.0.0", "0.0.0.0/0"}:
            return True
        return any(
            address in documentation
            for documentation in (
                ipaddress.ip_network("192.0.2.0/24"),
                ipaddress.ip_network("198.51.100.0/24"),
                ipaddress.ip_network("203.0.113.0/24"),
            )
        )
    if candidate in {"::", "::/0", "::1", "::1/128"}:
        return True
    return address in ipaddress.ip_network("2001:db8::/32")


def allowed_domain(value: str) -> bool:
    domain = value.lower().rstrip(".")
    return (
        domain in ALLOWED_DOMAINS
        or domain in {"example.com", "example.net", "example.org", "example.test"}
        or domain.endswith(ALLOWED_EXAMPLE_SUFFIXES)
        or domain.endswith(".test")
    )


def scan(path: Path) -> list[str]:
    text = path.read_text(encoding="utf-8")
    findings: list[str] = []
    occupied_ip_spans: list[tuple[int, int]] = []
    for pattern in (IPV4_RE, IPV6_RE):
        for match in pattern.finditer(text):
            candidate = match.group(0)
            try:
                ipaddress.ip_network(candidate, strict=False)
            except ValueError:
                continue
            occupied_ip_spans.append(match.span())
            if not allowed_ip(candidate):
                line, column = location(text, match.start())
                findings.append(
                    f"{path}:{line}:{column}: non-documentation IP literal {candidate!r}"
                )
    for match in DOMAIN_RE.finditer(text):
        if any(start <= match.start() < end for start, end in occupied_ip_spans):
            continue
        candidate = match.group(0)
        if not allowed_domain(candidate):
            line, column = location(text, match.start())
            findings.append(f"{path}:{line}:{column}: non-example domain {candidate!r}")
    for match in LIKELY_SECRET_RE.finditer(text):
        candidate = match.group(1)
        if candidate.upper().startswith(("REPLACE_", "EXAMPLE_")):
            continue
        line, column = location(text, match.start(1))
        findings.append(
            f"{path}:{line}:{column}: likely real credential; replace it with a placeholder"
        )
    return findings


def validate(paths: list[Path]) -> None:
    findings: list[str] = []
    for path in paths:
        if not path.is_file():
            findings.append(f"{path}: required documentation file is missing")
            continue
        findings.extend(scan(path))
    if findings:
        raise ValueError("documentation privacy check failed:\n" + "\n".join(findings))


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("paths", nargs="*", type=Path)
    args = parser.parse_args()
    paths = args.paths or list(DEFAULT_DOCS)
    try:
        validate(paths)
    except ValueError as error:
        parser.exit(1, f"{error}\n")
    print(
        "sing-box documentation privacy PASS: "
        + ", ".join(str(path) for path in paths)
    )


if __name__ == "__main__":
    main()
