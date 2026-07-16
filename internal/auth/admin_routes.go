package auth

import "github.com/gin-gonic/gin"

func SetupAdminRoutes(r *gin.Engine, h *AdminHandler, authMiddleware gin.HandlerFunc, adminMiddleware gin.HandlerFunc) {
	adminGroup := r.Group("/api/admin", authMiddleware, adminMiddleware)
	{
		adminGroup.GET("/users", h.GetUsers)
		adminGroup.PATCH("/users/:userid/block", h.BlockUser)
		adminGroup.PATCH("/users/userid/unblock", h.UnblockUser)
		adminGroup.PATCH("/users/:userid/warn", h.WarnUser)
		adminGroup.PATCH("/users/:userid/unwarn", h.UnwarnUser)

	}
}
