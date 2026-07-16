package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	m.Run()
}

// InMemoryNotificationRepo для изолированных тестов
type InMemoryNotificationRepo struct {
	notifications []Notification
}

func NewInMemoryNotificationRepo() *InMemoryNotificationRepo {
	return &InMemoryNotificationRepo{notifications: []Notification{}}
}

func (r *InMemoryNotificationRepo) Create(n *Notification) error {
	n.ID = len(r.notifications) + 1
	r.notifications = append(r.notifications, *n)
	return nil
}

func (r *InMemoryNotificationRepo) GetForUser(userID int) ([]Notification, error) {
	var result []Notification
	for _, n := range r.notifications {
		if n.UserID == 0 || n.UserID == userID {
			result = append(result, n)
		}
	}
	return result, nil
}

// setupTestRouter настраивает роутер. Мы можем передать userID, чтобы имитировать конкретного юзера
func setupTestRouter(userID int) *gin.Engine {
	repo := NewInMemoryNotificationRepo()
	handler := NewNotificationHandler(repo)

	r := gin.New()

	// Заглушка для авторизации: кладет нужный userID в контекст
	authMw := func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}
	// Заглушка для админа: просто пропускает дальше
	adminMw := func(c *gin.Context) {
		c.Next()
	}

	userGroup := r.Group("/api/notifications", authMw)
	{
		userGroup.GET("", handler.GetMyNotifications)
	}

	adminGroup := r.Group("/api/admin/notifications", authMw, adminMw)
	{
		adminGroup.POST("", handler.CreateNotification)
	}

	return r
}

func makeRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonData, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, _ := http.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// func (r *InMemoryUserRepo) GetUserByID(id int) (*User, error) {
//     for i, u := range r.users {
//         if u.ID == id {
//             return &r.users[i], nil
//         }
//     }
//     return nil, errors.New("пользователь не найден")
// }

// 1. Тест: Админ успешно создает глобальную рассылку
func TestCreateGlobalNotification(t *testing.T) {
	r := setupTestRouter(999) // 999 - условный ID админа
	body := map[string]interface{}{
		"title":    "Технические работы",
		"message":  "Сервер будет недоступен сегодня ночью.",
		"isGlobal": true,
	}

	w := makeRequest(r, "POST", "/api/admin/notifications", body)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Уведомление успещно создано")
}

// 2. Тест: Ошибка при отсутствии текста (проверка binding:"required")
func TestCreateNotificationMissingFields(t *testing.T) {
	r := setupTestRouter(999)
	body := map[string]interface{}{
		"title": "Заголовок есть",
		// Сообщение отсутствует
	}

	w := makeRequest(r, "POST", "/api/admin/notifications", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Заполните title и message")
}

// 3. Тест: Ошибка при попытке создать личное уведомление без указания UserID
func TestCreatePersonalNotificationWithoutUserID(t *testing.T) {
	r := setupTestRouter(999)
	body := map[string]interface{}{
		"title":    "Привет",
		"message":  "Это тебе",
		"isGlobal": false,
	}

	w := makeRequest(r, "POST", "/api/admin/notifications", body)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Укажите userId")
}

// 4. Тест: Пользователь получает свои и глобальные уведомления, но не чужие
func TestGetMyNotifications(t *testing.T) {
	repo := NewInMemoryNotificationRepo()
	handler := NewNotificationHandler(repo)

	// Заполняем репозиторий тестовыми данными напрямую
	repo.Create(&Notification{UserID: 0, Title: "Глобальное", Message: "Для всех"})
	repo.Create(&Notification{UserID: 5, Title: "Личное для 5", Message: "Только для юзера 5"})
	repo.Create(&Notification{UserID: 6, Title: "Чужое", Message: "Для юзера 6"})

	// Настраиваем роутер, имитируя вход пользователя с ID = 5
	r := gin.New()
	r.GET("/api/notifications", func(c *gin.Context) {
		c.Set("userID", 5)
		handler.GetMyNotifications(c)
	})

	w := makeRequest(r, "GET", "/api/notifications", nil)
	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что в ответе ровно 2 уведомления (Глобальное и Личное для 5)
	var resp []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	assert.Len(t, resp, 2)

	// Проверяем, что среди них нет чужого уведомления
	titles := fmt.Sprintf("%v", resp)
	assert.Contains(t, titles, "Глобальное")
	assert.Contains(t, titles, "Личное для 5")
	assert.NotContains(t, titles, "Чужое")
}
