import os
import tempfile
import unittest
import zipfile

os.environ["CLAMD_ENABLED"] = "false"

try:
    from pipeline import ScanPipeline
except ImportError as exc:
    raise unittest.SkipTest(f"scanner dependencies are not installed: {exc}")


class ScanPipelineTests(unittest.TestCase):
    def test_clean_archive_returns_file_manifest(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            archive_path = os.path.join(temp_dir, "clean.zip")
            with zipfile.ZipFile(archive_path, "w") as zf:
                zf.writestr("translation/dialogues.json", "{}")
                zf.writestr("readme.txt", "ok")

            verdict = ScanPipeline.scan_archive(archive_path)

        self.assertEqual("approved", verdict["status"])
        self.assertTrue(verdict["is_safe"])
        self.assertEqual([], verdict["threats"])
        self.assertEqual(
            ["readme.txt", "translation/dialogues.json"],
            sorted(file_info["fileName"] for file_info in verdict["files"]),
        )
        for file_info in verdict["files"]:
            self.assertIn("hash", file_info)
            self.assertIn("size", file_info)

    def test_zip_slip_archive_is_rejected(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            archive_path = os.path.join(temp_dir, "zip-slip.zip")
            with zipfile.ZipFile(archive_path, "w") as zf:
                zf.writestr("../evil.txt", "outside")

            verdict = ScanPipeline.scan_archive(archive_path)

        self.assertEqual("rejected", verdict["status"])
        self.assertFalse(verdict["is_safe"])
        self.assertEqual([], verdict["files"])
        self.assertTrue(any("Zip Slip" in threat for threat in verdict["threats"]))


if __name__ == "__main__":
    unittest.main()
