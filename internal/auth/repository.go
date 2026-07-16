package auth

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *User) error
	GetUserByNickname(nickname string) (*User, error)
	GetUsers(limit int, offset int) ([]User, int64, error)
	BlockUser(id int) error
	UnblockUser(id int) error
	WarnUser(id int, reason string) error
	UnwarnUser(id int) error
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

func (r *SqlUserRepo) BlockUser(id int) error {
	result := r.db.Model(&User{}).Where("id = ?", id).Update("is_blocked", true)
	if result.RowsAffected == 0 {
		return errors.New("Пользователь не найден")
	}

	return result.Error
}

func (r *SqlUserRepo) UnblockUser(id int) error {
	result := r.db.Model(&User{}).Where("id = ?", id).Update("is_blocked", false)
	if result.RowsAffected == 0 {
		return errors.New("Пользователь не найден")
	}

	return result.Error
}

func (r *SqlUserRepo) WarnUser(id int, reason string) error {
	tx := r.db.Begin()

	warning := Warning{
		UserID:    id,
		Reason:    reason,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(6 * 30 * 24 * time.Hour), //Пол года
	}

	if err := tx.Create(&warning).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(&User{}).Where("id = ?", id).UpdateColumn("warn_count", gorm.Expr("warn_count + ?", 1)).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *SqlUserRepo) UnwarnUser(id int) error {
	tx := r.db.Begin()

	var warning Warning

	if err := tx.Where("user_id = ?", id).Order("created_at desc").First(&warning).Error; err != nil {
		tx.Rollback()
		return errors.New("У пользователя нету варнов")
	}

	if err := tx.Delete(&warning).Error; err != nil {
		tx.Rollback()
		return err
	}

	//Уменьшение счётчика
	if err := tx.Model(&User{}).Where("id = ?", id).UpdateColumn("warn_count", gorm.Expr("warn_count - ?", 1)).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *SqlUserRepo) GetUsers(limit int, offset int) ([]User, int64, error) {
	var users []User
	var totalCount int64

	//Считывание общего кол-ва пользователей
	r.db.Model(&User{}).Count(&totalCount)

	//Доставание нужной пачки пользователей
	err := r.db.Preload("Warnings").Limit(limit).Offset(offset).Order("id desc").Find(&users).Error

	return users, totalCount, err
}
