# Сканер архивов: FastAPI + ClamAV-клиент + YARA.
FROM python:3.12-slim

# unar нужен библиотеке rarfile для распаковки .rar (свободный, поддерживает RAR5,
# rarfile находит его автоматически); gcc/libmagic — на случай сборки yara-python.
RUN apt-get update && apt-get install -y --no-install-recommends \
        unar \
        gcc \
        libmagic1 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY . .

# Каталоги для распаковки и карантина (совпадают с дефолтами в config.py)
RUN mkdir -p /tmp/av_extract /var/quarantine /etc/yara/rules

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
