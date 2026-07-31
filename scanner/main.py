import logging

import httpx
from fastapi import FastAPI, BackgroundTasks, Header, HTTPException
from pathlib import Path
from pydantic import BaseModel

from pipeline import run_pipeline, SecurityError, ZipBombError
from sandbox import fetch_report, is_malicious
from config import settings

app = FastAPI()
logger = logging.getLogger("uvicorn.error")

# Сопоставление task_id песочницы -> transId основного сервиса.
# Заполняется при отправке в песочницу, используется в callback'е.
_sandbox_map: dict[int, int] = {}


class ScanTask(BaseModel):
    transId: int
    filePath: str


def _require_internal_key(x_internal_key: str | None) -> None:
    if not x_internal_key or x_internal_key != settings.INTERNAL_KEY:
        raise HTTPException(status_code=401, detail="unauthorized")


async def _report_to_go(trans_id: int, status: str, details: str) -> None:
    """Отправляет результат сканирования обратно в Go-сервис."""
    payload = {"transId": trans_id, "status": status, "details": details}
    headers = {"X-Internal-Key": settings.INTERNAL_KEY}
    async with httpx.AsyncClient(timeout=30) as client:
        try:
            await client.post(settings.MAIN_API_URL, json=payload, headers=headers)
            logger.info(f"Результат по transId {trans_id} отправлен в Go: {status}")
        except Exception as e:
            logger.error(f"Не удалось отправить результат в Go: {e}")


async def process_and_callback(task: ScanTask):
    """Запускает пайплайн и отправляет результат в Go."""
    result_status = "pending"
    result_details = "clean"

    try:
        archive_path = Path(task.filePath)
        if not archive_path.exists():
            result_status = "rejected_by_scanner"
            result_details = "Файл не найден на диске"
        else:
            scan_result = await run_pipeline(archive_path, archive_path.name)

            if scan_result.status == "clean":
                result_status = "pending"
                result_details = "clean"
            elif scan_result.status == "sandbox_running":
                # Файл ушёл в песочницу — запоминаем связь для callback'а
                result_status = "pending_sandbox"
                result_details = f"Анализ в песочнице. Task: {scan_result.sandbox_tid}"
                if scan_result.sandbox_tid is not None:
                    _sandbox_map[scan_result.sandbox_tid] = task.transId
            elif scan_result.status == "needs_review":
                # Файл не удалось проверить автоматически (>4 ГБ или отказ движка).
                # Это НЕ «заражён»: уводим на ручную модерацию (Go-статус pending —
                # он же попадает в очередь модерации), а не в rejected_by_scanner.
                result_status = "pending"
                reasons = "; ".join(scan_result.review_reasons) or "причина не указана"
                result_details = (
                    "ТРЕБУЕТСЯ РУЧНАЯ ПРОВЕРКА: автоматическое сканирование "
                    f"выполнено не полностью. {reasons}"
                )
            elif scan_result.status == "infected":
                result_status = "rejected_by_scanner"
                # Сборка всех угроз в читаемую строку
                threats = []
                for t in scan_result.threats:
                    name = t.get("name", "Unknown")
                    engine = t.get("engine", "")
                    fname = t.get("file", "")
                    threats.append(f"[{engine}] {fname}: {name}")
                result_details = "Обнаружены угрозы: " + "; ".join(threats)

    except ZipBombError as e:
        result_status = "rejected_by_scanner"
        result_details = f"Архивная бомба: {e}"
    except SecurityError as e:
        result_status = "rejected_by_scanner"
        result_details = f"Небезопасный архив: {e}"
    except Exception as e:
        logger.error(f"Непредвиденная ошибка сканера: {e}")
        result_status = "rejected_by_scanner"
        result_details = f"Внутренняя ошибка сканера: {e}"

    await _report_to_go(task.transId, result_status, result_details)


@app.post("/scan")
async def scan_archive(
    task: ScanTask,
    background_tasks: BackgroundTasks,
    x_internal_key: str | None = Header(default=None),
):
    """Go вызывает этот эндпоинт, передавая путь к файлу."""
    _require_internal_key(x_internal_key)
    # Пайплайн уходит в фон, сразу отвечаем 202-подобным ответом
    background_tasks.add_task(process_and_callback, task)
    return {"message": "Задача принята в работу", "transId": task.transId}


@app.post("/internal/sandbox-callback")
async def sandbox_callback(
    payload: dict,
    background_tasks: BackgroundTasks,
    x_internal_key: str | None = Header(default=None),
):
    """Эндпоинт, который дёргает CAPE/песочница по завершении анализа."""
    _require_internal_key(x_internal_key)

    task_id = payload.get("task_id")
    # Если в callback'е нет готового отчёта — тянем его сами
    report = payload.get("report")
    if not report and task_id is not None:
        try:
            report = await fetch_report(int(task_id))
        except Exception as e:
            logger.error(f"Не удалось получить отчёт песочницы {task_id}: {e}")
            report = {}
    report = report or {}

    malicious = is_malicious(report)
    logger.info(f"Песочница завершила анализ. Task: {task_id}. Вредонос: {malicious}")

    # Замыкаем цикл обратно на Go по сохранённому transId
    trans_id = _sandbox_map.pop(int(task_id), None) if task_id is not None else None
    if trans_id is not None:
        if malicious:
            status, details = "rejected_by_scanner", "Песочница: обнаружено вредоносное поведение"
        else:
            status, details = "pending", "Песочница: поведение чистое"
        background_tasks.add_task(_report_to_go, trans_id, status, details)
    else:
        logger.warning(f"Нет transId для task_id песочницы {task_id}")

    return {"ok": True, "malicious": malicious}
