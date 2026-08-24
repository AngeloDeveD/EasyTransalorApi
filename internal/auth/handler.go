package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Repo UserRepository
	Jwt  *JWTManager
}

func NewAuthHandler(repo UserRepository, jwt *JWTManager) *AuthHandler {
	return &AuthHandler{Repo: repo, Jwt: jwt}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера при хэшировании"})
		return
	}

	user := &User{
		FirstName:      req.FirstName,
		SecondName:     req.LastName,
		Nickname:       req.Nickname,
		PasswordHash:   string(hashedPassword),
		Role:           "author",
		RegistrationIP: c.ClientIP(),
		LastLoginIP:    c.ClientIP(),
	}

	if err := h.Repo.CreateUser(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Пользователь успешно зарегистрирован",
		"userId":  user.ID,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	user, err := h.Repo.GetUserByNickname(req.Nickname)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный никнейм или пароль"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный никнейм или пароль"})
		return
	}

	if user.IsBlocked {
		c.JSON(http.StatusForbidden, gin.H{"error": "Аккаунт заблокирован"})
		return
	}

	_ = h.Repo.UpdateLastLoginInfo(user.ID, c.ClientIP())

	token, err := h.Jwt.GenerateToken(user.ID, user.Role)
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

func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	role, _ := c.Get("role")

	c.JSON(http.StatusOK, gin.H{
		"message": "Вы получили доступ к защищенным данным!",
		"userId":  userID,
		"role":    role,
	})
}
