package game

import (
	"log/slog"
	"net/http"
	"strconv"

	"myapi/internal/audit"
	"myapi/internal/auth"
	"myapi/internal/notification"

	"github.com/gin-gonic/gin"
)

type ModerationHandler struct {
	GameRepo  GameRepository
	UserRepo  auth.UserRepository
	NotifRepo notification.NotificationRepository
	AuditRepo audit.Repository
}

func NewModerationHandler(gameRepo GameRepository, userRepo auth.UserRepository, notifRepo notification.NotificationRepository, auditRepos ...audit.Repository) *ModerationHandler {
	auditRepo := audit.Repository(audit.NoopRepository{})
	if len(auditRepos) > 0 && auditRepos[0] != nil {
		auditRepo = auditRepos[0]
	}
	return &ModerationHandler{GameRepo: gameRepo, UserRepo: userRepo, NotifRepo: notifRepo, AuditRepo: auditRepo}
}

func (h *ModerationHandler) recordAudit(c *gin.Context, action string, targetID int, details string) {
	if err := h.AuditRepo.Create(audit.NewEventFromContext(c, action, audit.TargetTranslation, targetID, details)); err != nil {
		slog.ErrorContext(c.Request.Context(), "audit log failed", slog.String("action", action), slog.Int("target_id", targetID), slog.Any("error", err))
	}
}

// GET /api/admin/moderation?page=1
func (h *ModerationHandler) GetQueue(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 100 {
		limit = 20
	}
	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit

	translations, total, err := h.GameRepo.GetModerationQueue(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получении очереди"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  translations,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// PATCH /api/admin/moderation/:transid/change-status/:status
func (h *ModerationHandler) ChangeStatus(c *gin.Context) {
	transId, err := strconv.Atoi(c.Param("transid"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := c.Param("status")

	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Статус не найден!"})
		return
	}

	if err := h.GameRepo.ChangeStatusTranslation(transId, status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.recordAudit(c, "translation.change_status", transId, status)
	c.JSON(http.StatusOK, gin.H{"message": "Статус перевода изменён"})
}

// PATCH /api/admin/moderation/:transid/approve
func (h *ModerationHandler) Approve(c *gin.Context) {
	transIdStr := c.Param("transid")
	transId, err := strconv.Atoi(transIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.GameRepo.ApproveTranslation(transId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.recordAudit(c, "translation.approve", transId, "")
	c.JSON(http.StatusOK, gin.H{"message": "Перевод одобрен и опубликован"})
}

// PATCH /api/admin/moderation/:transid/reject
func (h *ModerationHandler) Reject(c *gin.Context) {
	transIdStr := c.Param("transid")
	transId, err := strconv.Atoi(transIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный id перевода"})
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите причину откланения"})
		return
	}

	//отклонение перевода
	if err := h.GameRepo.RejectTranslation(transId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	translation, err := h.GameRepo.GetTranslationByID(transId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось найти перевод для отправки уведомления"})
		return
	}

	notif := &notification.Notification{
		UserID:  translation.AuthorId,
		Title:   "Ваше перевод был отклонён",
		Message: "Ваш перевод (ID: " + transIdStr + ") был отклонён. Причина: " + req.Reason,
	}
	h.NotifRepo.Create(notif)
	h.recordAudit(c, "translation.reject", transId, req.Reason)
	c.JSON(http.StatusOK, gin.H{"message": "Перевод отклонён. Автору отправлено уведомление."})
}
