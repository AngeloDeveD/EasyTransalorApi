from typing import Iterable
from config import settings

class ZipBombError(Exception): ...

def check_bomb(entries: Iterable[dict]) -> None:
    """Проверка на архивную бомбу.

    entries: список словарей с ключами name, compressed_size, uncompressed_size.
    Формат-независим — подходит для zip/rar/7z.
    """
    total_compressed = 0
    total_uncompressed = 0
    count = 0
    for e in entries:
        count += 1
        if count > settings.MAX_ENTRIES:
            raise ZipBombError(f"too many entries: {count}")

        cs = e["compressed_size"]
        ucs = e["uncompressed_size"]

        # Коэффициент сжатия отдельной записи
        if cs > 0:
            ratio = ucs / cs
            if ratio > settings.MAX_COMPRESSION_RATIO:
                raise ZipBombError(
                    f"entry '{e.get('name')}' ratio {ratio:.1f} > "
                    f"{settings.MAX_COMPRESSION_RATIO}"
                )

        total_compressed += cs
        total_uncompressed += ucs

        if total_uncompressed > settings.MAX_TOTAL_UNCOMPRESSED:
            raise ZipBombError(
                f"total uncompressed {total_uncompressed} > "
                f"{settings.MAX_TOTAL_UNCOMPRESSED}"
            )

    # Суммарный коэффициент сжатия — ловит "тонкие" бомбы,
    # где ни одна запись по отдельности не превышает лимит
    if total_compressed > 0:
        total_ratio = total_uncompressed / total_compressed
        if total_ratio > settings.MAX_COMPRESSION_RATIO:
            raise ZipBombError(
                f"overall ratio {total_ratio:.1f} > {settings.MAX_COMPRESSION_RATIO}"
            )
