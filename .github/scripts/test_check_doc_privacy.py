import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest import mock

from check_doc_privacy import diff_added_lines, validate_changed, validate_paths


class DocumentationPrivacyTest(unittest.TestCase):
    def candidate(self, text: str) -> Path:
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        path = Path(directory.name) / "candidate.md"
        path.write_text(text, encoding="utf-8")
        return path

    def test_documentation_addresses_domains_and_placeholders_pass(self):
        validate_paths(
            [
                self.candidate(
                    "127.0.0.1 0.0.0.0 192.0.2.10 198.51.100.2 "
                    "203.0.113.9 2001:db8::1 proxy.example.com github.com "
                    "token=REPLACE_WITH_TOKEN\n"
                )
            ]
        )

    def test_public_private_endpoints_and_secrets_fail(self):
        for value in (
            "8.8.8.8",
            "10.23.45.67",
            "2606:4700:4700::1111",
            "edge.private-network.us",
            '"private_key":"dGVzdC1rZXktdGhhdC1pcy10b28tbG9uZw=="',
        ):
            with self.subTest(value=value), self.assertRaises(ValueError):
                validate_paths([self.candidate(value)])

    def test_diff_parser_only_returns_added_markdown_lines(self):
        diff = """diff --git a/guide.md b/guide.md
--- a/guide.md
+++ b/guide.md
@@ -2 +2,2 @@
-old.example.test
+new.example.test
+203.0.113.5
"""
        completed = subprocess.CompletedProcess([], 0, stdout=diff)
        with mock.patch("subprocess.run", return_value=completed):
            additions = diff_added_lines("base")
        self.assertEqual(
            [(str(item.path), item.number, item.text) for item in additions],
            [("guide.md", 2, "new.example.test"), ("guide.md", 3, "203.0.113.5")],
        )

    def test_changed_validation_rejects_added_real_endpoint(self):
        diff = """diff --git a/guide.md b/guide.md
--- a/guide.md
+++ b/guide.md
@@ -0,0 +1 @@
+server: edge.private-network.us
"""
        completed = subprocess.CompletedProcess([], 0, stdout=diff)
        with mock.patch("subprocess.run", return_value=completed):
            with self.assertRaisesRegex(ValueError, "guide.md:1"):
                validate_changed("base")


if __name__ == "__main__":
    unittest.main()
