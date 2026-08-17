package game

import "github.com/gin-gonic/gin"

func SetupModerationRoutes(r *gin.Engine, h *ModerationHandler, authMiddleware gin.HandlerFunc, adminMiddleware gin.HandlerFunc) {
	modGroupe := r.Group("/api/admin/moderation", authMiddleware, adminMiddleware)
	{
		modGroupe.GET("", h.GetQueue)
		modGroupe.PATCH("/:transid/approve", h.Approve)
		modGroupe.PATCH("/:transid/reject", h.Reject)
		modGroupe.PATCH("/:transid/change-status/:status", h.ChangeStatus)
	}
}
