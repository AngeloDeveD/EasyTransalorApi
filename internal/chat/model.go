package chat

import "time"

type ChatMessage struct {
	ID         int `json:"id" gorm:"primaryKey;autoIncrement"`
	FromUserID int `json:"fromUserId" gorm:"index"`
	ToUserId   int `json:"toUserId" gorm:"index"`
	//Зашифрованный текст будет лежать в базе в виде среза байтов
	EncryptedText []byte    `json:"-"`
	IsRead        bool      `json:"isRead" gorm:"default:false"`
	CreatedAt     time.Time `json:"createdAt"`
}
