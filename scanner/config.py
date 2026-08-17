import os

class Config:
    # Лимиты размеров
    MAX_ARCHIVE_SIZE: int = 5 * 1024 * 1024 * 1024       # 5 GB (макс. вес архива)
    MAX_TOTAL_EXTRACTED: int = 10 * 1024 * 1024 * 1024   # 10 GB (макс. распакованный размер)
    MAX_EXECUTABLE_SIZE: int = 100 * 1024 * 1024         # 100 MB (лимит на .dll/.exe/скрипты)
    MAX_FILE_COUNT: int = 50_000                         # Макс. кол-во файлов в архиве
    MAX_COMPRESSION_RATIO: float = 100.0                 # Защита от Zip-бомб
    CHUNK_SIZE: int = 2 * 1024 * 1024                    # Чтение по 2 MB

    # Настройки ClamAV
    CLAMAV_HOST: str = os.getenv("CLAMD_HOST", "127.0.0.1")
    CLAMAV_PORT: int = int(os.getenv("CLAMD_PORT", "3310"))
    CLAMAV_TIMEOUT: int = 300                            # 5 минут
    CLAMAV_ENABLED: bool = os.getenv("CLAMD_ENABLED", "true").lower() in ("true", "1", "yes")

    API_WEBHOOK_URL = os.getenv("API_WEBHOOK_URL", "http://api:8080/api/internal/scan-result")
    INTERNAL_KEY = os.getenv("INTERNAL_KEY", "your-secret-internal-key")

    # Расширения исполняемых файлов и скриптов
    EXECUTABLE_EXTENSIONS: set = {
        '.dll', '.exe', '.asi', '.cleo', '.so', '.dylib', '.sys',
        '.bat', '.cmd', '.ps1', '.vbs', '.js', '.lua', '.luac', '.py', '.pyc', '.lnk'
    }