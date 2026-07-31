import clamd, asyncio, logging
from pathlib import Path

from config import settings


class ScannerUnavailableError(Exception):
    """Движок ClamAV не смог просканировать файл (сокет закрыт, демон недоступен).

    Это НЕ вердикт «заражён» — это отказ движка. Пайплайн уводит такой файл
    на ручную проверку, а не отклоняет: иначе легитимный файл ложно попадёт
    в rejected_by_scanner (как было при Broken pipe на крупных файлах).
    """


# TCP, когда сканер и clamd — разные контейнеры; unix-сокет для локального демона.
if settings.CLAMD_HOST:
    _cd = clamd.ClamdNetworkSocket(settings.CLAMD_HOST, settings.CLAMD_PORT)
else:
    _cd = clamd.ClamdUnixSocket(settings.CLAMD_SOCKET)

async def clamav_scan(path: Path) -> tuple[bool, str]:
    """Возвращает (clean, description).

    Бросает ScannerUnavailableError, если демон закрыл соединение или недоступен
    (BrokenPipeError / clamd.ConnectionError). Вызывающий код обязан отличать
    этот отказ движка от честного вердикта FOUND.
    """

    logger = logging.getLogger("uvicorn.error")
    
    loop = asyncio.get_running_loop()
    def _scan():
        #instream стримит файл в clamd, не требует доступа демона в пути
        try:
            with open(path, "rb") as f:
                res = _cd.instream(f)
        except (BrokenPipeError, clamd.ConnectionError, ConnectionError) as e:
            # Демон оборвал INSTREAM (например, файл превысил StreamMaxLength)
            # либо вообще недоступен — это отказ движка, а не «заражён».
            logger.error(f"ClamAV scan failed for {path.name}: {e}")
            raise ScannerUnavailableError(f"clamd недоступен: {e}") from e
        #res: {'stream': ('OK', None)} | {'stream': ('FOUND', 'Eicar-...')}
        status, desc = res.get("stream", ("ERROR", "no response"))
        return status == "OK", desc or ""
    return await loop.run_in_executor(None, _scan)