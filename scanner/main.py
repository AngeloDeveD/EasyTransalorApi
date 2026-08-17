import os
import logging
import httpx
from fastapi import FastAPI, BackgroundTasks, Header, HTTPException, status
from pydantic import BaseModel
from typing import Optional, List
from pipeline import ScanPipeline
from config import Config

# Настройка логирования
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("scanner")

app = FastAPI(title="Game Translation & Mod Security Scanner")

config = Config()

# Модель входящего запроса от Go
class ScanRequest(BaseModel):
    transId: int
    filePath: str


# Фоновая задача: сканирование и вызов Webhook в Go
async def process_scan_and_notify(trans_id: int, file_path: str):
    logger.info(f"Начало проверки файла для transId={trans_id}, путь: {file_path}")
    
    if not os.path.exists(file_path):
        logger.error(f"Файл не найден на диске: {file_path}")
        payload = {
            "transId": trans_id,
            "status": "error",
            "isSafe": False,
            "threats": [f"File not found on scanner disk: {file_path}"],
            "details": f"File not found on scanner disk: {file_path}",
            "files": [],
            "error": "Файл не найден"
        }
    else:
        # Запуск нашего пайплайна (SafeExtract, BombGuard, MagicBytes, ClamAV, YARA)
        try:
            verdict = ScanPipeline.scan_archive(file_path)
            payload = {
                "transId": trans_id,
                "status": verdict.get("status", "error"),
                "isSafe": verdict.get("is_safe", False),
                "threats": verdict.get("threats", []),
                "details": "; ".join(verdict.get("threats", [])) or verdict.get("error") or "",
                "files": verdict.get("files", []),
                "error": verdict.get("error")
            }
        except Exception as e:
            logger.exception(f"Критическая ошибка при сканировании {file_path}: {e}")
            payload = {
                "transId": trans_id,
                "status": "error",
                "isSafe": False,
                "threats": [],
                "details": str(e),
                "files": [],
                "error": str(e)
            }

    logger.info(f"Сканирование transId={trans_id} завершено. Вердикт: {payload['status']} | Ошибка: {payload.get('error')} | Угрозы: {payload.get('threats')}")

    # Отправляем вердикт обратно в Go API
    headers = {
        "Content-Type": "application/json",
        "X-Internal-Key": config.INTERNAL_KEY
    }
    
    async with httpx.AsyncClient(timeout=30.0) as client:
        try:
            resp = await client.post(config.API_WEBHOOK_URL, json=payload, headers=headers)
            if resp.status_code == 200:
                logger.info(f"Go API успешно принял результат для transId={trans_id}")
            else:
                logger.error(f"Go API ответил ошибкой {resp.status_code}: {resp.text}")
        except Exception as e:
            logger.error(f"Не удалось доставить результат в Go API ({config.API_WEBHOOK_URL}): {e}")


@app.post("/scan")
async def scan_endpoint(
    req: ScanRequest, 
    background_tasks: BackgroundTasks,
    x_internal_key: Optional[str] = Header(None, alias="X-Internal-Key")
):
    # Проверка безопасности между контейнерами
    if config.INTERNAL_KEY and x_internal_key != config.INTERNAL_KEY:
        raise HTTPException(
            status_code=status.HTTP_403_FORBIDDEN, 
            detail="Invalid X-Internal-Key"
        )

    # Ставим сканирование в фоновую задачу, чтобы не блокировать Go API
    background_tasks.add_task(process_scan_and_notify, req.transId, req.filePath)
    
    # Сразу возвращаем 200 OK
    return {"status": "queued", "message": f"Scan task for transId={req.transId} started"}