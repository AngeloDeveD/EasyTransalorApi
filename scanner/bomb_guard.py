from typing import Iterable
from config import settings

class ZipBomError(Exception): ...

def check_bomb(entries: Iterable[dict]) -> None:
    """entries: список с ключами compressed_size, uncompressed_size"""
    total_compressed = 0
    total_uncompressed = 0
    count = 0
    for e in entries:
        count += 1
        if count > settings.MAX_ENTRIES:
            raise ZipBomError(f"too many entries: {count}")

        cs = e["compressed_size"]
        ucs = e["uncompressed_size"]
        if cs > 0:
            ratio = ucs / cs
            if ratio > settings.MAX_COMPRESSION_RATIO:
                raise ZipBomError(
                    f"entry '{e.get('name')}' ratio {ratio:.1f} > "
                    f"{settings.MAX_COMPRESSION_RATIO}"
                )
        total_compressed += cs
        total_uncompressed += ucs
        if total_uncompressed > settings.MAX_TOTAL_UNCOMPRESSED:
            raise ZipBomError(
                f"total uncompressed {total_uncompressed} > "
                f"{settings.MAX_TOTAL_UNCOMPRESSED}"
            )