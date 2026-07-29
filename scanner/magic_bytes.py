MAGIC = {
    "zip": (b"PK\x03\x04", b"PK\x05\x06", b"PK\x07\x08"),
    "rar": (b"Rar!\x1a\x07\x00", b"Rar!\x1a\x07\x01\x00"),
    "7z": (b"7z\xbc\xaf\x27\x1c",),
}

def detect_type(head: bytes) -> str | None:
    for fmt, sigs in MAGIC.items():
        if any(head.startswith(s) for s in sigs):
            return fmt
    return None