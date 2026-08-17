import os
import zipfile
import tempfile
from contextlib import contextmanager
from typing import Generator
from config import Config
from bomb_guard import BombGuard, SecurityException

class SafeExtractor:
    @staticmethod
    def _is_safe_path(base_dir: str, file_path: str) -> bool:
        """Защита от Zip Slip (Path Traversal)."""
        abs_base = os.path.abspath(base_dir)
        abs_dest = os.path.abspath(os.path.join(base_dir, file_path))
        return os.path.commonpath([abs_base, abs_dest]) == abs_base

    @classmethod
    @contextmanager
    def extract_zip(cls, archive_path: str) -> Generator[str, None, None]:
        archive_size = os.path.getsize(archive_path)

        with tempfile.TemporaryDirectory(prefix="scan_sandbox_") as temp_dir:
            with zipfile.ZipFile(archive_path, "r") as zf:
                # 1. Проверка метаданных
                BombGuard.inspect_metadata(zf, archive_size)

                total_extracted = 0

                # 2. Потоковая безопасная распаковка
                for member in zf.infolist():
                    if member.is_dir():
                        continue

                    if not cls._is_safe_path(temp_dir, member.filename):
                        raise SecurityException(f"Zip Slip атака в пути: '{member.filename}'")

                    dest_file_path = os.path.join(temp_dir, member.filename)
                    os.makedirs(os.path.dirname(dest_file_path), exist_ok=True)

                    with zf.open(member) as src, open(dest_file_path, "wb") as dst:
                        while chunk := src.read(Config.CHUNK_SIZE):
                            total_extracted += len(chunk)
                            if total_extracted > Config.MAX_TOTAL_EXTRACTED:
                                raise SecurityException("Превышен лимит 10 GB при распаковке!")
                            dst.write(chunk)

            yield temp_dir