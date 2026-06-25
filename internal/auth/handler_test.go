package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestMain настраивает окружение для всех тестов в пакете
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	m.Run()
}

// InMemoryUserRepo - заглушка репозитория для тестов (как у тебя в game)
type InMemoryUserRepo struct {
	users []User
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{users: []User{}}
}

func (r *InMemoryUserRepo) CreateUser(user *User) error {
	for _, u := range r.users {
		if u.Nickname == user.Nickname {
			return assert.AnError // Имитируем ошибку уникальности
		}
	}
	user.ID = len(r.users) + 1
	r.users = append(r.users, *user)
	return nil
}

func (r *InMemoryUserRepo) GetUserByNickname(nickname string) (*User, error) {
	for _, u := range r.users {
		if u.Nickname == nickname {
			return &u, nil
		}
	}
	return nil, assert.AnError // Имитируем "пользователь не найден"
}

// setupTestRouter поднимает тестовый роутер с InMemory репозиторием
func setupTestRouter() *gin.Engine {
	repo := NewInMemoryUserRepo()
	jwtManager := NewJWTManager("test_secret_key")
	handler := NewAuthHandler(repo, jwtManager)

	r := gin.New()

	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", handler.Register)
		authGroup.POST("/login", handler.Login)
		authGroup.GET("/me", AuthMiddleware(jwtManager), handler.Me)
	}

	return r
}

// Функция-помощник для отправки JSON запросов
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

// 1. Тест успешной регистрации
func TestRegisterSuccess(t *testing.T) {
	r := setupTestRouter()
	body := map[string]string{
		"firstName": "Иван",
		"lastName":  "Иванов",
		"nickname":  "ivan123",
		"password":  "supersecret",
	}

	w := makeRequest(r, "POST", "/api/auth/register", body, "")
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "Пользователь успешно зарегистрирован", response["message"])
}

// 2. Тест регистрации с ошибкой (не заполнены поля)
func TestRegisterBadRequest(t *testing.T) {
	r := setupTestRouter()
	body := map[string]string{
		"firstName": "Иван",
	}

	w := makeRequest(r, "POST", "/api/auth/register", body, "")
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// 3. Тест успешного логина
func TestLoginSuccess(t *testing.T) {
	r := setupTestRouter()

	// Сначала регистрируем пользователя
	regBody := map[string]string{
		"firstName": "Тест",
		"lastName":  "Тестов",
		"nickname":  "testuser",
		"password":  "password123",
	}
	makeRequest(r, "POST", "/api/auth/register", regBody, "")

	// Теперь пытаемся залогиниться
	loginBody := map[string]string{
		"nickname": "testuser",
		"password": "password123",
	}
	w := makeRequest(r, "POST", "/api/auth/login", loginBody, "")

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	token, ok := response["token"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, token)
}

// 4. Тест логина с неверным паролем
func TestLoginWrongPassword(t *testing.T) {
	r := setupTestRouter()

	regBody := map[string]string{
		"firstName": "Тест",
		"lastName":  "Тестов",
		"nickname":  "testuser2",
		"password":  "correct_password",
	}
	makeRequest(r, "POST", "/api/auth/register", regBody, "")

	loginBody := map[string]string{
		"nickname": "testuser2",
		"password": "wrong_password",
	}
	w := makeRequest(r, "POST", "/api/auth/login", loginBody, "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Неверный никнейм или пароль")
}

// 5. Тест доступа к защищенному роуту БЕЗ токена
func TestProtectedRouteNoToken(t *testing.T) {
	r := setupTestRouter()

	w := makeRequest(r, "GET", "/api/auth/me", nil, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Требуется авторизация")
}

// 6. Тест доступа к защищенному роуту С токеном
func TestProtectedRouteWithToken(t *testing.T) {
	r := setupTestRouter()

	regBody := map[string]string{
		"firstName": "Админ",
		"lastName":  "Админов",
		"nickname":  "admin",
		"password":  "adminpass",
	}
	makeRequest(r, "POST", "/api/auth/register", regBody, "")

	loginBody := map[string]string{
		"nickname": "admin",
		"password": "adminpass",
	}
	wLogin := makeRequest(r, "POST", "/api/auth/login", loginBody, "")

	var loginResponse map[string]interface{}
	json.Unmarshal(wLogin.Body.Bytes(), &loginResponse)
	token := loginResponse["token"].(string)

	wMe := makeRequest(r, "GET", "/api/auth/me", nil, token)

	assert.Equal(t, http.StatusOK, wMe.Code)
	assert.Contains(t, wMe.Body.String(), "Вы получили доступ к защищенным данным")
}

// 7. Тест доступа с поддельным (невалидным) токеном
func TestProtectedRouteInvalidToken(t *testing.T) {
	r := setupTestRouter()

	w := makeRequest(r, "GET", "/api/auth/me", nil, "fake.token.string")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Невалидный или истекший токен")
}
