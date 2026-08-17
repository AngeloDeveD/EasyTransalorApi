package main

import (
	"fmt"
	"log"
	"strconv"

	"myapi/internal/auth"
	"myapi/internal/config"
	"myapi/internal/database"
)

func makeAdmin(cfg *config.Config, idStr string) {
	changeRole(cfg, idStr, "admin")
}

func makeModerator(cfg *config.Config, idStr string) {
	changeRole(cfg, idStr, "moderator")
}

func changeRole(cfg *config.Config, idStr string, role string) {
	// Конвертируем строку в число
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Fatalf("Ошибка: ID должен быть числом. Вы ввели: %s", idStr)
	}

	db, err := database.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Ошибка БД: %v", err)
	}

	var user auth.User
	// Ищем по Primary Key (ID)
	result := db.First(&user, id)

	if result.Error != nil {
		log.Fatalf("Пользователь с ID %d не найден!", id)
	}

	user.Role = role
	db.Save(&user)

	fmt.Printf("Пользователь '%s' (ID: %d) успешно назначен %s!\n", user.Nickname, id, role)
}
