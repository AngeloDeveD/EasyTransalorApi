package auth

import (
	"log/slog"
	"myapi/internal/audit"
	"myapi/internal/notification"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	Repo      UserRepository
	NotifRepo notification.NotificationRepository
	AuditRepo audit.Repository
}

func NewAdminHandler(repo UserRepository, notifRepo notification.NotificationRepository, auditRepos ...audit.Repository) *AdminHandler {
	auditRepo := audit.Repository(audit.NoopRepository{})
	if len(auditRepos) > 0 && auditRepos[0] != nil {
		auditRepo = auditRepos[0]
	}
	return &AdminHandler{Repo: repo, NotifRepo: notifRepo, AuditRepo: auditRepo}
}

func (h *AdminHandler) recordAudit(c *gin.Context, action string, targetID int, details string) {
	if err := h.AuditRepo.Create(audit.NewEventFromContext(c, action, audit.TargetUser, targetID, details)); err != nil {
		slog.ErrorContext(c.Request.Context(), "audit log failed", slog.String("action", action), slog.Int("target_id", targetID), slog.Any("error", err))
	}
}

// Структура для принятия варна
type WarnRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// Структура для смены роли пользователя
type SetRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

//GET /api/admin/users?page=1&limit=20
/*Получение списка пользователя*/
func (h *AdminHandler) GetUsers(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	users, totalCount, err := h.Repo.GetUsers(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения пользователей"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       users,
		"total":      totalCount,
		"page":       page,
		"limit":      limit,
		"totalPages": (totalCount + int64(limit) - 1) / int64(limit),
	})
}

func (h *AdminHandler) GetAuditEvents(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 50
	}

	offset := (page - 1) * limit
	events, totalCount, err := h.AuditRepo.List(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       events,
		"total":      totalCount,
		"page":       page,
		"limit":      limit,
		"totalPages": (totalCount + int64(limit) - 1) / int64(limit),
	})
}

//PATCH /api/admin/users/:userid/block
/*Блокирует пользователя*/
func (h *AdminHandler) BlockUser(c *gin.Context) {
	idStr := c.Param("userid")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	if err := h.Repo.BlockUser(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.recordAudit(c, "user.block", id, "")
	c.JSON(http.StatusOK, gin.H{"message": "Пользователь заблокирован"})
}

//PATCH /api/admin/users/:userid/unblock
/*Разблокирует пользователя*/
func (h *AdminHandler) UnblockUser(c *gin.Context) {
	idStr := c.Param("userid")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	if err := h.Repo.UnblockUser(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.recordAudit(c, "user.unblock", id, "")
	c.JSON(http.StatusOK, gin.H{"message": "Пользователь разблокирован"})
}

//PATCH /api/admin/users/:userid/warn
/*Выдаёт пользователю варн*/
func (h *AdminHandler) WarnUser(c *gin.Context) {
	idStr := c.Param("userid")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	var req WarnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите причину варна"})
		return
	}

	if err := h.Repo.WarnUser(id, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка выдачи варна"})
		return
	}

	//Отправка сообщения о варне
	notif := &notification.Notification{
		UserID:  id,
		Title:   "Вам выдано предупреждение",
		Message: "Причина: " + req.Reason,
	}
	h.NotifRepo.Create(notif)
	h.recordAudit(c, "user.warn", id, req.Reason)

	//проверка кол-ва варнов. Если больше трёх - бан
	user, err := h.Repo.GetUserById(id)
	if err == nil && user.WarnCount >= 3 {
		h.Repo.BlockUser(id)
		h.recordAudit(c, "user.auto_block_after_warnings", id, req.Reason)

		//Уведомление пользователя о блокировке
		banNotif := &notification.Notification{
			UserID:  id,
			Title:   "Ваш аккаунт заблокирован",
			Message: "Ваш аккаунт заблокирован за нарушения. Теперь вам можно только устанавливать файлы",
		}
		h.NotifRepo.Create(banNotif)
		c.JSON(http.StatusOK, gin.H{"message": "Варн выдан. Пользователь заблокирован за 3 нарушения."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Варн успешно выдан"})
}

//PATCH /api/admin/users/:userid/unwarn
/*Убирает у пользователя варн*/
func (h *AdminHandler) UnwarnUser(c *gin.Context) {
	idStr := c.Param("userid")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	if err := h.Repo.UnwarnUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.recordAudit(c, "user.unwarn", id, "")
	c.JSON(http.StatusOK, gin.H{"message": "Последний варн снят"})
}

//PATCH /api/admin/users/:userid/role
/*Меняет роль пользователя. Доступно ТОЛЬКО админу (маршрут закрыт AdminMiddleware),
модераторы сюда не допускаются — назначение ролей остаётся эксклюзивом создателя.*/
func (h *AdminHandler) SetRole(c *gin.Context) {
	idStr := c.Param("userid")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	var req SetRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите роль (author, moderator или admin)"})
		return
	}

	if err := h.Repo.SetRole(id, req.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Уведомление пользователя о смене роли
	notif := &notification.Notification{
		UserID:  id,
		Title:   "Ваша роль изменена",
		Message: "Администратор назначил вам роль: " + req.Role,
	}
	h.NotifRepo.Create(notif)
	h.recordAudit(c, "user.set_role", id, req.Role)

	c.JSON(http.StatusOK, gin.H{"message": "Роль пользователя обновлена"})
}
