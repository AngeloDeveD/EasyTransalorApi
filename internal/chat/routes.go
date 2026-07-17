package chat

import "github.com/gin-gonic/gin"

func SetupChatRouter(r *gin.Engine, h *ChatHandler, authMiddleware gin.HandlerFunc) {
	chatGroup := r.Group("/api/chat", authMiddleware)
	{
		chatGroup.GET("/ws", h.HandleChat)
	}
}
