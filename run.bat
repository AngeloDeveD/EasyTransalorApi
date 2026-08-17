@echo off
chcp 65001 >nul
title Управление EasyTranslator Server

:: Проверка быстрых аргументов командной строки
if "%1"=="dev" goto dev
if "%1"=="full" goto full
if "%1"=="prod" goto full
if "%1"=="stop" goto stop
if "%1"=="reset" goto reset
if "%1"=="logs" goto logs
if "%1"=="admin" goto cli_admin
if "%1"=="mod" goto cli_mod
if "%1"=="moderator" goto cli_mod

:menu
cls
echo ======================================================
echo           EasyTranslator Server Management
echo ======================================================
echo.
echo  [1] Быстрый запуск DEV (Без ClamAV, ~120 MB RAM)
echo  [2] Полный запуск PROD (С ClamAV, ~1 GB RAM)
echo  [3] Остановить все контейнеры (down)
echo  [4] Полный сброс БД (Wipe Database)
echo  [5] Просмотр логов в реальном времени
echo ------------------------------------------------------
echo  [6] Назначить АДМИНИСТРАТОРА (--make-admin)
echo  [7] Назначить МОДЕРАТОРА (--make-moderator)
echo ------------------------------------------------------
echo  [0] Выход
echo.
echo ======================================================
set /p choice="Выберите действие [0-7]: "

if "%choice%"=="1" goto dev
if "%choice%"=="2" goto full
if "%choice%"=="3" goto stop
if "%choice%"=="4" goto reset
if "%choice%"=="5" goto logs
if "%choice%"=="6" goto menu_admin
if "%choice%"=="7" goto menu_mod
if "%choice%"=="0" exit /b
goto menu

:dev
cls
echo [INFO] Запуск DEV (No ClamAV)...
docker compose -f docker-compose.yml -f docker-compose.no-av.yml up --build
goto end

:full
cls
echo [INFO] Запуск PROD (с ClamAV)...
docker compose up --build
goto end

:stop
cls
echo [INFO] Остановка всех сервисов...
docker compose down
echo [OK] Контейнеры остановлены.
pause
goto menu

:reset
cls
echo [ВНИМАНИЕ] Это полностью удалит все данные из PostgreSQL!
set /p confirm="Вы уверены? (y/N): "
if /i "%confirm%"=="y" (
    docker compose down -v
    echo [OK] База данных и тома очищены.
)
pause
goto menu

:logs
cls
docker compose logs -f
goto end

:menu_admin
cls
set /p uid="Введите ID пользователя для назначения АДМИНОМ: "
if "%uid%"=="" goto menu
docker compose exec api ./myapi --make-admin %uid%
pause
goto menu

:menu_mod
cls
set /p uid="Введите ID пользователя для назначения МОДЕРАТОРОМ: "
if "%uid%"=="" goto menu
docker compose exec api ./myapi --make-moderator %uid%
pause
goto menu

:cli_admin
if "%2"=="" (
    echo [ERROR] Укажите ID пользователя. Пример: run.bat admin 1
    exit /b 1
)
docker compose exec api ./myapi --make-admin %2
exit /b

:cli_mod
if "%2"=="" (
    echo [ERROR] Укажите ID пользователя. Пример: run.bat mod 1
    exit /b 1
)
docker compose exec api ./myapi --make-moderator %2
exit /b

:end