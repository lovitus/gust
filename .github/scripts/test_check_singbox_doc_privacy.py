import tempfile
import unittest
from pathlib import Path

from check_singbox_doc_privacy import validate


class DocumentationPrivacyTest(unittest.TestCase):
    def candidate(self, text: str) -> Path:
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        path = Path(directory.name) / "candidate.md"
        path.write_text(text, encoding="utf-8")
        return path

    def test_documentation_addresses_and_domains_pass(self):
        validate(
            [
                self.candidate(
                    "127.0.0.1 0.0.0.0 192.0.2.10 198.51.100.2 "
                    "203.0.113.9 2001:db8::1 proxy.example.com github.com\n"
                )
            ]
        )

    def test_public_and_private_endpoints_fail(self):
        for value in (
            "8.8.8.8",
            "10.23.45.67",
            "2606:4700:4700::1111",
            "edge.private-network.us",
        ):
            with self.subTest(value=value), self.assertRaises(ValueError):
                validate([self.candidate(value)])

    def test_likely_secret_fails(self):
        with self.assertRaisesRegex(ValueError, "likely real credential"):
            validate([self.candidate('"private_key":"dGVzdC1rZXktdGhhdC1pcy10b28tbG9uZw=="')])


if __name__ == "__main__":
    unittest.main()
