import yara
from typing import List, Optional

# Правила для детекта подозрительных инжекций и дропперов
YARA_RULES_SRC = """
rule Suspicious_Process_Injection {
    meta:
        description = "Detects WinAPI used for DLL Injection and Memory Patching"
    strings:
        $api1 = "VirtualAllocEx" ascii wide
        $api2 = "WriteProcessMemory" ascii wide
        $api3 = "CreateRemoteThread" ascii wide
        $api4 = "SetWindowsHookEx" ascii wide
    condition:
        2 of ($api*)
}

rule Suspicious_Script_Download {
    meta:
        description = "Detects PowerShell / Web Client download commands in scripts"
    strings:
        $p1 = "DownloadString" ascii wide nocase
        $p2 = "powershell -enc" ascii wide nocase
        $p3 = "IEX" ascii wide nocase
    condition:
        any of them
}
"""

class YaraScanner:
    def __init__(self):
        try:
            self.rules = yara.compile(source=YARA_RULES_SRC)
        except Exception as e:
            self.rules = None
            print(f"Warning: Failed to compile YARA rules: {e}")

    def scan_file(self, file_path: str) -> List[str]:
        if not self.rules:
            return []
        try:
            matches = self.rules.match(file_path)
            return [m.rule for m in matches]
        except Exception:
            return []