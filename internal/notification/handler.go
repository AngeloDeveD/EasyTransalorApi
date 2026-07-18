package notification

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	Repo NotificationRepository
}

func NewNotificationHandler(repo NotificationRepository) *NotificationHandler {
	return &NotificationHandler{Repo: repo}
}

type CreateNotificationRequest struct {
	Title    string `json:"title" binding:"required"`
	Message  string `json:"message" binding:"required"`
	IsGlobal bool   `json:"isGlobal"`
	UserID   int    `json:"userId"`
}

// POST /api/admin/notifications
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Заполните title и message"})
		return
	}

	n := &Notification{
		Title:   req.Title,
		Message: req.Message,
	}

	if req.IsGlobal {
		n.UserID = 0 //Глобальная отправка всем пользователям
	} else {
		if req.UserID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите userId или установите isGlobal: true"})
			return
		}
		n.UserID = req.UserID //Личное уведомление
	}

	if err := h.Repo.Create(n); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания уведомления"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Уведомление успещно создано"})
}

// GET api/notifications
func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	userID, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	notifications, err := h.Repo.GetForUser(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения уведомлений"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}
