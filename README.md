# Инструкция по запуску

### Локальный запуск (SQLite)

1. Установить зависимости:
```bash
go mod tidy
```
2. Запустить сервер:
```bash
go run ./cmd/api/
```

### Назначение администратора

Остановить сервер и выполнить команду (заменить `nickname` на нужный):
```bash
go run ./cmd/api/ --make-admin nickname
```

### Запуск через Docker (PostgreSQL)

1. Собрать и запустить контейнеры:
```bash
docker-compose up -d --build
```
2. Остановка:
```bash
docker-compose down
```

### Переменные окружения

* `APP_PORT` - порт (по умолчанию 8080)
* `DB_TYPE` - `sqlite` или `postgres` (по умолчанию sqlite)
* `DB_NAME` - имя БД (по умолчанию app.db)
* `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS` - настройки для PostgreSQL
* `JWT_SECRET` - секретный ключ для JWT

### Тесты

```bash
go test ./... -v
```
