package auth

type User struct {
	ID           int    `json:"id" gorm:"primaryKey;autoIncrement"`
	FirstName    string `json:"firstName" gorm:"not null"`
	SecondName   string `json:"secondName" gorm:"not null"`
	Nickname     string `json:"nickname" gorm:"uniqueIndex;not null"` //Уникальный никнейм
	PasswordHash string `json:"-" gorm:"not null"`
	Role         string `json:"role" gorm:"default:author"` //Роль: либо автор либо админ
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
