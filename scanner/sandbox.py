import httpx, asyncio
from pathlib import Path

from config import settings

async def submit_to_sandbox(path: Path, filename: str) -> int:
    """Возвращает task_id в песочнице"""
    async with httpx.AsyncClient(timeout=30) as client:
        with open(path, "rb") as f:
            resp = await client.post(
                f"{settings.SANDBOX_API}/tasks/create/file",
                headers={"Authorization": f"Token {settings.SANDBOX_TOKEN}"},
                files = {"file": (filename, f, "application/octet-stream")},
            )
        resp.raise_for_status()
        return resp.json()["task_id"]

async def fetch_report(task_id: int) -> dict:
    async with httpx.AsyncClient(timeout=30) as client:
        r = await client.get(
            f"{settings.SANDBOX_API}/tasks/report/{task_id}",
            headers={"Authorization": f"Token {settings.SANDBOX_TOKEN}"},
            params={"format": "json"},
        )
        r.raise_for_status()
        return r.json()

def is_malicious(report: dict) -> bool:
    """Эвристика по отчёту CAPE/Cuckoo"""
    if report.get("malscore", 0) >= 7.0:
        return True
    if any(s.get("severity") == 5 for s in report.get("signatures", [])):
        return True
    return False