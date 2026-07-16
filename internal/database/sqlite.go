package database

import (
	"myapi/internal/auth"
	"myapi/internal/game"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConnectSqlite(dbName string) (*gorm.DB, error) {
	//Открытие файла
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	if err := db.Exec("PRAGMA journal_mode=WAL;").Error; err != nil {
		return nil, err
	}

	//Авто-миграция: автоматическое создание таблиц на основе структур
	err = db.AutoMigrate(
		&game.GameCard{},
		&game.GameInfo{},
		&game.TranslateCard{},
		&auth.User{},
		&auth.Warning{},
	)

	if err != nil {
		return nil, err
	}

	return db, nil
}
