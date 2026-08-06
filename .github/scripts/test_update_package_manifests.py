from __future__ import annotations

import importlib.util
import tempfile
import unittest
from pathlib import Path


SCRIPT = Path(__file__).with_name("update_package_manifests.py")
SPEC = importlib.util.spec_from_file_location("update_package_manifests", SCRIPT)
assert SPEC is not None and SPEC.loader is not None
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class LoadChecksumsTest(unittest.TestCase):
    def test_normalizes_sha256sum_path_markers(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "checksums.txt"
            path.write_text(
                "aaa  ./gost-darwin-arm64-3.2.10.tar.gz\n"
                "bbb *gost-windows-amd64-3.2.10.zip\n"
                "ccc  gost-linux-amd64-3.2.10.tar.gz\n",
                encoding="utf-8",
            )

            self.assertEqual(
                MODULE.load_checksums(path),
                {
                    "gost-darwin-arm64-3.2.10.tar.gz": "aaa",
                    "gost-windows-amd64-3.2.10.zip": "bbb",
                    "gost-linux-amd64-3.2.10.tar.gz": "ccc",
                },
            )


if __name__ == "__main__":
    unittest.main()
