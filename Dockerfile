# --- Этап 1: Сборка приложения ---
# Версия образа должна быть не ниже go-директивы в go.mod (сейчас go 1.26.3),
# иначе тулчейн откажется собирать модуль.
FROM golang:1.26-alpine AS builder

# Устанавливаем git (нужен для go mod) и gcc (нужен для SQLite)
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник. 
# CGO_ENABLED=1 нужен для SQLite. 
# Флаг -ldflags="-s -w" урезает дебаг-инфу, делая файл меньше.
ENV CGO_ENABLED=1 GOOS=linux
RUN go build -ldflags="-s -w" -o myapi ./cmd/api/

# --- Этап 2: Финальный легкий образ ---
FROM alpine:latest

# Устанавливаем ca-certificates (для HTTPS) и снова gcc (для запуска SQLite)
RUN apk add --no-cache ca-certificates gcc musl-dev

WORKDIR /app

# Копируем собранный бинарник из первого этапа
COPY --from=builder /app/myapi .
# Копируем папку для загрузок (создастся, если нет)
RUN mkdir -p uploads

# Открываем порт
EXPOSE 8080

# Команда запуска
CMD ["./myapi"]