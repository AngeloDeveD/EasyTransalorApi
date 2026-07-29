import os, zipfile, shutil, tempfile, stat
from pathlib import Path

from bomb_guard import check_bomb
from config import settings

class SecurityError(Exception): ...

def _validate_entry_path(dest: Path, name: str) -> Path:
    #запрет абсолютных путей
    if name.startswith("/") or ".." in Path(name).parts:
        raise SecurityError(f"unsafe path: {name}")
    real_dest = dest.resolve()
    target = (real_dest/name).resolve()
    if target != real_dest and not str(target).startswith(str(real_dest) + os.sep):
        raise SecurityError(f"path traversal: {name}")
    return target

def safe_extract_zip(archive_path: Path, dest: Path) -> list[Path]:
    dest.mkdir(parents=True, exist_ok=True)
    os.chmod(dest, 0o700)
    exracted: list[Path] = []
    with zipfile.ZipFile(archive_path) as zf:
        #Проверка на бомбу
        entries = [
            {
                "name": i.filename,
                "compressed_size": i.compress_size,
                "uncompressed_size": i.file_size
            }
            for i in zf.infolist()
        ]
        check_bomb(entries)
        #Извлечение по одной записи
        for info in zf.infolist():
            if info.is_dir():
                continue
            if info.filename.endswith("/"):
                continue
            #отключение символических ссылок
            if info.external_attr >> 16 and stat.S_ISLNK(info.external_attr >> 16):
                raise SecurityError(f"symlink entries are forbidden: {info.filename}")
            target = _validate_entry_path(dest, info.filename)
            target.parent.mkdir(parents=True, exist_ok=True)

            # Потоковое чтение чанками, чтобы не гшрузить весь файл в память
            with zf.open(info) as src, open(target, "wb") as dst:
                written = 0
                while chunk := src.read(64 * 1024):
                    written += len(chunk)
                    if written > settings.MAX_TOTAL_UNCOMPRESSED:
                        raise SecurityError("uncompressed limit exceed mi-extract")
                    dst.write(chunk)
                os.chmod(target, 0o600)
            exracted.append(target)
    return exracted
