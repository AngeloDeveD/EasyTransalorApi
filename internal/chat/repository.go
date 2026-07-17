package chat

import "gorm.io/gorm"

type ChatRepository interface {
	SaveMesage(msg *ChatMessage) error
	GetHistory(user1ID int, user2ID int) ([]ChatMessage, error)
}

type SqliteChatRepo struct {
	db *gorm.DB
}

func NewSqliteChatRepo(db *gorm.DB) *SqliteChatRepo {
	return &SqliteChatRepo{db: db}
}

func (r *SqliteChatRepo) SaveMessage(msg *ChatMessage) error {
	return r.db.Create(msg).Error
}

func (r *SqliteChatRepo) GetHistory(user1ID int, user2ID int) ([]ChatMessage, error) {
	var messages []ChatMessage
	//Получение переписки между двуми юзерами
	err := r.db.Where(
		"(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
		user1ID, user2ID, user2ID, user1ID,
	).Order("created_at asc").Find(&messages).Error

	return messages, err
}
