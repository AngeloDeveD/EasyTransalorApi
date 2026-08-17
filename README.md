# EasyTranslator API

Бэкенд-сервис для платформы игровых переводов и модификаций с асинхронным конвейером безопасности (Static Analysis, Bomb Guard, YARA, ClamAV Antivirus).

---

## Сетевая карта сервисов (Порты)

| Сервис | Назначение | Порт внутри Docker | Внешний доступ (Хост) |
| :--- | :--- | :--- | :--- |
| **Go API** | Основной бэкенд и модерация | `:8080` | `http://localhost:8080` |
| **Scanner** | Сервис анализа файлов (FastAPI) | `:8000` | `http://localhost:8000` *(в Dev-режиме)* |
| **PostgreSQL** | База данных | `:5432` | `localhost:5432` |
| **ClamAV** | Антивирусный демон | `:3310` | `clamav:3310` *(только внутренняя сеть)* |

---

## Быстрый запуск (Скрипты управления)

В проект встроены удобные скрипты с интерактивным меню и быстрыми командами:

### Windows:
* **Интерактивное меню:** запустите `run.bat`.
* **Быстрые команды:**
  ```cmd
  run.bat dev    # Быстрый запуск DEV (без ClamAV, ~120 MB RAM)
  run.bat full   # Полный боевой запуск PROD (с ClamAV, ~1 GB RAM)
  run.bat stop   # Остановка контейнеров
  run.bat reset  # Полный сброс базы данных
  run.bat logs   # Просмотр логов в реальном времени
  run.bat mod/moderator <userId> # Назначить роль модератора по userId
  run.bat admin <userId> # Назначить админку по userId
  ```

### Linux / macOS:
Перед первым использованием выдайте права на исполнение: `chmod +x run.sh`
* **Интерактивное меню:** `./run.sh`
* **Быстрые команды:**
  ```bash
  ./run.sh dev    # Быстрый запуск DEV (без ClamAV)
  ./run.sh full   # Полный запуск PROD (с ClamAV)
  ./run.sh stop   # Остановка
  ./run.sh reset  # Полный сброс базы данных
  ./run.sh logs   # Просмотр логов
  ./run.sh mod/moderator <userId> # Назначить роль модератора по userId
  ./run.sh admin <userId> # Назначить админку по userId
  ```

---

## Ручной запуск через Docker Compose

### 1. Режим быстрой разработки (DEV — без ClamAV)
> Старт за 1.5 секунды, потребляет всего ~120 МБ RAM. Антивирус отключен, но работают проверки метаданных, защита от Zip-бомб и YARA.

```bash
docker compose -f docker-compose.yml -f docker-compose.no-av.yml up -d --build
```

### 2. Полный запуск (PROD — с проверками ClamAV)
> Полная проверка сигнатур ClamAV (3.6+ млн баз). Требует ~1–1.2 GB RAM.

```bash
docker compose up -d --build
```

### 3. Остановка сервисов
```bash
docker compose down
```


---

## Swagger UI

После запуска Go API интерактивная документация доступна по адресу:

```text
http://localhost:8080/swagger/index.html
```

OpenAPI JSON доступен по адресу:

```text
http://localhost:8080/swagger/doc.json
```

Для запросов с авторизацией нажмите **Authorize** и укажите JWT в формате `Bearer <token>`.

---

## Политики безопасности и лимиты файлов

* **Максимальный размер архива:** 5 GB.
* **Максимальный распакованный объем:** 10 GB.
* **Лимит на исполняемые файлы/скрипты (`.dll`, `.exe`, `.lua`, `.bat`):** строго до 100 MB *(защита от Binary Bloating)*.
* **Защита от атак:** встроенная валидация Path Traversal (Zip Slip), отсечение POSIX Symlinks и проверка коэффициента сжатия (Zip Bomb Guard).
* **Поддерживаемые архивы:** `.zip`, `.7z`, `.rar`.

---

## Тестирование сканера безопасности

Для проверки работы конвейера и детекта тестовых сигнатур (EICAR):

```bash
python scanner/test_scanner.py
```
*(Скрипт автоматически создаст тестовые файлы и выведет статус проверки с интерактивным прогресс-баром в консоли).*

---

## Команды администрирования (CLI)

Управление правами пользователей через консоль контейнера:

* **Назначить администратора:**
  ```bash
  docker compose exec api ./myapi --make-admin <userID>
  ```
* **Назначить модератора:**
  ```bash
  docker compose exec api ./myapi --make-moderator <userID>
  ```

---

## Управление базой данных и очистка

**Просмотр логов антивирусного сканера:**
```bash
docker compose logs -f scanner
```
**Просмотр логов основного API:**
```bash
docker compose logs -f api
```
**Полная очистка базы данных PostgreSQL:**
```bash
docker compose down -v
```

---

## Миграции БД

При старте API применяет версионированные миграции из `internal/migrations` и записывает применённые версии в таблицу `schema_migrations`.

Правила изменения схемы:

1. Добавляйте новую миграцию с новым ID в `internal/migrations`.
2. Не меняйте уже применённые миграции задним числом.
3. Проверяйте изменения через `go test ./...` или Docker-прогон из CI.


---

## Переменные окружения (`.env`)

Создайте файл `.env` в корне проекта при необходимости переопределить стандартные значения:

```ini
# База данных PostgreSQL
POSTGRES_USER=postgres
POSTGRES_PASSWORD=secret
POSTGRES_DB=translations_db

# Безопасность и авторизация
JWT_SECRET=super_duper_secret_prod_key_998
INTERNAL_KEY=super_secret_cloud_key_998
ENCRYPT_KEY=12345678901234567890123456789012

# Конфигурация ClamAV
CLAMD_HOST=clamav
CLAMD_PORT=3310

# Настройки Webhook
API_WEBHOOK_URL=http://api:8080/api/internal/scan-result
```
---

## Rate Limiting

API ограничивает частоту запросов in-memory внутри Go-приложения:

```ini
RATE_LIMIT_ENABLED=true
RATE_LIMIT_GLOBAL_REQUESTS=120
RATE_LIMIT_GLOBAL_WINDOW=1m
RATE_LIMIT_AUTH_REQUESTS=10
RATE_LIMIT_AUTH_WINDOW=1m
RATE_LIMIT_WRITE_REQUESTS=1
RATE_LIMIT_WRITE_WINDOW=10s
```

* `GLOBAL` - общий лимит на IP для всего API.
* `AUTH` - отдельный лимит на `/api/auth/login` и `/api/auth/register`, чтобы снижать риск брутфорса.
* `WRITE` - cooldown на повторные write-запросы одного пользователя к одному route.
