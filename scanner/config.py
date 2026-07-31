import os
from pathlib import Path

class Settings:
    # Лимиты размеров/архивов
    MAX_UPLOADED_BYTES = 5 * 1024 ** 3          # 5 ГБ — максимальный размер самого архива
    # ClamAV физически не сканирует файлы крупнее 4 ГБ (жёсткий потолок демона).
    # В clamd.conf стоит 4000M; файлы крупнее этого порога уводим на ручную проверку,
    # а не отклоняем и не пытаемся стримить (иначе Broken pipe / ложный reject).
    CLAMAV_MAX_SCAN_BYTES = 4000 * 1024 ** 2    # 4000 МБ — синхронно с clamd.conf
    MAX_TOTAL_UNCOMPRESSED = 10 * 1024 ** 3     # 10 ГБ — суммарный распакованный объём
    MAX_ENTRIES = 10_000                        # максимум записей в архиве
    MAX_COMPRESSION_RATIO = 100                 # максимальное сжатие одной записи
    MAX_ARCHIVE_DEPTH = 1                        # запас на будущее (вложенные архивы)

    # Внешние сервисы
    # ClamAV: если задан CLAMD_HOST — работаем по TCP (нужно для docker-контейнеров),
    # иначе падаем на unix-сокет (локальный запуск демона на той же машине).
    CLAMD_HOST     = os.getenv("CLAMD_HOST")
    CLAMD_PORT     = int(os.getenv("CLAMD_PORT", "3310"))
    CLAMD_SOCKET   = os.getenv("CLAMD_SOCKET", "/var/run/clamav/clamd.ctl")
    YARA_RULES_DIR = os.getenv("YARA_RULES_DIR", "/etc/yara/rules")

    # Песочница по умолчанию отключена: статический анализ (ClamAV + YARA)
    # работает без неё. Включается только когда есть реальный CAPE/Cuckoo.
    SANDBOX_ENABLED = os.getenv("SANDBOX_ENABLED", "false").lower() in ("1", "true", "yes")
    SANDBOX_API     = os.getenv("SANDBOX_API", "")
    SANDBOX_TOKEN   = os.getenv("SANDBOX_TOKEN")
    SANDBOX_TIMEOUT = 180                        # сек на анализ

    # Пути
    EXTRACT_ROOT    = Path(os.getenv("EXTRACT_ROOT", "/tmp/av_extract"))
    QUARANTINE_ROOT = Path(os.getenv("QUARANTINE_ROOT", "/var/quarantine"))

    # Интеграция с основным Go-сервисом
    MAIN_API_URL = os.getenv("MAIN_API_URL", "http://localhost:8080/api/internal/scan-result")
    INTERNAL_KEY = os.getenv("InternalKey", "super_secret_cloud_key_998")

settings = Settings()
