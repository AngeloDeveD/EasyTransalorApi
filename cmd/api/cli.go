package main

import (
	"fmt"
	"log"
	"myapi/internal/auth"
	"myapi/internal/config"
	"myapi/internal/database"
)

func makeAdmin(cfg *config.Config, nickname string) {
	changeRole(cfg, nickname, "admin")
}

func makeModerator(cfg *config.Config, nickname string) {
	changeRole(cfg, nickname, "moderator")
}

func changeRole(cfg *config.Config, nickname string, role string) {
	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Ошибка БД: %v", err)
	}

	var user auth.User
	result := db.Where("nickname = ?", nickname).First(&user)

	if result.Error != nil {
		log.Fatalf("Пользователь '%s' не найден!", nickname)
	}

	user.Role = role
	db.Save(&user)

	fmt.Println("Пользователь '%s' успешно назначен %s!\n", nickname, role)
}
