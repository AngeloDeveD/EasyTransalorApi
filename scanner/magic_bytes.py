import os

def is_pe_header(file_path: str) -> bool:
    """Проверяет сигнатуру MZ (0x4D 0x5A) — исполняемые файлы Windows (DLL/EXE)."""
    try:
        with open(file_path, "rb") as f:
            header = f.read(2)
            return header == b"MZ"
    except Exception:
        return False

def is_elf_or_script(file_path: str) -> bool:
    """Проверяет сигнатуры Linux ELF (\x7fELF) или шебанги скриптов (#!)."""
    try:
        with open(file_path, "rb") as f:
            header = f.read(4)
            return header.startswith(b"\x7fELF") or header.startswith(b"#!")
    except Exception:
        return False

def is_executable_by_content(file_path: str) -> bool:
    """Проверяет, является ли файл исполняемым независимо от его расширения."""
    return is_pe_header(file_path) or is_elf_or_script(file_path)