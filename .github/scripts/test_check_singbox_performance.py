import json
import tempfile
import unittest
from pathlib import Path

from check_singbox_performance import validate


class PerformanceBaselineTest(unittest.TestCase):
    def setUp(self):
        self.root = Path(__file__).resolve().parents[2]
        self.baseline = self.root / "SINGBOX-PERFORMANCE-BASELINE.json"
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.ref = Path(self.directory.name) / "singbox-gust-x.ref"
        data = json.loads(self.baseline.read_text(encoding="utf-8"))
        self.ref.write_text(data["source"]["gust_x_revision"] + "\n", encoding="utf-8")

    def test_repository_baseline_passes(self):
        result = validate(self.baseline, self.ref)
        self.assertGreaterEqual(result["tcp_ratio"], 0.9)
        self.assertGreaterEqual(result["udp_pps_ratio"], 0.9)
        self.assertLessEqual(result["tcp_p99_ns_ratio"], 1.1)
        self.assertLessEqual(result["udp_p99_ns_ratio"], 1.1)

    def test_revision_mismatch_fails(self):
        data = json.loads(self.baseline.read_text(encoding="utf-8"))
        data["source"]["gust_x_revision"] = "wrong"
        with tempfile.TemporaryDirectory() as directory:
            candidate = Path(directory) / "baseline.json"
            candidate.write_text(json.dumps(data), encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "does not match pin"):
                validate(candidate, self.ref)

    def test_runtime_allocation_regression_fails(self):
        data = json.loads(self.baseline.read_text(encoding="utf-8"))
        data["benchmarks"]["runtime_handle"]["retained"]["allocs_per_op"] = [2] * 5
        with tempfile.TemporaryDirectory() as directory:
            candidate = Path(directory) / "baseline.json"
            candidate.write_text(json.dumps(data), encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "retained runtime allocation budget"):
                validate(candidate, self.ref)

    def test_packet_read_allocation_regression_fails(self):
        data = json.loads(self.baseline.read_text(encoding="utf-8"))
        data["benchmarks"]["packet_read"]["proxy_headroom"]["bytes_per_op"] = [65] * 5
        with tempfile.TemporaryDirectory() as directory:
            candidate = Path(directory) / "baseline.json"
            candidate.write_text(json.dumps(data), encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "proxy packet read byte budget"):
                validate(candidate, self.ref)


if __name__ == "__main__":
    unittest.main()
