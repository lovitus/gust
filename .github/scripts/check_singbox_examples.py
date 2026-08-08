#!/usr/bin/env python3
"""Validate the stable structure of published embedded sing-box examples."""

from __future__ import annotations

import json
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
EXAMPLES = ROOT / "examples" / "singbox"


def main() -> int:
    files = sorted(EXAMPLES.glob("*.json"))
    if not files:
        raise SystemExit("no sing-box JSON examples found")
    for path in files:
        value = json.loads(path.read_text(encoding="utf-8"))
        if not isinstance(value, dict):
            raise SystemExit(f"{path}: top-level value must be an object")
        if "type" not in value and not any(key in value for key in ("inbounds", "outbounds", "endpoints")):
            raise SystemExit(f"{path}: expected a native object or complete config")
    print(f"validated {len(files)} sing-box example files")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
