package chat

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	Hub *Hub
}

func NewChatHandler(hub *Hub) *ChatHandler {
	return &ChatHandler{Hub: hub}
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
		log.Printf("Ошибка апгрейда WS:", err)
		return
	}

	//Регистрация юзера в хаб
	h.Hub.Register(myID, conn)

	//Гарантийное отключение пользователя при выходе из функции
	defer h.Hub.Unregistered(myID)

	log.Printf("Пользотвалеь %d подключился к чату", myID)

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

		//Формирование исходящего сообщения для получателя
		outgoingMsg := map[string]interface{}{
			"from": myID,
			"text": incomingMsg.Text,
		}
		outgoingBytes, _ := json.Marshal(outgoingMsg)

		//Отправка через Hub пользователю
		err = h.Hub.SendMessage(incomingMsg.To, outgoingBytes)
		if err != nil {
			//Если пользотвалеь не в сети, то сообщаем об этом отправителю
			errorMag := map[string]string{"error": "Пользователь " + string(rune(incomingMsg.To)) + "ещё не в сети"}
			errBytes, _ := json.Marshal(errorMag)
			h.Hub.SendMessage(myID, errBytes)
		}
	}
}
