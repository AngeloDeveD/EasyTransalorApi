package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"myapi/internal/notification"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	m.Run()
}

// InMemoryUserRepo для тестов auth
type InMemoryUserRepo struct {
	users []User
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{users: []User{}}
}

func (r *InMemoryUserRepo) CreateUser(user *User) error {
	for _, u := range r.users {
		if u.Nickname == user.Nickname {
			return errors.New("никнейм занят")
		}
	}
	user.ID = len(r.users) + 1
	r.users = append(r.users, *user)
	return nil
}

func (r *InMemoryUserRepo) GetUserByNickname(nickname string) (*User, error) {
	for i, u := range r.users {
		if u.Nickname == nickname {
			return &r.users[i], nil
		}
	}
	return nil, errors.New("пользователь не найден")
}

func (r *InMemoryUserRepo) GetUserById(id int) (*User, error) {
	for i, u := range r.users {
		if u.ID == id {
			return &r.users[i], nil
		}
	}
	return nil, errors.New("пользователь не найден")
}

func (r *InMemoryUserRepo) GetUsers(limit int, offset int) ([]User, int64, error) {
	total := int64(len(r.users))
	if offset >= len(r.users) {
		return []User{}, total, nil
	}
	end := offset + limit
	if end > len(r.users) {
		end = len(r.users)
	}
	return r.users[offset:end], total, nil
}

func (r *InMemoryUserRepo) BlockUser(id int) error {
	for i := range r.users {
		if r.users[i].ID == id {
			r.users[i].IsBlocked = true
			return nil
		}
	}
	return errors.New("не найден")
}

func (r *InMemoryUserRepo) UnblockUser(id int) error {
	for i := range r.users {
		if r.users[i].ID == id {
			r.users[i].IsBlocked = false
			return nil
		}
	}
	return errors.New("не найден")
}

func (r *InMemoryUserRepo) WarnUser(id int, reason string) error {
	for i := range r.users {
		if r.users[i].ID == id {
			r.users[i].WarnCount++
			r.users[i].Warnings = append(r.users[i].Warnings, Warning{Reason: reason})
			return nil
		}
	}
	return errors.New("не найден")
}

func (r *InMemoryUserRepo) UnwarnUser(id int) error {
	for i := range r.users {
		if r.users[i].ID == id {
			if r.users[i].WarnCount > 0 {
				r.users[i].WarnCount--
				return nil
			}
		}
	}
	return errors.New("нет варнов")
}

// Локальный мок для уведомлений (чтобы не тянуть БД в тесты).
// Сигнатуры должны точно совпадать с notification.NotificationRepository.
type mockNotifRepo struct{}

func (m *mockNotifRepo) Create(n *notification.Notification) error { return nil }
func (m *mockNotifRepo) GetForUser(userID int) ([]notification.Notification, error) {
	return nil, nil
}

// Настройка роутера с реальными middleware
func setupTestRouter() *gin.Engine {
	repo := NewInMemoryUserRepo()
	jwtManager := NewJWTManager("test_secret_key")
	authHandler := NewAuthHandler(repo, jwtManager)

	// WarnUser отправляет уведомление, поэтому передаём рабочий мок, а не nil.
	adminHandler := NewAdminHandler(repo, &mockNotifRepo{})

	r := gin.New()

	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.GET("/me", AuthMiddleware(jwtManager), authHandler.Me)
	}

	adminGroup := r.Group("/api/admin", AuthMiddleware(jwtManager), AdminMiddleware())
	{
		adminGroup.GET("/users", adminHandler.GetUsers)
		adminGroup.PATCH("/users/:userid/block", adminHandler.BlockUser)
		adminGroup.PATCH("/users/:userid/unblock", adminHandler.UnblockUser)
		adminGroup.PATCH("/users/:userid/warn", adminHandler.WarnUser)
		adminGroup.PATCH("/users/:userid/unwarn", adminHandler.UnwarnUser)
	}

	return r
}

func makeRequest(r *gin.Engine, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonData, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, _ := http.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func getAdminToken(jwt *JWTManager) string {
	token, _ := jwt.GenerateToken(999, "admin")
	return token
}

// --- ТЕСТЫ АВТОРИЗАЦИИ ---
func TestRegisterSuccess(t *testing.T) {
	r := setupTestRouter()
	body := map[string]string{"firstName": "И", "lastName": "И", "nickname": "ivan", "password": "123"}
	w := makeRequest(r, "POST", "/api/auth/register", body, "")
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestLoginSuccess(t *testing.T) {
	r := setupTestRouter()
	regBody := map[string]string{"firstName": "Т", "lastName": "Т", "nickname": "test", "password": "123"}
	makeRequest(r, "POST", "/api/auth/register", regBody, "")

	loginBody := map[string]string{"nickname": "test", "password": "123"}
	w := makeRequest(r, "POST", "/api/auth/login", loginBody, "")
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.NotEmpty(t, response["token"])
}

func TestRegisterDuplicate(t *testing.T) {
	r := setupTestRouter()
	body := map[string]string{"firstName": "И", "lastName": "И", "nickname": "dup", "password": "123"}
	makeRequest(r, "POST", "/api/auth/register", body, "")
	w := makeRequest(r, "POST", "/api/auth/register", body, "")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- ТЕСТЫ АДМИН-ПАНЕЛИ ---
func TestAdminGetUsers(t *testing.T) {
	r := setupTestRouter()
	regBody := map[string]string{"firstName": "Ю", "lastName": "Ю", "nickname": "user1", "password": "123"}
	makeRequest(r, "POST", "/api/auth/register", regBody, "")

	adminToken := getAdminToken(NewJWTManager("test_secret_key"))
	w := makeRequest(r, "GET", "/api/admin/users", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user1")
}

func TestAdminBlockUser(t *testing.T) {
	r := setupTestRouter()
	regBody := map[string]string{"firstName": "Ю", "lastName": "Ю", "nickname": "user1", "password": "123"}
	makeRequest(r, "POST", "/api/auth/register", regBody, "")

	adminToken := getAdminToken(NewJWTManager("test_secret_key"))
	w := makeRequest(r, "PATCH", "/api/admin/users/1/block", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "заблокирован")
}

func TestAdminWarnUser(t *testing.T) {
	r := setupTestRouter()
	regBody := map[string]string{"firstName": "Ю", "lastName": "Ю", "nickname": "user1", "password": "123"}
	makeRequest(r, "POST", "/api/auth/register", regBody, "")

	adminToken := getAdminToken(NewJWTManager("test_secret_key"))
	warnBody := map[string]string{"reason": "Залил вирус"}
	w := makeRequest(r, "PATCH", "/api/admin/users/1/warn", warnBody, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Варн успешно выдан")
}

func TestAdminForbiddenForAuthor(t *testing.T) {
	r := setupTestRouter()
	regBody := map[string]string{"firstName": "А", "lastName": "А", "nickname": "author1", "password": "123"}
	makeRequest(r, "POST", "/api/auth/register", regBody, "")

	loginBody := map[string]string{"nickname": "author1", "password": "123"}
	wLogin := makeRequest(r, "POST", "/api/auth/login", loginBody, "")
	var loginResp map[string]interface{}
	json.Unmarshal(wLogin.Body.Bytes(), &loginResp)
	authorToken := loginResp["token"].(string)

	w := makeRequest(r, "GET", "/api/admin/users", nil, authorToken)
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestAdminGetUsersPagination(t *testing.T) {
	r := setupTestRouter()
	adminToken := getAdminToken(NewJWTManager("test_secret_key"))

	for i := 1; i <= 3; i++ {
		body := map[string]string{"firstName": "И", "lastName": "И", "nickname": fmt.Sprintf("user_%d", i), "password": "123"}
		makeRequest(r, "POST", "/api/auth/register", body, "")
	}

	w := makeRequest(r, "GET", "/api/admin/users?page=1&limit=2", nil, adminToken)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(3), resp["total"])
}
