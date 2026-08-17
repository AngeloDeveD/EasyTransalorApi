import os
import hashlib
from typing import Dict, Any, List
from config import Config
from magic_bytes import is_executable_by_content
from safe_extract import SafeExtractor, SecurityException
from av_scan import ClamAVScanner
from yara_scan import YaraScanner

yara_scanner = YaraScanner()

def calculate_file_sha256(file_path: str, chunk_size: int = 1024 * 1024) -> str:
    """Потоковое вычисление SHA-256 (чанками по 1 МБ) без загрузки файла в RAM."""
    hasher = hashlib.sha256()
    with open(file_path, "rb") as f:
        while chunk := f.read(chunk_size):
            hasher.update(chunk)
    return hasher.hexdigest()

def get_readalbe_size(path):
    #Получение размера в байтах
    bytes_size = os.path.getsize(path)

    for unit in ['b', 'Kb', 'Mb', 'Gb']:
        if bytes_size < 1024.0:
            return f"{bytes_size:.2f} {unit}", bytes_size
        bytes_size /= 1024.0

    return f"{bytes_size:.2f} Tb", bytes_size

class ScanPipeline:
    @classmethod
    def scan_archive(cls, archive_path: str) -> Dict[str, Any]:
        result = {
            "status": "approved",
            "is_safe": True,
            "threats": [],
            "files": [],
            "scanned_executables_count": 0,
            "error": None
        }

        try:
            with SafeExtractor.extract_zip(archive_path) as extracted_dir:
                high_risk_files: List[str] = []
                files_manifest: List[Dict[str, Any]] = []
                
                # Триаж и категоризация файлов
                for root, _, files in os.walk(extracted_dir):
                    for file in files:
                        file_path = os.path.join(root, file)
                        ext = os.path.splitext(file)[1].lower()
                        formatted_size, size = get_readalbe_size(file_path)

                        #Формирование относительного пути: "папка/файл"
                        rel_path = os.path.relpath(file_path, extracted_dir).replace("\\", "/")

                        #Вычисление SHA-256
                        file_hash = calculate_file_sha256(file_path)

                        files_manifest.append({
                            "fileName": rel_path,
                            "hash": file_hash,
                            "size": formatted_size
                        })

                        # Проверяем, исполняемый ли файл по расширению или заголовку MZ
                        is_exec = (ext in Config.EXECUTABLE_EXTENSIONS) or is_executable_by_content(file_path)

                        if is_exec:
                            # 1. Защита от Binary Bloating (> 100 MB)
                            if size > Config.MAX_EXECUTABLE_SIZE:
                                return {
                                    "status": "rejected",
                                    "is_safe": False,
                                    "threats": [f"EXECUTABLE_TOO_LARGE: '{file}' exceeds 100MB limit"],
                                    "files": [],
                                    "error": "Исполняемый файл превышает допустимый лимит безопасности."
                                }
                            high_risk_files.append(file_path)

                result["scanned_executables_count"] = len(high_risk_files)
                result["files"] = files_manifest

                # Сканируем только опасные файлы через ClamAV и YARA
                for file_path in high_risk_files:
                    if Config.CLAMAV_ENABLED:

                        # А. Проверка ClamAV
                        is_clean, threat = ClamAVScanner.scan_file(file_path)
                        if not is_clean:
                            result["status"] = "rejected"
                            result["is_safe"] = False
                            result["threats"].append(f"ClamAV: {threat} in {os.path.basename(file_path)}")
                            return result
                    else:
                        pass

                    # Б. Проверка YARA
                    yara_matches = yara_scanner.scan_file(file_path)
                    if yara_matches:
                        result["status"] = "rejected"
                        result["is_safe"] = False
                        for match in yara_matches:
                            result["threats"].append(f"YARA: {match} in {os.path.basename(file_path)}")

                if not result["is_safe"]:
                    return result

        except SecurityException as se:
            return {
                "status": "rejected",
                "is_safe": False,
                "threats": [f"SECURITY_POLICY_VIOLATION: {str(se)}"],
                "files": [],
                "error": str(se)
            }
        except Exception as e:
            return {
                "status": "pending",
                "is_safe": False,
                "threats": [],
                "files": [],
                "error": f"Ошибка сканирования: {str(e)}"
            }

        return result
