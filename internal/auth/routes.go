package auth

import "github.com/gin-gonic/gin"

func SetupAuthRoutes(r *gin.Engine, h *AuthHandler) {
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)

		authGroup.GET("/me", AuthMiddleware(h.Jwt), h.Me)
	}
}
