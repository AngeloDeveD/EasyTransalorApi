import yara
from pathlib import Path
from functools import lru_cache

from config import settings

EXEC_EXT = {".exe", ".dll", ".sys", ".scr"}

@lru_cache(maxsize=1)
def _compiled_rulse():
    filepaths = {}
    for i, p in enumerate(sorted(Path(settings.YARA_RULES_DIR).glob("*.yar"))):
        filepaths[f"ns{i}"] = str(p)
    return yara.compile(filepaths=filepaths) if filepaths else None

def yara_scan(path: Path) -> list[str]:
    rules = _compiled_rulse()
    if rules is None:
        return []
    matches = rules.match(filepath=str(path))
    return [m.rule for m in matches]