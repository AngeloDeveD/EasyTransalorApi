package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Хранение зависимости для авторизации
type AuthHandler struct {
	Repo UserRepository
	Jwt  *JwtManager
}

func NewAuthHandler(repo UserRepository, jwt *JwtManager) *AuthHandler {
	return &AuthHandler{Repo: repo, Jwt: jwt}
}

// Регистрация
// POST: /register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	//Хэширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера при хэшировании"})
		return
	}

	//Создание пользователя
	user := &User{
		FirstName:    req.FirstName,
		SecondName:   req.LastName,
		Nickname:     req.Nickname,
		PasswordHash: string(hashedPassword),
		Role:         "author", // по умолчанию все пользователи имеют роль "author"
	}

	//Сохранение в БД
	if err := h.Repo.CreateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Пользотваль успешно зарегестрировани",
		"userId":  user.ID,
	})
}

// Вход
// POST: /login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	//Поиск пользователя по никнейму
	user, err := h.Repo.GetUserByNickname(req.Nickname)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
		return
	}

	//Проверка пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
		return
	}

	//Создание jwt-токена
	token, err := h.Jwt.GenerateToken(int64(user.ID), user.Role)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании токена"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"nickname": user.Nickname,
			"role":     user.Role,
		},
	})
}
