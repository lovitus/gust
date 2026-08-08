#!/usr/bin/env python3
"""Ensure every top-level Docker E2E suite belongs to exactly one CI shard."""

from __future__ import annotations

import re
from collections import Counter
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
WORKFLOW = ROOT / ".github" / "workflows" / "go-ci.yml"
E2E = ROOT / "tests" / "e2e"


def validate() -> tuple[set[str], list[str]]:
    declared: list[str] = []
    workflow = WORKFLOW.read_text(encoding="utf-8")
    for expression in re.findall(r"^\s+tests:\s*'([^']+)'\s*$", workflow, re.MULTILINE):
        declared.extend(re.findall(r"Test[A-Za-z0-9_]+", expression))

    discovered: set[str] = set()
    for path in E2E.glob("*_test.go"):
        discovered.update(
            name
            for name in re.findall(
                r"^func\s+(Test[A-Za-z0-9_]+)\s*\(",
                path.read_text(encoding="utf-8"),
                re.MULTILINE,
            )
            if name != "TestMain"
        )

    counts = Counter(declared)
    duplicates = sorted(name for name, count in counts.items() if count != 1)
    missing = sorted(discovered - counts.keys())
    unknown = sorted(counts.keys() - discovered)
    if duplicates or missing or unknown:
        raise ValueError(
            f"invalid E2E shards: duplicates={duplicates}, missing={missing}, unknown={unknown}"
        )
    if len(re.findall(r"^\s+- suite:", workflow, re.MULTILINE)) < 2:
        raise ValueError("Docker E2E must remain split across independent shards")
    return discovered, declared


def main() -> None:
    discovered, _ = validate()
    print(f"Docker E2E shard coverage PASS: {len(discovered)} top-level suites")


if __name__ == "__main__":
    main()
