# -*- coding: utf-8 -*-
"""Тесты сканера архивов (stdlib unittest, без pytest).

Запуск из каталога scanner/:
    python -m unittest test_scanner -v
или просто:
    python test_scanner.py

Тесты, которым нужен реально работающий ClamAV/YARA/песочница, эти зависимости
не поднимают: clamav_scan подменяется, каталоги распаковки/карантина
переопределяются на временные, а YARA-правил нет (yara_scan вернёт []).
"""

import os
import sys
import shutil
import zipfile
import tempfile
import unittest
from pathlib import Path

# Чтобы `import config`, `import pipeline` и т.д. работали независимо от cwd.
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from magic_bytes import detect_type
from bomb_guard import check_bomb, ZipBombError
from config import settings
import safe_extract
from safe_extract import (
    safe_extract as do_extract,
    safe_extract_zip,
    _validate_entry_path,
    SecurityError,
)
import sandbox
import pipeline


def _force_rmtree(path):
    """rmtree, устойчивый к файлам карантина 0o400 на Windows."""
    p = Path(path)
    if not p.exists():
        return
    for root, _dirs, files in os.walk(path):
        for name in files:
            try:
                os.chmod(os.path.join(root, name), 0o700)
            except OSError:
                pass
    shutil.rmtree(path, ignore_errors=True)


# --------------------------------------------------------------------------- #
#  magic_bytes.detect_type
# --------------------------------------------------------------------------- #
class TestDetectType(unittest.TestCase):
    def test_zip_signatures(self):
        self.assertEqual(detect_type(b"PK\x03\x04rest-of-header"), "zip")
        self.assertEqual(detect_type(b"PK\x05\x06"), "zip")
        self.assertEqual(detect_type(b"PK\x07\x08"), "zip")

    def test_rar_signatures(self):
        self.assertEqual(detect_type(b"Rar!\x1a\x07\x00"), "rar")
        self.assertEqual(detect_type(b"Rar!\x1a\x07\x01\x00extra"), "rar")

    def test_7z_signature(self):
        self.assertEqual(detect_type(b"7z\xbc\xaf\x27\x1c\x00\x04"), "7z")

    def test_unknown_and_empty(self):
        self.assertIsNone(detect_type(b"just some plain text"))
        self.assertIsNone(detect_type(b""))
        self.assertIsNone(detect_type(b"%PDF-1.7"))


# --------------------------------------------------------------------------- #
#  bomb_guard.check_bomb
# --------------------------------------------------------------------------- #
class TestCheckBomb(unittest.TestCase):
    @staticmethod
    def _entry(name, cs, ucs):
        return {"name": name, "compressed_size": cs, "uncompressed_size": ucs}

    def test_empty_ok(self):
        # Пустой список записей не считается бомбой.
        check_bomb([])

    def test_normal_archive_ok(self):
        entries = [self._entry(f"f{i}.txt", 1000, 3000) for i in range(5)]
        check_bomb(entries)  # ratio 3 — не бомба

    def test_too_many_entries(self):
        entries = (self._entry(f"f{i}", 1, 1) for i in range(settings.MAX_ENTRIES + 1))
        with self.assertRaises(ZipBombError):
            check_bomb(entries)

    def test_single_entry_ratio(self):
        # Одна запись с коэффициентом сжатия > 100.
        entries = [self._entry("evil", 10, 10 * (settings.MAX_COMPRESSION_RATIO + 5))]
        with self.assertRaises(ZipBombError):
            check_bomb(entries)

    def test_total_uncompressed_limit(self):
        # Коэффициент в норме (60), но суммарный распакованный объём > 10 ГБ.
        cs = 200 * 1024 ** 2          # 200 МБ
        ucs = 12 * 1024 ** 3          # 12 ГБ, ratio ~61 <= 100
        with self.assertRaises(ZipBombError):
            check_bomb([self._entry("big", cs, ucs)])

    def test_thin_bomb_overall_ratio(self):
        # "Тонкая" бомба: запись с compressed_size == 0 не проверяется по
        # отдельному ratio, но задирает суммарный коэффициент сжатия.
        entries = [
            self._entry("a.txt", 100, 100),   # ratio 1
            self._entry("b.bin", 0, 50_000),  # индивидуальная проверка пропущена
        ]
        with self.assertRaises(ZipBombError):
            check_bomb(entries)


# --------------------------------------------------------------------------- #
#  safe_extract — валидация путей и реальная распаковка zip
# --------------------------------------------------------------------------- #
class TestValidateEntryPath(unittest.TestCase):
    def setUp(self):
        self.dest = Path(tempfile.mkdtemp(prefix="dest_"))
        self.addCleanup(_force_rmtree, self.dest)

    def test_rejects_absolute_unix(self):
        with self.assertRaises(SecurityError):
            _validate_entry_path(self.dest, "/etc/passwd")

    def test_rejects_leading_backslash(self):
        with self.assertRaises(SecurityError):
            _validate_entry_path(self.dest, "\\\\server\\share\\x")

    def test_rejects_parent_traversal(self):
        with self.assertRaises(SecurityError):
            _validate_entry_path(self.dest, "../evil.txt")

    def test_rejects_windows_drive(self):
        with self.assertRaises(SecurityError):
            _validate_entry_path(self.dest, "C:\\Windows\\system32")

    def test_accepts_normal_relative_name(self):
        target = _validate_entry_path(self.dest, "sub/file.txt")
        self.assertTrue(str(target).startswith(str(self.dest.resolve())))


class TestSafeExtractZip(unittest.TestCase):
    def setUp(self):
        self.tmp = Path(tempfile.mkdtemp(prefix="sx_"))
        self.addCleanup(_force_rmtree, self.tmp)

    def _make_zip(self, path, entries, compress=zipfile.ZIP_DEFLATED):
        with zipfile.ZipFile(path, "w", compress) as zf:
            for name, data in entries:
                zf.writestr(name, data)

    def test_roundtrip(self):
        arc = self.tmp / "ok.zip"
        self._make_zip(arc, [("a.txt", "hello"), ("dir/b.txt", "world")])
        dest = self.tmp / "out"

        extracted = safe_extract_zip(arc, dest)
        self.assertEqual(len(extracted), 2)

        names = {p.name for p in extracted}
        self.assertEqual(names, {"a.txt", "b.txt"})
        self.assertEqual((dest / "a.txt").read_text(), "hello")
        self.assertEqual((dest / "dir" / "b.txt").read_text(), "world")

    def test_zip_bomb_detected(self):
        # 300 КБ нулей ужимаются в считанные байты → ratio >> 100.
        arc = self.tmp / "bomb.zip"
        self._make_zip(arc, [("zeros.bin", b"\x00" * (300 * 1024))])
        with self.assertRaises(ZipBombError):
            safe_extract_zip(arc, self.tmp / "out")

    def test_path_traversal_entry(self):
        arc = self.tmp / "trav.zip"
        self._make_zip(arc, [("../escape.txt", "pwn")])
        with self.assertRaises(SecurityError):
            safe_extract_zip(arc, self.tmp / "out")

    def test_unsupported_format(self):
        with self.assertRaises(SecurityError):
            do_extract("tar", self.tmp / "whatever", self.tmp / "out")


# --------------------------------------------------------------------------- #
#  sandbox.is_malicious
# --------------------------------------------------------------------------- #
class TestIsMalicious(unittest.TestCase):
    def test_high_malscore(self):
        self.assertTrue(sandbox.is_malicious({"malscore": 8.0}))

    def test_threshold_malscore(self):
        # Порог включительный (>= 7.0).
        self.assertTrue(sandbox.is_malicious({"malscore": 7.0}))

    def test_critical_signature(self):
        self.assertTrue(
            sandbox.is_malicious({"malscore": 1.0, "signatures": [{"severity": 5}]})
        )

    def test_benign_report(self):
        self.assertFalse(
            sandbox.is_malicious({"malscore": 2.0, "signatures": [{"severity": 3}]})
        )
        self.assertFalse(sandbox.is_malicious({}))


# --------------------------------------------------------------------------- #
#  pipeline.run_pipeline — асинхронный сквозной прогон
# --------------------------------------------------------------------------- #
class TestPipeline(unittest.IsolatedAsyncioTestCase):
    def setUp(self):
        self.tmp = Path(tempfile.mkdtemp(prefix="pipe_"))
        self.addCleanup(_force_rmtree, self.tmp)

        # Переопределяем каталоги на временные и восстанавливаем после теста.
        self._orig_extract = settings.EXTRACT_ROOT
        self._orig_quarant = settings.QUARANTINE_ROOT
        self._orig_sandbox = settings.SANDBOX_ENABLED
        self._orig_clamav = pipeline.clamav_scan

        settings.EXTRACT_ROOT = self.tmp / "extract"
        settings.QUARANTINE_ROOT = self.tmp / "quarantine"
        settings.SANDBOX_ENABLED = False

    def tearDown(self):
        settings.EXTRACT_ROOT = self._orig_extract
        settings.QUARANTINE_ROOT = self._orig_quarant
        settings.SANDBOX_ENABLED = self._orig_sandbox
        pipeline.clamav_scan = self._orig_clamav

    def _make_zip(self, name, entries):
        arc = self.tmp / name
        with zipfile.ZipFile(arc, "w", zipfile.ZIP_DEFLATED) as zf:
            for entry_name, data in entries:
                zf.writestr(entry_name, data)
        return arc

    async def test_clean_no_executables(self):
        # ClamAV чист, исполняемых файлов нет → clean.
        async def fake_scan(path):
            return True, ""

        pipeline.clamav_scan = fake_scan
        arc = self._make_zip("clean.zip", [("readme.txt", "hi")])

        res = await pipeline.run_pipeline(arc, "clean.zip")
        self.assertEqual(res.status, "clean")
        self.assertEqual(res.threats, [])

    async def test_clean_executable_sandbox_disabled(self):
        # Есть .exe, но ClamAV/YARA чисто и песочница выключена → clean.
        async def fake_scan(path):
            return True, ""

        pipeline.clamav_scan = fake_scan
        arc = self._make_zip("app.zip", [("prog.exe", b"MZ\x90\x00stub")])

        res = await pipeline.run_pipeline(arc, "app.zip")
        self.assertEqual(res.status, "clean")

    async def test_infected_clamav_quarantines(self):
        # ClamAV помечает файл заражённым → infected + карантин.
        async def fake_scan(path):
            return False, "Eicar-Test-Signature"

        pipeline.clamav_scan = fake_scan
        arc = self._make_zip("virus.zip", [("payload.txt", "x")])

        res = await pipeline.run_pipeline(arc, "virus.zip")
        self.assertEqual(res.status, "infected")
        self.assertEqual(len(res.threats), 1)
        self.assertEqual(res.threats[0]["engine"], "clamav")

        quarantined = list(settings.QUARANTINE_ROOT.glob("*"))
        self.assertEqual(len(quarantined), 1)

    async def test_unsupported_archive_type(self):
        # Файл без сигнатуры архива → SecurityError.
        bogus = self.tmp / "notarchive.zip"
        bogus.write_bytes(b"this is just text, not an archive")

        with self.assertRaises(SecurityError):
            await pipeline.run_pipeline(bogus, "notarchive.zip")


if __name__ == "__main__":
    unittest.main(verbosity=2)
