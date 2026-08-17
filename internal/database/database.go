package database

import (
	"fmt"
	"log"
	"myapi/internal/audit"
	"myapi/internal/auth"
	"myapi/internal/chat"
	"myapi/internal/config"
	"myapi/internal/game"
	"myapi/internal/migrations"
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

	if err := migrations.Run(db, cfg.DBType); err != nil {
		return nil, fmt.Errorf("ошибка применения миграций БД: %w", err)
	}

	// Compatibility pass для разработки: миграции фиксируют историю схемы,
	// AutoMigrate страхует тесты и локальную БД от мелких расхождений моделей.
	err = db.AutoMigrate(
		&audit.Event{},
		&auth.User{},
		&auth.Warning{},
		&game.GameCard{},
		&game.GameInfo{},
		&game.TranslateCard{},
		&notification.Notification{},
		&chat.ChatMessage{},
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка синхронизации моделей БД: %w", err)
	}

	return db, nil
}
