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

Остановить сервер и выполнить команду (заменить `userId` на нужный):
```bash
go run ./cmd/api/ --make-admin userId
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
3. Просмотр логов проверки файлов:
```bash
docker compose logs -f scanner
```
4. Назначение пользователя модератором(--make-moderator)/админом(--make-admin):
```bash
docker-compose exec api ./myapi --make-admin userId
```