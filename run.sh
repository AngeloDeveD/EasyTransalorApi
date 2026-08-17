#!/bin/bash

# Цвета для вывода
GREEN="\033[0;32m"
BLUE="\033[0;34m"
YELLOW="\033[1;33m"
RED="\033[0;31m"
CYAN="\033[0;36m"
RESET="\033[0m"

# Определение команды docker compose
if docker compose version &>/dev/null; then
    DOCKER_COMPOSE="docker compose"
elif command -v docker-compose &>/dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    echo -e "${RED}[ERROR] Docker Compose не установлен!${RESET}"
    exit 1
fi

# Функция запуска с анимированным ползунком
function run_with_progress() {
    local cmd="$1"
    local mode_name="$2"
    local log_file="/tmp/easytranslator_build.log"

    rm -f "$log_file"
    clear
    echo -e "${BLUE}======================================================${RESET}"
    echo -e " [INFO] Запуск в режиме: ${CYAN}${mode_name}${RESET}"
    echo -e "${BLUE}======================================================${RESET}"
    echo ""

    # Запускаем docker compose в фоне и пишем логи в файл
    eval "$cmd" > "$log_file" 2>&1 &
    local pid=$!

    local progress=5
    local bar_width=30

    # Анимация пока фоновый процесс работает
    while kill -0 $pid 2>/dev/null; do
        if [ $progress -lt 92 ]; then
            progress=$((progress + 3))
        fi

        local filled=$((progress * bar_width / 100))
        local empty=$((bar_width - filled))
        local bar=""
        for ((i=0; i<filled; i++)); do bar="${bar}█"; done
        for ((i=0; i<empty; i++)); do bar="${bar}░"; done

        printf "\r${CYAN}Запуск контейнеров:${RESET} [${YELLOW}%s${RESET}] %3d%% " "$bar" "$progress"
        sleep 0.3
    done

    # Ждем завершения и получаем код выхода
    wait $pid
    local exit_code=$?

    if [ $exit_code -eq 0 ]; then
        local full_bar=""
        for ((i=0; i<bar_width; i++)); do full_bar="${full_bar}█"; done
        printf "\r${CYAN}Запуск контейнеров:${RESET} [${GREEN}%s${RESET}] 100%%\n\n" "$full_bar"
        echo -e "${GREEN}======================================================${RESET}"
        echo -e "${GREEN} [OK] Все сервисы успешно собраны и запущены в фоне!  ${RESET}"
        echo -e " API доступен: ${CYAN}http://localhost:8080${RESET}"
        echo -e "${GREEN}======================================================${RESET}"
        rm -f "$log_file"
    else
        printf "\r${CYAN}Запуск контейнеров:${RESET} [${RED} ОШИБКА ${RESET}]\n\n"
        echo -e "${RED}======================================================${RESET}"
        echo -e "${RED} [ERROR] Произошла ошибка при сборке/запуске!         ${RESET}"
        echo -e "${RED}======================================================${RESET}\n"
        if [ -f "$log_file" ]; then
            cat "$log_file"
            rm -f "$log_file"
        fi
        echo ""
    fi
}

function stop_all() {
    echo -e "${YELLOW}[INFO] Остановка контейнеров...${RESET}"
    $DOCKER_COMPOSE down
    echo -e "${GREEN}[OK] Контейнеры остановлены.${RESET}"
}

function reset_db() {
    echo -e "${RED}[ВНИМАНИЕ] Это действие удалит все данные из PostgreSQL!${RESET}"
    read -p "Вы уверены? (y/N): " confirm
    if [[ "$confirm" =~ ^[Yy]$ ]]; then
        $DOCKER_COMPOSE down -v
        echo -e "${GREEN}[OK] База данных и тома удалены.${RESET}"
    fi
}

function show_logs() {
    echo -e "${BLUE}[INFO] Открытие логов. Нажмите Ctrl+C для выхода...${RESET}"
    $DOCKER_COMPOSE logs -f
}

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

# Обработка аргументов CLI
case "$1" in
    dev)
        run_with_progress "$DOCKER_COMPOSE -f docker-compose.yml -f docker-compose.no-av.yml up -d --build" "DEV (без ClamAV)"
        exit 0
        ;;
    full|prod)
        run_with_progress "$DOCKER_COMPOSE up -d --build" "PROD (с ClamAV)"
        exit 0
        ;;
    stop) stop_all; exit 0 ;;
    reset) reset_db; exit 0 ;;
    logs) show_logs; exit 0 ;;
    admin) set_admin "$2"; exit 0 ;;
    mod|moderator) set_mod "$2"; exit 0 ;;
esac

# Главное интерактивное меню
while true; do
    clear
    echo -e "${BLUE}======================================================${RESET}"
    echo -e "           EasyTranslator Server Management           "
    echo -e "${BLUE}======================================================${RESET}"
    echo "  [1] Быстрый запуск DEV (Без ClamAV, ~120 MB RAM)"
    echo "  [2] Полный запуск PROD (С ClamAV, ~1 GB RAM)"
    echo "  [3] Остановить все контейнеры"
    echo "  [4] Полный сброс базы данных (Wipe Database)"
    echo "  [5] Просмотр логов в реальном времени"
    echo "  ----------------------------------------------------"
    echo "  [6] Назначить АДМИНИСТРАТОРА (--make-admin)"
    echo "  [7] Назначить МОДЕРАТОРА (--make-moderator)"
    echo "  ----------------------------------------------------"
    echo "  [0] Выход"
    echo -e "${BLUE}======================================================${RESET}"
    read -p "Выберите действие [0-7]: " choice

    case "$choice" in
        1)
            run_with_progress "$DOCKER_COMPOSE -f docker-compose.yml -f docker-compose.no-av.yml up -d --build" "DEV (без ClamAV)"
            read -p "Нажмите Enter для возврата в меню..."
            ;;
        2)
            run_with_progress "$DOCKER_COMPOSE up -d --build" "PROD (с ClamAV)"
            read -p "Нажмите Enter для возврата в меню..."
            ;;
        3) stop_all; read -p "Нажмите Enter для возврата в меню..." ;;
        4) reset_db; read -p "Нажмите Enter для возврата в меню..." ;;
        5) show_logs ;;
        6) set_admin; read -p "Нажмите Enter для возврата в меню..." ;;
        7) set_mod; read -p "Нажмите Enter для возврата в меню..." ;;
        0) exit 0 ;;
        *) echo -e "${RED}Неверный выбор.${RESET}"; sleep 1 ;;
    esac
done