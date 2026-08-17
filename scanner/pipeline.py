import os
from typing import Dict, Any, List
from config import Config
from magic_bytes import is_executable_by_content
from safe_extract import SafeExtractor, SecurityException
from av_scan import ClamAVScanner
from yara_scan import YaraScanner

yara_scanner = YaraScanner()

class ScanPipeline:
    @classmethod
    def scan_archive(cls, archive_path: str) -> Dict[str, Any]:
        result = {
            "status": "approved",
            "is_safe": True,
            "threats": [],
            "scanned_executables_count": 0,
            "error": None
        }

        try:
            with SafeExtractor.extract_zip(archive_path) as extracted_dir:
                high_risk_files: List[str] = []

                # Триаж и категоризация файлов
                for root, _, files in os.walk(extracted_dir):
                    for file in files:
                        file_path = os.path.join(root, file)
                        ext = os.path.splitext(file)[1].lower()
                        size = os.path.getsize(file_path)

                        # Проверяем, исполняемый ли файл по расширению или заголовку MZ
                        is_exec = (ext in Config.EXECUTABLE_EXTENSIONS) or is_executable_by_content(file_path)

                        if is_exec:
                            # 1. Защита от Binary Bloating (> 100 MB)
                            if size > Config.MAX_EXECUTABLE_SIZE:
                                return {
                                    "status": "rejected",
                                    "is_safe": False,
                                    "threats": [f"EXECUTABLE_TOO_LARGE: '{file}' exceeds 100MB limit"],
                                    "error": "Исполняемый файл превышает допустимый лимит безопасности."
                                }
                            high_risk_files.append(file_path)

                result["scanned_executables_count"] = len(high_risk_files)

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
                "error": str(se)
            }
        except Exception as e:
            return {
                "status": "pending",
                "is_safe": False,
                "threats": [],
                "error": f"Ошибка сканирования: {str(e)}"
            }

        return result