package chat

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	Hub       *Hub
	Repo      ChatRepository
	CryptoKey []byte //Ключ для шифрования
}

func NewChatHandler(hub *Hub, repo ChatRepository, cryptoKey []byte) *ChatHandler {
	return &ChatHandler{Hub: hub, Repo: repo, CryptoKey: cryptoKey}
}

// Переводчик из http в WS
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Структура сообщения, которые ожидается от клиента по ws
// Пример json: {"to": 1, "text": "Привет, говноед!"}
type IncomingMessage struct {
	To   int    `json:"to"`   //Кому отправка (userID)
	Text string `json:"text"` //Текст сообщения
}

// GET /api/chat/ws
func (h *ChatHandler) HandleChat(c *gin.Context) {
	//Получение текущего ID пользователя
	userID, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользотваель не авторизован"})
		return
	}
	myID := userID.(int)

	//Переделка соединения в WebSocket

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Ошибка апгрейда WS: %v", err)
		return
	}

	//Регистрация юзера в хаб
	h.Hub.Register(myID, conn)

	//Гарантийное отключение пользователя при выходе из функции
	defer h.Hub.Unregistered(myID)

	log.Printf("Пользователь %d подключился к чату", myID)

	//Бесконечный цикл для чтения
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Пользователь %d отключился", myID)
			break
		}

		//Парсинг JSON, который был отправлен клиентом
		var incomingMsg IncomingMessage
		err = json.Unmarshal(msgBytes, &incomingMsg)
		if err != nil {
			//если был отправлен мусор
			h.Hub.SendMessage(myID, []byte(`{"error": "Невернывй формат сообщения"}`))
			continue
		}

		//Шифрование
		encryptedText, err := Encrypt(incomingMsg.Text, h.CryptoKey)
		if err != nil {
			log.Println("Ошибка шифрования:", err)
			continue
		}

		dbMsg := &ChatMessage{
			FromUserID:    myID,
			ToUserId:      incomingMsg.To,
			EncryptedText: encryptedText,
		}
		if err := h.Repo.SaveMessage(dbMsg); err != nil {
			log.Println("Ошибка сохранения сообщенитя в БД:", err)
			continue
		}

		//Формирование исходящего сообщения для получателя
		outgoingMsg := map[string]interface{}{
			"from": myID,
			"text": incomingMsg.Text,
		}
		outgoingBytes, _ := json.Marshal(outgoingMsg)

		//Отправка через Hub пользователю
		_ = h.Hub.SendMessage(incomingMsg.To, outgoingBytes)
	}
}

// GET /api/chat/history/:userId
func (h *ChatHandler) GetHistory(c *gin.Context) {
	myID, _ := c.Get("userID")
	otherUserIDStr := c.Param("userId")

	otherUserID, err := strconv.Atoi(otherUserIDStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID неправильного формата"})
		return
	}

	//Получение сообщений из бд
	messages, err := h.Repo.GetHistory(myID.(int), otherUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получении истории"})
		return
	}

	//Расшифровка сообщений
	type DecryptedMessage struct {
		FromUserID int    `json:"fromUserId"`
		Text       string `json:"text"`
	}
	var result []DecryptedMessage

	for _, msg := range messages {
		decryptedText, err := Decrypt(msg.EncryptedText, h.CryptoKey)
		if err != nil {
			continue
		}
		result = append(result, DecryptedMessage{
			FromUserID: msg.FromUserID,
			Text:       decryptedText,
		})
	}

	c.JSON(http.StatusOK, result)
}
