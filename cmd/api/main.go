package main

import (
	"log"
	"myapi/internal/config"
	"os"
)

func main() {
	cfg := config.Load()

	// Обработка CLI команд
	if len(os.Args) > 2 && os.Args[1] == "--make-admin" {
		makeAdmin(cfg, os.Args[2])
		return
	}
	if len(os.Args) > 2 && os.Args[1] == "--make-moderator" {
		makeModerator(cfg, os.Args[2])
		return
	}

	// Обычный запуск сервера
	r := setupRouter(cfg)
	log.Println("Сервер запущен на порту :" + cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
