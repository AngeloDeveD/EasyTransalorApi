package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwt *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""

		//Попытка взять токен из заголовка
		authHeader := c.GetHeader("Authorization")
		if len(strings.Split(authHeader, " ")) == 2 {
			tokenString = strings.Split(authHeader, " ")[1]
		} else {
			//Если в заголовке НЕТ токена, поиск его в URL (?token=...)
			//Это нужно для WebSocket, так как браузеры не могут отправить заголовок при WS-подключении
			tokenString = c.Query("token")
		}

		//Если токена вообще нигде нет - отфутболивание
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			c.Abort()
			return
		}

		// Проверка токена
		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Невалидный или истекший токен"})
			c.Abort()
			return
		}

		// Добавление данных юзера в контекст
		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права администратора."})
			c.Abort()
			return
		}
		c.Next()
	}
}

func APIKeyMiddleware(validKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Internal-Key")
		if key == "" || key != validKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный сервисный ключ"})
			c.Abort()
			return
		}
		c.Next()
	}
}
