import clamd, asyncio
from pathlib import Path

from config import settings

_cd = clamd.ClamdUnixSocket(settings.CLAMD_SOCKET)

async def clamav_scan(path: Path) -> tuple[bool, str]:
    """Возвращает (clean, description)."""
    loop = asyncio.get_running_loop()
    def _scan():
        #instream стримит файл в clamd, не требует доступа демона в пути
        with open(path, "rb") as f:
            res = _cd.instream(f)
        #res: {'stream': ('OK', None)} | {'stream': ('FOUND', 'Eicar-...')}
        status, desc = res.get("stream", ("ERROR", "no response"))
        return status == "OK", desc or ""
    return await loop.run_in_executor(None, _scan)