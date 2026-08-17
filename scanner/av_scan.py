import socket
import struct
from typing import Tuple, Optional
from config import Config

class ClamAVScanner:
    @staticmethod
    def _send_instream(sock: socket.socket, file_path: str) -> None:
        """Потоковая отправка файла в clamd через команду INSTREAM."""
        sock.sendall(b"zINSTREAM\0")
        with open(file_path, "rb") as f:
            while chunk := f.read(Config.CHUNK_SIZE):
                size = struct.pack("!I", len(chunk))
                sock.sendall(size + chunk)
        # Завершающий чанк размером 0
        sock.sendall(struct.pack("!I", 0))

    @classmethod
    def scan_file(cls, file_path: str) -> Tuple[bool, Optional[str]]:
        """
        Возвращает (is_clean, threat_name).
        is_clean = True (чисто), False (найдена угроза).
        """
        try:
            with socket.create_connection(
                (Config.CLAMAV_HOST, Config.CLAMAV_PORT), 
                timeout=Config.CLAMAV_TIMEOUT
            ) as sock:
                cls._send_instream(sock, file_path)
                response = sock.recv(4096).decode("utf-8", errors="ignore").strip()

                if "FOUND" in response:
                    # Пример ответа: 'stream: Win.Test.EICAR_HDB-1 FOUND'
                    threat_name = response.split("FOUND")[0].replace("stream:", "").strip()
                    return False, threat_name
                elif "OK" in response:
                    return True, None
                else:
                    raise RuntimeError(f"ClamAV error response: {response}")
        except Exception as e:
            # При недоступности антивируса — безопасный отказ
            raise RuntimeError(f"Ошибка соединения с ClamAV: {e}")