import os
import sys
import time
import zipfile
import json
import unittest
from typing import List, Dict, Any

# Импортируем наш пайплайн и сканер ClamAV
try:
    from pipeline import ScanPipeline
    from av_scan import ClamAVScanner
    from config import Config
except ImportError as exc:
    # Если запуск не из папки scanner
    sys.path.append(os.path.dirname(os.path.abspath(__file__)))
    try:
        from pipeline import ScanPipeline
        from av_scan import ClamAVScanner
        from config import Config
    except ImportError as nested_exc:
        raise unittest.SkipTest(f"scanner dependencies are not installed: {nested_exc}") from exc


# Цвета для красивого вывода в терминал
class Colors:
    GREEN = "\033[92m"
    RED = "\033[91m"
    YELLOW = "\033[93m"
    CYAN = "\033[96m"
    BOLD = "\033[1m"
    RESET = "\033[0m"


def progress_bar(current: int, total: int, prefix: str = "", bar_length: int = 35):
    """Отображает интерактивный ползунок прогресса в консоли."""
    percent = float(current) * 100 / total
    filled_length = int(bar_length * current // total)
    bar = "█" * filled_length + "░" * (bar_length - filled_length)
    sys.stdout.write(f"\r{Colors.CYAN}{prefix}{Colors.RESET} |{Colors.YELLOW}{bar}{Colors.RESET}| {percent:.1f}%")
    sys.stdout.flush()
    if current == total:
        sys.stdout.write("\n")


def setup_test_environment(test_dir: str):
    """Создает тестовую папку и генерирует файлы, если они отсутствуют."""
    os.makedirs(test_dir, exist_ok=True)

    eicar_string = b"X5O!P%@AP[4\\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*"
    virus_path = os.path.join(test_dir, "virus.com")
    virus_zip_path = os.path.join(test_dir, "virus_archive.zip")
    clean_zip_path = os.path.join(test_dir, "clean.zip")

    # 1. Создаем тестовый файл EICAR virus.com
    if not os.path.exists(virus_path):
        with open(virus_path, "wb") as f:
            f.write(eicar_string)

    # 2. Создаем архив с вирусом внутри (для проверки распаковки)
    if not os.path.exists(virus_zip_path):
        with zipfile.ZipFile(virus_zip_path, "w") as zf:
            zf.writestr("infected_game_mod.dll", eicar_string)
            zf.writestr("readme.txt", "This is an infected translation mod.")

    # 3. Создаем чистый тестовый архив
    if not os.path.exists(clean_zip_path):
        with zipfile.ZipFile(clean_zip_path, "w") as zf:
            zf.writestr("translation/dialogues.json", json.dumps({"greeting": "Привет, мир!"}))
            zf.writestr("textures/ui_font.png", b"\x89PNG\r\n\x1a\n" + b"\x00" * 1024)
            zf.writestr("config.ini", "[Settings]\nLanguage=ru\n")

    return {
        "virus_com": virus_path,
        "virus_zip": virus_zip_path,
        "clean_zip": clean_zip_path
    }


def run_tests():
    base_dir = os.path.dirname(os.path.abspath(__file__))
    test_files_dir = os.path.join(base_dir, "test_files")
    
    print(f"\n{Colors.BOLD}=== Инициализация среды тестирования сканера ==={Colors.RESET}")
    files = setup_test_environment(test_files_dir)
    time.sleep(0.5)

    test_cases = [
        {
            "name": "Прямое сканирование тестового вируса (virus.com / EICAR)",
            "file": files["virus_com"],
            "type": "single_file",
            "expected_safe": False,
            "expected_status": "infected"
        },
        {
            "name": "Сканирование архива с скрытым вирусом (virus_archive.zip)",
            "file": files["virus_zip"],
            "type": "archive",
            "expected_safe": False,
            "expected_status": "infected"
        },
        {
            "name": "Сканирование чистого игрового архива (clean.zip)",
            "file": files["clean_zip"],
            "type": "archive",
            "expected_safe": True,
            "expected_status": "clean"
        }
    ]

    total_steps = len(test_cases)
    results = []

    print(f"\n{Colors.BOLD}=== Запуск тестирования пайплайна безопасности ==={Colors.RESET}\n")

    for idx, test in enumerate(test_cases, 1):
        print(f"{Colors.BOLD}[Тест {idx}/{total_steps}]: {test['name']}{Colors.RESET}")
        
        # Симуляция анимации ползунка при отправке и обработке
        for p in range(1, 11):
            time.sleep(0.04)
            progress_bar(p, 10, prefix="Обработка и сканирование:")

        verdict = {}
        # Запуск соответствующего обработчика
        if test["type"] == "single_file":
            is_clean, threat = ClamAVScanner.scan_file(test["file"])
            verdict = {
                "status": "clean" if is_clean else "infected",
                "is_safe": is_clean,
                "threats": [threat] if threat else []
            }
        else:
            verdict = ScanPipeline.scan_archive(test["file"])

        # Проверка соответствия ожиданиям
        passed = (verdict["is_safe"] == test["expected_safe"]) and (verdict["status"] == test["expected_status"])
        
        results.append({
            "name": test["name"],
            "passed": passed,
            "verdict": verdict
        })

        if passed:
            print(f"Статус проверки: {Colors.GREEN}✓ УСПЕШНО ПРОЙДЕН{Colors.RESET}")
        else:
            print(f"Статус проверки: {Colors.RED}✗ ОШИБКА{Colors.RESET}")

        print(f"Вердикт сканера: status='{verdict.get('status')}', threats={verdict.get('threats')}\n")

    # Вывод итоговой таблицы
    print(f"{Colors.BOLD}================ ИТОГИ ТЕСТОВ ================{Colors.RESET}")
    all_passed = True
    for r in results:
        status_text = f"{Colors.GREEN}[PASS]{Colors.RESET}" if r["passed"] else f"{Colors.RED}[FAIL]{Colors.RESET}"
        if not r["passed"]:
            all_passed = False
        print(f"{status_text} | {r['name']}")

    print("==============================================")
    if all_passed:
        print(f"{Colors.GREEN}{Colors.BOLD}Все тесты безопасности успешно пройдены!{Colors.RESET}\n")
    else:
        print(f"{Colors.RED}{Colors.BOLD}Некоторые тесты провалены. Проверьте настройки ClamAV/YARA.{Colors.RESET}\n")
        sys.exit(1)


if __name__ == "__main__":
    run_tests()