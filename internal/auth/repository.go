package auth

import (
	"errors"

	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *User) error
	GetUserByNickname(nickname string) (*User, error)
}

type SqlUserRepo struct {
	db *gorm.DB
}

func NewSqlUserRepo(db *gorm.DB) *SqlUserRepo {
	return &SqlUserRepo{db: db}
}

func (r *SqlUserRepo) CreateUser(user *User) error {
	if err := r.db.Create(user).Error; err != nil {
		return errors.New("Ошибка создания пользователя (мейби никнейм уже занят)")
	}

	return nil
}

func (r *SqlUserRepo) GetUserByNickname(nickname string) (*User, error) {
	var user User
	err := r.db.Where("nickname = ?", nickname).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("Пользователь не найден")
		}
		return nil, err
	}

	return &user, nil
}
