import importlib.util
import unittest
from pathlib import Path


SCRIPT = Path(__file__).with_name("check_e2e_shards.py")
SPEC = importlib.util.spec_from_file_location("check_e2e_shards", SCRIPT)
MODULE = importlib.util.module_from_spec(SPEC)
assert SPEC.loader is not None
SPEC.loader.exec_module(MODULE)


class E2EShardTests(unittest.TestCase):
    def test_every_top_level_suite_is_declared_once(self):
        discovered, declared = MODULE.validate()
        self.assertEqual(discovered, set(declared))
        self.assertEqual(len(discovered), len(declared))


if __name__ == "__main__":
    unittest.main()
