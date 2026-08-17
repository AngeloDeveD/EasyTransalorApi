import zipfile
import stat
from config import Config

class SecurityException(Exception):
    pass

class BombGuard:
    @staticmethod
    def inspect_metadata(zf: zipfile.ZipFile, archive_size: int) -> None:
        """Предварительный анализ метаданных ZIP без распаковки."""
        if archive_size > Config.MAX_ARCHIVE_SIZE:
            raise SecurityException(
                f"Размер архива ({archive_size / (1024**3):.2f} GB) превышает лимит 5 GB"
            )

        infolist = zf.infolist()
        if len(infolist) > Config.MAX_FILE_COUNT:
            raise SecurityException(
                f"Превышен лимит файлов: {len(infolist)} > {Config.MAX_FILE_COUNT}"
            )

        total_uncompressed = 0
        total_compressed = 0

        for info in infolist:
            total_uncompressed += info.file_size
            total_compressed += info.compress_size

            # Проверка POSIX-атрибута символической ссылки (0o120000)
            mode = info.external_attr >> 16
            if stat.S_ISLNK(mode):
                raise SecurityException(f"Символические ссылки запрещены: '{info.filename}'")

        if total_uncompressed > Config.MAX_TOTAL_EXTRACTED:
            raise SecurityException(
                f"Суммарный объем ({total_uncompressed / (1024**3):.2f} GB) превышает лимит 10 GB"
            )

        actual_comp = max(total_compressed, archive_size, 1)
        ratio = total_uncompressed / actual_comp
        if ratio > Config.MAX_COMPRESSION_RATIO:
            raise SecurityException(
                f"Коэффициент сжатия ({ratio:.1f}:1) превышает норму. Признак Zip-бомбы."
            )