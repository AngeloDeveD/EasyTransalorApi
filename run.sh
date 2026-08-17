#!/bin/bash

GREEN="\033[0;32m"
BLUE="\033[0;34m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
RESET="\033[0m"

if docker compose version &>/dev/null; then
    DOCKER_COMPOSE="docker compose"
elif command -v docker-compose &>/dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    echo -e "${RED}[ERROR] Docker Compose не установлен!${RESET}"
    exit 1
fi

function set_admin() {
    local uid=$1
    if [ -z "$uid" ]; then
        read -p "Введите ID пользователя для назначения АДМИНОМ: " uid
    fi
    if [ -n "$uid" ]; then
        $DOCKER_COMPOSE exec api ./myapi --make-admin "$uid"
    fi
}

function set_mod() {
    local uid=$1
    if [ -z "$uid" ]; then
        read -p "Введите ID пользователя для назначения МОДЕРАТОРОМ: " uid
    fi
    if [ -n "$uid" ]; then
        $DOCKER_COMPOSE exec api ./myapi --make-moderator "$uid"
    fi
}

case "$1" in
    dev)
        $DOCKER_COMPOSE -f docker-compose.yml -f docker-compose.no-av.yml up --build
        exit 0
        ;;
    full|prod)
        $DOCKER_COMPOSE up --build
        exit 0
        ;;
    stop)
        $DOCKER_COMPOSE down
        exit 0
        ;;
    reset)
        $DOCKER_COMPOSE down -v
        exit 0
        ;;
    logs)
        $DOCKER_COMPOSE logs -f
        exit 0
        ;;
    admin)
        set_admin "$2"
        exit 0
        ;;
    mod|moderator)
        set_mod "$2"
        exit 0
        ;;
esac

clear
echo -e "${BLUE}======================================================${RESET}"
echo -e "           EasyTranslator Server Management           "
echo -e "${BLUE}======================================================${RESET}"
echo "  [1] Быстрый запуск DEV (Без ClamAV, ~120 MB RAM)"
echo "  [2] Полный запуск PROD (С ClamAV, ~1 GB RAM)"
echo "  [3] Остановить все контейнеры"
echo "  [4] Полный сброс базы данных (Wipe Database)"
echo "  [5] Посмотреть логи в реальном времени"
echo "  ----------------------------------------------------"
echo "  [6] Назначить АДМИНИСТРАТОРА (--make-admin)"
echo "  [7] Назначить МОДЕРАТОРА (--make-moderator)"
echo "  ----------------------------------------------------"
echo "  [0] Выход"
echo -e "${BLUE}======================================================${RESET}"
read -p "Выберите действие [0-7]: " choice

case "$choice" in
    1) $DOCKER_COMPOSE -f docker-compose.yml -f docker-compose.no-av.yml up --build ;;
    2) $DOCKER_COMPOSE up --build ;;
    3) $DOCKER_COMPOSE down ;;
    4) $DOCKER_COMPOSE down -v ;;
    5) $DOCKER_COMPOSE logs -f ;;
    6) set_admin ;;
    7) set_mod ;;
    0) exit 0 ;;
    *) echo -e "${RED}Неверный выбор.${RESET}" ;;
esac