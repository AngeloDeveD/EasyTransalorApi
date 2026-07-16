package notification

import "time"

type Notification struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int       `json:"userId"`
	Title     string    `json:"title" gorm:"not null"`
	Message   string    `json:"message" gorm:"not null"`
	IsRead    bool      `json:"isRead" gorm:"default:false"`
	CreatedAt time.Time `json:"createdAt"`
}
