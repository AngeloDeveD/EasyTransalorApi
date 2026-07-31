package auth

import "github.com/gin-gonic/gin"

// SetupAdminRoutes регистрирует маршруты админки.
// Действия уровня «почти как админ» (просмотр, блокировки, предупреждения)
// доступны модераторам — их закрывает moderatorMiddleware.
// Назначение ролей остаётся эксклюзивом создателя-админа и закрыто adminMiddleware.
func SetupAdminRoutes(r *gin.Engine, h *AdminHandler, authMiddleware gin.HandlerFunc, moderatorMiddleware gin.HandlerFunc, adminMiddleware gin.HandlerFunc) {
	//Группа для модераторов и админов
	modGroup := r.Group("/api/admin", authMiddleware, moderatorMiddleware)
	{
		modGroup.GET("/users", h.GetUsers)
		modGroup.PATCH("/users/:userid/block", h.BlockUser)
		modGroup.PATCH("/users/:userid/unblock", h.UnblockUser)
		modGroup.PATCH("/users/:userid/warn", h.WarnUser)
		modGroup.PATCH("/users/:userid/unwarn", h.UnwarnUser)
	}

	//Только для создателя-админа: назначение ролей другим пользователям
	adminOnly := r.Group("/api/admin", authMiddleware, adminMiddleware)
	{
		adminOnly.PATCH("/users/:userid/role", h.SetRole)
	}
}
