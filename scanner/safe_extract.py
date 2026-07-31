import os, zipfile, stat
from pathlib import Path

from bomb_guard import check_bomb
from config import settings

# Опциональные зависимости — импортируем лениво, чтобы отсутствие unrar/py7zr
# не роняло весь модуль, если формат не используется.
try:
    import rarfile
except ImportError:  # pragma: no cover
    rarfile = None

try:
    import py7zr
except ImportError:  # pragma: no cover
    py7zr = None


class SecurityError(Exception): ...


def _validate_entry_path(dest: Path, name: str) -> Path:
    """Защита от path traversal и абсолютных путей."""
    # запрет абсолютных путей (unix + windows) и выхода наверх
    if name.startswith("/") or name.startswith("\\") or ".." in Path(name).parts:
        raise SecurityError(f"unsafe path: {name}")
    if Path(name).is_absolute() or (len(name) > 1 and name[1] == ":"):
        raise SecurityError(f"absolute path: {name}")

    real_dest = dest.resolve()
    target = (real_dest / name).resolve()
    if target != real_dest and not str(target).startswith(str(real_dest) + os.sep):
        raise SecurityError(f"path traversal: {name}")
    return target


def _stream_to_file(src, target: Path, total_written: int) -> int:
    """Потоково пишет src в target чанками, следит за суммарным лимитом.

    Возвращает новое значение total_written.
    """
    target.parent.mkdir(parents=True, exist_ok=True)
    with open(target, "wb") as dst:
        while chunk := src.read(64 * 1024):
            total_written += len(chunk)
            if total_written > settings.MAX_TOTAL_UNCOMPRESSED:
                raise SecurityError("uncompressed limit exceeded mid-extract")
            dst.write(chunk)
    os.chmod(target, 0o600)
    return total_written


def _prepare_dest(dest: Path) -> None:
    dest.mkdir(parents=True, exist_ok=True)
    os.chmod(dest, 0o700)


def safe_extract_zip(archive_path: Path, dest: Path) -> list[Path]:
    _prepare_dest(dest)
    extracted: list[Path] = []
    with zipfile.ZipFile(archive_path) as zf:
        # Проверка на бомбу до извлечения
        entries = [
            {
                "name": i.filename,
                "compressed_size": i.compress_size,
                "uncompressed_size": i.file_size,
            }
            for i in zf.infolist()
        ]
        check_bomb(entries)

        written = 0
        for info in zf.infolist():
            if info.is_dir() or info.filename.endswith("/"):
                continue
            # отключение символических ссылок
            mode = info.external_attr >> 16
            if mode and stat.S_ISLNK(mode):
                raise SecurityError(f"symlink entries are forbidden: {info.filename}")

            target = _validate_entry_path(dest, info.filename)
            with zf.open(info) as src:
                written = _stream_to_file(src, target, written)
            extracted.append(target)
    return extracted


def safe_extract_rar(archive_path: Path, dest: Path) -> list[Path]:
    if rarfile is None:
        raise SecurityError("rar support is not installed (rarfile/unrar missing)")
    _prepare_dest(dest)
    extracted: list[Path] = []
    with rarfile.RarFile(archive_path) as rf:
        entries = [
            {
                "name": i.filename,
                "compressed_size": i.compress_size or 0,
                "uncompressed_size": i.file_size or 0,
            }
            for i in rf.infolist()
        ]
        check_bomb(entries)

        written = 0
        for info in rf.infolist():
            if info.isdir():
                continue
            # rarfile помечает симлинки в атрибутах хоста unix
            if getattr(info, "is_symlink", None) and info.is_symlink():
                raise SecurityError(f"symlink entries are forbidden: {info.filename}")

            target = _validate_entry_path(dest, info.filename)
            with rf.open(info) as src:
                written = _stream_to_file(src, target, written)
            extracted.append(target)
    return extracted


def safe_extract_7z(archive_path: Path, dest: Path) -> list[Path]:
    if py7zr is None:
        raise SecurityError("7z support is not installed (py7zr missing)")
    _prepare_dest(dest)
    extracted: list[Path] = []
    with py7zr.SevenZipFile(archive_path, mode="r") as zf:
        # Проверка на бомбу по метаданным до распаковки
        entries = []
        for info in zf.list():
            if info.is_directory:
                continue
            entries.append(
                {
                    "name": info.filename,
                    "compressed_size": info.compressed or 0,
                    "uncompressed_size": info.uncompressed or 0,
                }
            )
            # валидация пути каждого имени заранее
            _validate_entry_path(dest, info.filename)
        check_bomb(entries)

        # py7zr не даёт потокового API по записям, читаем в bytes-буферы,
        # но общий распакованный объём уже ограничен check_bomb выше.
        real_dest = dest.resolve()
        for name, bio in zf.readall().items():
            target = _validate_entry_path(dest, name)
            if target == real_dest:
                continue
            target.parent.mkdir(parents=True, exist_ok=True)
            data = bio.read()
            with open(target, "wb") as dst:
                dst.write(data)
            os.chmod(target, 0o600)
            extracted.append(target)
    return extracted


# Диспетчер по формату
EXTRACTORS = {
    "zip": safe_extract_zip,
    "rar": safe_extract_rar,
    "7z": safe_extract_7z,
}


def safe_extract(fmt: str, archive_path: Path, dest: Path) -> list[Path]:
    extractor = EXTRACTORS.get(fmt)
    if extractor is None:
        raise SecurityError(f"unsupported archive type: {fmt}")
    return extractor(archive_path, dest)
