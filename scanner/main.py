import os
import clamd
import requests
from flask import Flask, request, jsonify

app = Flask(__name__)

MAIN_API_URL = "http://localhost:8080/api/internal/scan-result"
INTERNAL_KEY = "super_secret_cloud_key_998"

try:
    cd = clamd.ClamdNetworkSocket('localhost', 3310)
    cd.ping()
    print("ClamAV подключён")
except Exception as e:
    print("Не удалось подключиться к ClamAV. Программа должна быть установлена и запущена")
    cd = None

@app.route('/scan', methods=['POST'])
def scan_file():
    data = request.json
    trans_id = data.get('transId')
    file_path = data.get('filePath')

    if not os.path.exists(file_path):
        return send_result(trans_id, "rejected_by_scanner", "Файл не найден на диске"), 500
    
    if not cd:
        #Если не работает ClamAV, то файл не пропускается
        return send_result(trans_id, "rejected_by_scanner", "Антивирусник недоступен"), 500
    
    try:
        #Сканирование файла
        #ClamAV умеет заглядывать внутрь архивов
        scan_result = cd.scan_file(file_path)

        status, virus_name = list(scan_file.values())[0]

        if status == "OK":
            print(f"Файл {file_path} чист. ID: {trans_id}")
            return send_result(trans_id, "pending", "clean"), 200
        elif status == "FOUND":
            msg = f"Обнаружен вирус: {virus_name}"
            print(f"{msg}. ID: {trans_id}")
            return send_result(trans_id, "rejected_by_scanner", msg), 200
        else:
            return send_result(trans_id, "rejected_by_scanner", "Ошибка сканирования"), 200
    
    except Exception as e:
        return send_result(trans_id, "rejected_by_scanner", str(e)), 500



def send_result(trans_id, status, details):
    "Отправка результата сканирования (и ошибок)"
    payload = {
        "transId": trans_id,
        "status": status,
        "details": details
    }

    headers = {"X-Internal-Key": INTERNAL_KEY}
    try:
        requests.post(MAIN_API_URL, json=payload, headers=headers)
    except Exception as e:
        print(f"Не удалось отправить результат: {e}")

    return jsonify({"message": "Сканирование завершено"})

if __name__ == "__main__":
    app.run(port=5000)