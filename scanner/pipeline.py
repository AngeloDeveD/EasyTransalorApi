import asyncio, shutil, tempfile, uuid, os
from pathlib import Path

from magic_bytes import detect_type
from safe_extract import safe_extract, SecurityError
from bomb_guard import ZipBombError
from av_scan import clamav_scan, ScannerUnavailableError
from yara_scan import yara_scan, EXEC_EXT
from sandbox import submit_to_sandbox
from config import settings


class ScanResult:
    def __init__(self, task_id: str):
        self.task_id = task_id
        self.status = "pending"
        self.threats = []          # список dict: file/engine/name/desc/rules
        self.sandbox_tid = None
        self.review_reasons = []   # причины, по которым файл ушёл на ручную проверку


async def run_pipeline(archive_path: Path, original_name: str) -> ScanResult:
    result = ScanResult(task_id=str(uuid.uuid4()))

    # Каталог для распаковки должен существовать до mkdtemp
    settings.EXTRACT_ROOT.mkdir(parents=True, exist_ok=True)
    workdir = Path(tempfile.mkdtemp(prefix="av_", dir=settings.EXTRACT_ROOT))
    try:
        # Ограничение на размер самого архива
        size = archive_path.stat().st_size
        if size > settings.MAX_UPLOADED_BYTES:
            raise SecurityError(
                f"archive too large: {size} > {settings.MAX_UPLOADED_BYTES}"
            )

        # Определение типа по magic-байтам (не по расширению)
        with open(archive_path, "rb") as f:
            fmt = detect_type(f.read(16))
        if fmt not in ("zip", "rar", "7z"):
            raise SecurityError(f"unsupported archive type: {fmt}")

        # Распаковка с проверкой на архивную бомбу (блокирующая — в executor)
        loop = asyncio.get_running_loop()
        extracted = await loop.run_in_executor(
            None, safe_extract, fmt, archive_path, workdir
        )

        # ClamAV по каждому распакованному файлу
        for fpath in extracted:
            # ClamAV не сканирует файлы крупнее ~4 ГБ. Не отклоняем и не стримим
            # (иначе Broken pipe / ложный reject) — уводим на ручную проверку.
            fsize = fpath.stat().st_size
            if fsize > settings.CLAMAV_MAX_SCAN_BYTES:
                result.review_reasons.append(
                    f"{fpath.name}: {fsize} байт > лимита ClamAV "
                    f"({settings.CLAMAV_MAX_SCAN_BYTES}) — не просканирован автоматически"
                )
                continue

            # Отказ движка (сокет закрыт, демон недоступен) — это НЕ «заражён».
            # Тоже уводим на ручную проверку, а не в rejected_by_scanner.
            try:
                clean, desc = await clamav_scan(fpath)
            except ScannerUnavailableError as e:
                result.review_reasons.append(f"{fpath.name}: движок не ответил ({e})")
                continue

            if not clean:
                result.status = "infected"
                result.threats.append(
                    {"file": fpath.name, "engine": "clamav", "name": desc, "rules": []}
                )
                _quarantine(fpath)
                return result

        # Если хотя бы один файл не удалось просканировать автоматически
        # (>4 ГБ или отказ движка) — уводим весь архив на ручную проверку.
        # НЕ отклоняем: это не вердикт «заражён», а невозможность проверить.
        if result.review_reasons:
            result.status = "needs_review"
            return result

        # YARA + песочница только для исполняемых файлов
        execs = [p for p in extracted if p.suffix.lower() in EXEC_EXT]
        if not execs:
            result.status = "clean"
            return result

        infected = False
        for p in execs:
            hits = yara_scan(p)
            if hits:
                infected = True
                result.threats.append(
                    {
                        "file": p.name,
                        "engine": "yara",
                        "name": ", ".join(hits),
                        "rules": hits,
                    }
                )
                _quarantine(p)

        if infected:
            result.status = "infected"
            return result

        # Все исполняемые файлы чисты по сигнатурам.
        # Песочница — опциональный глубокий анализ поведения. Если она выключена
        # (нет CAPE/Cuckoo), статического анализа ClamAV + YARA достаточно — файл чист.
        if not settings.SANDBOX_ENABLED:
            result.status = "clean"
            return result

        # Песочница включена: отправляем на анализ.
        # Файлы не в карантине, поэтому execs[0] всё ещё на месте.
        result.status = "sandbox_running"
        result.sandbox_tid = await submit_to_sandbox(execs[0], original_name)
        return result

    finally:
        shutil.rmtree(workdir, ignore_errors=True)


def _quarantine(p: Path):
    settings.QUARANTINE_ROOT.mkdir(parents=True, exist_ok=True)
    q = settings.QUARANTINE_ROOT / f"{uuid.uuid4()}_{p.name}"
    shutil.move(str(p), str(q))
    os.chmod(q, 0o400)
