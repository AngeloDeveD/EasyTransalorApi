package game

import (
	"net/http"
	"strconv"

	"myapi/internal/auth"
	"myapi/internal/notification"

	"github.com/gin-gonic/gin"
)

type ModerationHandler struct {
	GameRepo  GameRepository
	UserRepo  auth.UserRepository
	NotifRepo notification.NotificationRepository
}

func NewModerationHandler(gameRepo GameRepository, userRepo auth.UserRepository, notifRepo notification.NotificationRepository) *ModerationHandler {
	return &ModerationHandler{GameRepo: gameRepo, UserRepo: userRepo, NotifRepo: notifRepo}
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

	//TODO: отправка сообщению пользователю по ID
	/*
		authorID := h.GameRepo.GetTranslationAuthorID(transId)
		notif := &notification.Notification{
			UserID: authorID,
			Title: "Ваше перевод был отклонён",
			Message: "Ваш перевод (ID: " + transIdStr + ") был отклонён. Причина: " + req.Reason,
		}
		h.NotifRepo.Create(notif)
	*/
	c.JSON(http.StatusOK, gin.H{"message": "Перевод отклонён"})
}
