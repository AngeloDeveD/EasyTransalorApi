package auth

import "time"

type User struct {
	ID           int       `json:"id" gorm:"primaryKey;autoIncrement"`
	FirstName    string    `json:"firstName" gorm:"not null"`
	SecondName   string    `json:"secondName" gorm:"not null"`
	Nickname     string    `json:"nickname" gorm:"uniqueIndex;not null"` //Уникальный никнейм
	PasswordHash string    `json:"-" gorm:"not null"`
	Role         string    `json:"role" gorm:"default:author"` //Роль: либо автор либо админ
	CreatedAt    time.Time `json:"createdAt"`
	IsBlocked    bool      `json:"isBlocked" gorm:"default:false"`
	WarnCount    int       `json:"warnCount" gorm:"default:0"`
	Warnings     []Warning `json:"warnings,omitempty" gorm:"foreignKey:UserID"`
}

type Warning struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int       `json:"userId"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type Notification struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type RegisterRequest struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Nickname  string `json:"nickname" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Nickname string `json:"nickname" binding:"required"`
	Password string `json:"password" binding:"required"`
}
