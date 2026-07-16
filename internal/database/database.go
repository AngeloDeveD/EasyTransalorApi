package database

import (
	"fmt"
	"log"
	"myapi/internal/auth"
	"myapi/internal/config"
	"myapi/internal/game"
	"myapi/internal/notification"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ConnectDB выбирает драйвер на основе конфига
func ConnectDB(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	if cfg.DBType == "postgres" {
		// Формируем строку подключения (DSN) для PostgreSQL
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
			cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort)

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
		}
		log.Println("Подключено к PostgreSQL")
	} else {
		// По умолчанию используем SQLite
		db, err = gorm.Open(sqlite.Open(cfg.DBName), &gorm.Config{})
		if err != nil {
			return nil, fmt.Errorf("ошибка подключения к SQLite: %w", err)
		}
		log.Println("Подключено к SQLite")
	}

	// Миграции (работают одинаково для обеих БД!)
	err = db.AutoMigrate(
		&auth.User{},
		&auth.Warning{},
		&game.GameCard{},
		&game.GameInfo{},
		&game.TranslateCard{},
		&notification.Notification{},
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка миграции БД: %w", err)
	}

	return db, nil
}
