package chat

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

// Обёртка над подключением, чтобюы избежать конфликтов при одновременной записи
type Client struct {
	conn *websocket.Conn
	mu   sync.Mutex //Замок, чтобы два сообщения не отправились одновременно и не сломали соединение
}

// Хаб
type Hub struct {
	clients map[int]*Client //Карта: UserID -> Подключение
	mu      sync.RWMutex    //Замок для безопасного чтения/записи самой карты
}

func NewHub() *Hub {
	return &Hub{clients: make(map[int]*Client)}
}

// Подкелючение нового юзера к хабу
func (h *Hub) Register(userID int, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	//Проверка на старое подключение и его закрытие
	if oldClient, ok := h.clients[userID]; ok {
		oldClient.conn.Close()
	}

	h.clients[userID] = &Client{conn: conn}
}

// Отключение пользователя
func (h *Hub) Unregistered(userID int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client, ok := h.clients[userID]; ok {
		client.conn.Close()
		delete(h.clients, userID)
	}
}

// Отправка сообщения конкреному пользователю
func (h *Hub) SendMessage(toUserID int, message []byte) error {
	h.mu.RLock()
	client, ok := h.clients[toUserID]
	h.mu.RUnlock()

	if !ok {
		return errors.New("Пользователь не в сети!")
	}

	//Блокировка замка пользователя перед отправкой, чтобы устранить race condition
	client.mu.Lock()
	defer client.mu.Unlock()

	err := client.conn.WriteMessage(websocket.TextMessage, message)
	return err
}
