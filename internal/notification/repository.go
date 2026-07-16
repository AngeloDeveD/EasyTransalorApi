package notification

import (
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(n *Notification) error
	GetForUser(userID int) ([]Notification, error)
}

// 1. Структура репозитория должна называться иначе, чем модель!
type SqliteNotificationRepo struct {
	db *gorm.DB
}

// 2. Конструктор должен возвращать указатель на РЕПОЗИТОРИЙ, а не на модель
func NewSqlNotificationRepo(db *gorm.DB) *SqliteNotificationRepo {
	return &SqliteNotificationRepo{db: db}
}

func (r *SqliteNotificationRepo) Create(n *Notification) error {
	return r.db.Create(n).Error
}

func (r *SqliteNotificationRepo) GetForUser(userID int) ([]Notification, error) {
	var notifications []Notification
	err := r.db.Where("user_id = ? OR user_id = ?", 0, userID).Order("created_at desc").Find(&notifications).Error
	return notifications, err
}
