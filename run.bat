@echo off
setlocal EnableDelayedExpansion
chcp 65001 >nul
title Управление EasyTranslator Server

:: Получаем ANSI Escape символ (ESC)
for /f %%a in ('echo prompt $E^| cmd') do set "ESC=%%a"

:: Проверка аргументов командной строки
if "%1"=="dev" (
    call :run_with_progress "docker compose -f docker-compose.yml -f docker-compose.no-av.yml up -d --build" "DEV (без ClamAV)"
    exit /b
)
if "%1"=="full" (
    call :run_with_progress "docker compose up -d --build" "PROD (с ClamAV)"
    exit /b
)
if "%1"=="prod" (
    call :run_with_progress "docker compose up -d --build" "PROD (с ClamAV)"
    exit /b
)
if "%1"=="stop" goto stop
if "%1"=="reset" goto reset
if "%1"=="logs" goto logs
if "%1"=="admin" goto cli_admin
if "%1"=="mod" goto cli_mod

:menu
cls
echo ======================================================
echo           EasyTranslator Server Management
echo ======================================================
echo.
echo  [1] Быстрый запуск DEV в фоне (Без ClamAV, ~120 MB RAM)
echo  [2] Полный запуск PROD в фоне (С ClamAV, ~1 GB RAM)
echo  [3] Остановить все контейнеры (down)
echo  [4] Полный сброс БД (Wipe Database)
echo  [5] Просмотр логов в реальном времени
echo  ----------------------------------------------------
echo  [6] Назначить АДМИНИСТРАТОРА (--make-admin)
echo  [7] Назначить МОДЕРАТОРА (--make-moderator)
echo  ----------------------------------------------------
echo  [0] Выход
echo.
echo ======================================================
set /p choice="Выберите действие [0-7]: "

if "%choice%"=="1" (
    call :run_with_progress "docker compose -f docker-compose.yml -f docker-compose.no-av.yml up -d --build" "DEV (без ClamAV)"
    pause
    goto menu
)
if "%choice%"=="2" (
    call :run_with_progress "docker compose up -d --build" "PROD (с ClamAV)"
    pause
    goto menu
)
if "%choice%"=="3" goto stop
if "%choice%"=="4" goto reset
if "%choice%"=="5" goto logs
if "%choice%"=="6" goto menu_admin
if "%choice%"=="7" goto menu_mod
if "%choice%"=="0" exit /b
goto menu

:: === Функция запуска с анимированным ползунком на одной строке ===
:run_with_progress
cls
set "CMD_TO_RUN=%~1"
set "MODE_NAME=%~2"
set "LOG_FILE=%TEMP%\easytranslator_build.log"
set "STATUS_FILE=%TEMP%\easytranslator_status.txt"

del "%LOG_FILE%" 2>nul
del "%STATUS_FILE%" 2>nul

echo ======================================================
echo  [INFO] Запуск в режиме: %MODE_NAME%
echo ======================================================
echo.

:: Запуск сборки в фоновом процессе (без лишних пробелов перед >)
start /b "" cmd /c "(%CMD_TO_RUN%) > "%LOG_FILE%" 2>&1 & (echo %%ERRORLEVEL%%)>"%STATUS_FILE%""

set /a progress=5
set "bar_total=25"

:loop_progress
if exist "%STATUS_FILE%" goto finish_progress

:: Увеличиваем процент пока идет процесс
if !progress! LSS 92 (
    set /a progress+=3
)

:: Формируем строку ползунка
set /a filled=(!progress! * bar_total) / 100
set /a empty=bar_total - filled
set "bar="
for /l %%i in (1,1,!filled!) do set "bar=!bar!█"
for /l %%i in (1,1,!empty!) do set "bar=!bar!░"

:: !ESC![1G возвращает курсор в колонку 1, !ESC![2K очищает строку
<nul set /p "=!ESC![1G!ESC![2KЗапуск контейнеров: [!bar!] !progress!%%"
ping 127.0.0.1 -n 2 >nul
goto loop_progress

:finish_progress
set /p EXIT_CODE=<"%STATUS_FILE%"
del "%STATUS_FILE%" 2>nul

:: Очищаем значение от возможных пробелов и символов переноса строки
set "EXIT_CODE=!EXIT_CODE: =!"

if "!EXIT_CODE!"=="0" (
    set "bar=█████████████████████████"
    <nul set /p "=!ESC![1G!ESC![2KЗапуск контейнеров: [!bar!] 100%%"
    echo.
    echo.
    echo ======================================================
    echo  [OK] Все контейнеры успешно собраны и запущены в фоне!
    echo  API доступен по адресу: http://localhost:8080
    echo ======================================================
    echo.
    del "%LOG_FILE%" 2>nul
) else (
    <nul set /p "=!ESC![1G!ESC![2KЗапуск контейнеров: [ ОШИБКА ]"
    echo.
    echo.
    echo ======================================================
    echo  [ERROR] Произошла ошибка при сборке/запуске!
    echo ======================================================
    echo.
    if exist "%LOG_FILE%" (
        type "%LOG_FILE%"
        del "%LOG_FILE%" 2>nul
    )
    echo.
)
exit /b

:stop
cls
echo [INFO] Остановка всех сервисов...
docker compose down
echo [OK] Все сервисы остановлены.
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
echo [INFO] Открытие логов (Ctrl+C для возврата)...
docker compose logs -f
goto menu

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
docker compose exec api ./myapi --make-admin %2
exit /b

:cli_mod
docker compose exec api ./myapi --make-moderator %2
exit /b