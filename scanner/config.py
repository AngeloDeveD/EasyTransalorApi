import math
import os
from pathlib import Path

class Settings:
    MAX_UPLOADED_BYTES = 5 * math.pow(1024, 3) #5гб
    MAX_TOTAL_UNCOMPRESSED = 10 * math.pow(1024, 3) #10гб
    MAX_ENTRIES = 10_000
    MAX_COMPRESSION_RATIO = 100
    MAX_ARCHIVE_DEPTH = 1
    CLAMD_SOCKET            = "/var/run/clamav/clamd.ctl"
    YARA_RULES_DIR          = "/etc/yara/rules"
    SANDBOX_API             = os.getenv("SANDBOX_API", "http://cape:8090")
    SANDBOX_TOKEN           = os.getenv("SANDBOX_TOKEN")
    SANDBOX_TIMEOUT         = 180                    # сек на анализ
    EXTRACT_ROOT            = Path("/tmp/av_extract")
    QUARANTINE_ROOT         = Path("/var/quarantine")

settings = Settings()