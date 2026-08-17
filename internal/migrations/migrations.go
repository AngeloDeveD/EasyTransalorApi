package migrations

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	ID       string
	SQLite   []string
	Postgres []string
}

type SchemaMigration struct {
	ID        string    `gorm:"primaryKey;size:255"`
	AppliedAt time.Time `gorm:"not null"`
}

func Run(db *gorm.DB, dbType string) error {
	if err := db.AutoMigrate(&SchemaMigration{}); err != nil {
		return fmt.Errorf("ошибка создания таблицы миграций: %w", err)
	}

	for _, migration := range allMigrations {
		applied, err := isApplied(db, migration.ID)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		statements := migration.SQLite
		if strings.EqualFold(dbType, "postgres") {
			statements = migration.Postgres
		}

		if err := apply(db, migration.ID, statements); err != nil {
			return err
		}
	}

	return nil
}

func isApplied(db *gorm.DB, id string) (bool, error) {
	var count int64
	if err := db.Model(&SchemaMigration{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, fmt.Errorf("ошибка проверки миграции %s: %w", id, err)
	}
	return count > 0, nil
}

func apply(db *gorm.DB, id string, statements []string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return fmt.Errorf("ошибка применения миграции %s: %w", id, err)
			}
		}

		return tx.Create(&SchemaMigration{
			ID:        id,
			AppliedAt: time.Now().UTC(),
		}).Error
	})
}
