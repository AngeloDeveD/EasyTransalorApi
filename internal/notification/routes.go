package notification

import "github.com/gin-gonic/gin"

func SetupNotificationRoutes(
	r *gin.Engine,
	h *NotificationHandler,
	authMiddleware gin.HandlerFunc,
	adminMiddleware gin.HandlerFunc,
) {
	userGroup := r.Group("/api/notifications", authMiddleware)
	{
		userGroup.GET("", h.GetMyNotifications)
	}

	adminGroupe := r.Group("/api/admin/notifications", authMiddleware, adminMiddleware)
	{
		adminGroupe.POST("", h.CreateNotification)
	}
}
