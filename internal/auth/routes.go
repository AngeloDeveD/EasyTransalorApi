package auth

import "github.com/gin-gonic/gin"

func SetupAuthRoutes(r *gin.Engine, h *AuthHandler, middlewares ...gin.HandlerFunc) {
	authGroup := r.Group("/api/auth")
	{
		registerHandlers := append([]gin.HandlerFunc{}, middlewares...)
		registerHandlers = append(registerHandlers, h.Register)
		authGroup.POST("/register", registerHandlers...)

		loginHandlers := append([]gin.HandlerFunc{}, middlewares...)
		loginHandlers = append(loginHandlers, h.Login)
		authGroup.POST("/login", loginHandlers...)

		authGroup.GET("/me", AuthMiddleware(h.Jwt), h.Me)
	}
}
