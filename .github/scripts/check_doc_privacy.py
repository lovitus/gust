#!/usr/bin/env python3
"""Reject private endpoints and likely credentials in new documentation."""

from __future__ import annotations

import argparse
import ipaddress
import re
import subprocess
from dataclasses import dataclass
from pathlib import Path


IPV4_RE = re.compile(
    r"(?<![0-9.])(?:[0-9]{1,3}\.){3}[0-9]{1,3}(?:/[0-9]{1,2})?(?![0-9.])"
)
IPV6_RE = re.compile(
    r"(?<![0-9A-Fa-f:])(?:[0-9A-Fa-f]{0,4}:){2,7}"
    r"[0-9A-Fa-f]{0,4}(?:/[0-9]{1,3})?(?![0-9A-Fa-f:])"
)
DOMAIN_RE = re.compile(
    r"(?<![A-Za-z0-9_-])(?:[A-Za-z0-9-]+\.)+"
    r"(?:com|net|org|io|us|cn|dev|app|cloud|xyz|top|me|co|ai|test)"
    r"(?![A-Za-z0-9_-])",
    re.IGNORECASE,
)
LIKELY_SECRET_RE = re.compile(
    r"(?i)(?:password|passwd|private_key|token|secret)"
    r"(?:\\?[\"']?\s*[:=]\s*\\?[\"']?)"
    r"([A-Za-z0-9_+/=-]{24,})"
)
DIFF_HEADER_RE = re.compile(r"^\+\+\+ b/(.+)$")
DIFF_HUNK_RE = re.compile(r"^@@ -\d+(?:,\d+)? \+(\d+)(?:,\d+)? @@")

ALLOWED_DOMAINS = {
    "docs.docker.com",
    "github.com",
    "gost.run",
    "www.youtube.com",
    "t.me",
    "groups.google.com",
}
ALLOWED_EXAMPLE_SUFFIXES = (
    ".example.com",
    ".example.net",
    ".example.org",
    ".example.test",
)


@dataclass(frozen=True)
class AddedLine:
    path: Path
    number: int
    text: str


def allowed_ip(value: str) -> bool:
    try:
        network = ipaddress.ip_network(value, strict=False)
    except ValueError:
        return False
    address = network.network_address if "/" in value else ipaddress.ip_address(value)
    if address.version == 4:
        if address.is_loopback or value in {"0.0.0.0", "0.0.0.0/0"}:
            return True
        return any(
            address in documentation
            for documentation in (
                ipaddress.ip_network("192.0.2.0/24"),
                ipaddress.ip_network("198.51.100.0/24"),
                ipaddress.ip_network("203.0.113.0/24"),
            )
        )
    if value in {"::", "::/0", "::1", "::1/128"}:
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


def scan_text(text: str, source: str, line: int = 1) -> list[str]:
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
                findings.append(
                    f"{source}:{line}:{match.start() + 1}: "
                    f"non-documentation IP literal {candidate!r}"
                )
    for match in DOMAIN_RE.finditer(text):
        if any(start <= match.start() < end for start, end in occupied_ip_spans):
            continue
        candidate = match.group(0)
        if not allowed_domain(candidate):
            findings.append(
                f"{source}:{line}:{match.start() + 1}: non-example domain {candidate!r}"
            )
    for match in LIKELY_SECRET_RE.finditer(text):
        candidate = match.group(1)
        if candidate.upper().startswith(("REPLACE_", "EXAMPLE_")):
            continue
        findings.append(
            f"{source}:{line}:{match.start(1) + 1}: likely real credential; "
            "replace it with a placeholder"
        )
    return findings


def validate_paths(paths: list[Path]) -> None:
    findings: list[str] = []
    for path in paths:
        if not path.is_file():
            findings.append(f"{path}: documentation file is missing")
            continue
        for number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            findings.extend(scan_text(line, str(path), number))
    if findings:
        raise ValueError("documentation privacy check failed:\n" + "\n".join(findings))


def diff_added_lines(base: str) -> list[AddedLine]:
    if not base or set(base) == {"0"}:
        base = "HEAD^"
    result = subprocess.run(
        [
            "git",
            "diff",
            "--unified=0",
            "--no-color",
            f"{base}..HEAD",
            "--",
            "*.md",
        ],
        check=True,
        text=True,
        stdout=subprocess.PIPE,
    )
    current_path: Path | None = None
    new_line = 0
    additions: list[AddedLine] = []
    for raw_line in result.stdout.splitlines():
        header = DIFF_HEADER_RE.match(raw_line)
        if header:
            current_path = Path(header.group(1))
            continue
        hunk = DIFF_HUNK_RE.match(raw_line)
        if hunk:
            new_line = int(hunk.group(1))
            continue
        if raw_line.startswith("+") and not raw_line.startswith("+++"):
            if current_path is not None:
                additions.append(AddedLine(current_path, new_line, raw_line[1:]))
            new_line += 1
        elif raw_line.startswith(" "):
            new_line += 1
    return additions


def validate_changed(base: str) -> None:
    findings: list[str] = []
    for addition in diff_added_lines(base):
        findings.extend(
            scan_text(addition.text, str(addition.path), addition.number)
        )
    if findings:
        raise ValueError("documentation privacy check failed:\n" + "\n".join(findings))


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("paths", nargs="*", type=Path)
    parser.add_argument("--changed-from", metavar="GIT_REVISION")
    args = parser.parse_args()
    try:
        if args.changed_from is not None:
            if args.paths:
                parser.error("paths cannot be combined with --changed-from")
            validate_changed(args.changed_from)
        elif args.paths:
            validate_paths(args.paths)
        else:
            parser.error("provide Markdown paths or --changed-from")
    except (subprocess.CalledProcessError, ValueError) as error:
        parser.exit(1, f"{error}\n")
    print("documentation privacy PASS")


if __name__ == "__main__":
    main()
