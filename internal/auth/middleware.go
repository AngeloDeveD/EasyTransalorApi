package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Роли пользователей в системе.
// author    — обычный пользователь (значение по умолчанию);
// moderator — почти как админ, но НЕ может назначать роли другим;
// admin      — создатель, может всё, включая назначение ролей.
const (
	RoleAuthor    = "author"
	RoleModerator = "moderator"
	RoleAdmin     = "admin"
)

func AuthMiddleware(jwt *JWTManager) gin.HandlerFunc {
	return authMiddleware(jwt, nil)
}

func AuthMiddlewareWithUserCheck(jwt *JWTManager, repo UserRepository) gin.HandlerFunc {
	return authMiddleware(jwt, repo)
}

func authMiddleware(jwt *JWTManager, repo UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""

		authHeader := c.GetHeader("Authorization")
		parts := strings.Fields(authHeader)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			tokenString = parts[1]
		} else if c.FullPath() == "/api/chat/ws" || c.Request.URL.Path == "/api/chat/ws" {
			// Query token оставлен только для WebSocket, где браузер не может удобно отправить Authorization header.
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			c.Abort()
			return
		}

		claims, err := jwt.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Невалидный или истекший токен"})
			c.Abort()
			return
		}

		if repo != nil {
			user, err := repo.GetUserById(claims.UserID)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
				c.Abort()
				return
			}
			if user.IsBlocked {
				c.JSON(http.StatusForbidden, gin.H{"error": "Аккаунт заблокирован"})
				c.Abort()
				return
			}
			claims.Role = user.Role
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права администратора."})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ModeratorMiddleware пропускает и модераторов, и админов.
// Используется для действий, которые доступны модератору («почти как админ»):
// модерация переводов, блокировки/предупреждения, управление играми, рассылка уведомлений.
// Назначение ролей другим пользователям сюда НЕ входит — оно остаётся за AdminMiddleware.
func ModeratorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || (role != RoleAdmin && role != RoleModerator) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен. Требуются права модератора или администратора."})
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
