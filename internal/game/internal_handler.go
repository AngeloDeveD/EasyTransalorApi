package game

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"myapi/internal/auth"
	"myapi/internal/files"
	"myapi/internal/notification"

	"github.com/gin-gonic/gin"
)

type InternalHandler struct {
	GameRepo  GameRepository
	UserRepo  auth.UserRepository
	FileRepo  files.FileRepository
	NotifRepo notification.NotificationRepository
}

// Обновляем конструктор
func NewInternalHandler(gameRepo GameRepository, userRepo auth.UserRepository, fileRepo files.FileRepository, notifRepo notification.NotificationRepository) *InternalHandler {
	return &InternalHandler{
		GameRepo:  gameRepo,
		UserRepo:  userRepo,
		FileRepo:  fileRepo,
		NotifRepo: notifRepo,
	}
}

type ScanResultRequest struct {
	TransID int    `json:"transId" binding:"required"`
	Status  string `json:"status" binding:"required"`
	Details string `json:"details"`
}

// POST /api/internal/scan-result
func (h *InternalHandler) ReceiveScanResult(c *gin.Context) {
	var req ScanResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	if err := h.GameRepo.UpdateScanResult(req.TransID, req.Status, req.Details); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Status == "rejected_by_scanner" {

		//Проверка на статус самого http
		if c.Writer.Status() != http.StatusAccepted {
			err := fmt.Sprintf("Ошибка: %s", req.Details)

			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		// Достаем перевод, чтобы узнать автора и путь к файлу
		translation, err := h.GameRepo.GetTranslationByID(req.TransID)
		if err == nil {
			//Отправка уведомления
			notif := &notification.Notification{
				UserID:  translation.AuthorId,
				Title:   "В архиве обнаружен вирус!",
				Message: "Ваш перевод отклонен антивирусной проверкой. Детали: " + req.Details,
			}
			h.NotifRepo.Create(notif)

			// Автоматическа выдача варн пользователю
			_ = h.UserRepo.WarnUser(translation.AuthorId, "Загрузка файла с вирусом: "+req.Details)

			//Проверка, не превысил ли лимит (3 варна = бан)
			user, err := h.UserRepo.GetUserById(translation.AuthorId)
			if err == nil && user.WarnCount >= 3 {
				_ = h.UserRepo.BlockUser(translation.AuthorId)

				banNotif := &notification.Notification{
					UserID:  translation.AuthorId,
					Title:   "Аккаунт заблокирован",
					Message: "Вы получили 3 предупреждения за нарушение правил.",
				}
				h.NotifRepo.Create(banNotif)
			}

			// Удаление файла с диска
			if translation.UrlToDownload != "" {
				filePath := strings.Replace(translation.UrlToDownload, "/static/", "uploads/", 1)
				_ = os.Remove(filePath) // Удаляем файл, ошибку игнорируем (если файла уже нет)
			}
		}
	} else if req.Status == "pending_sandbox" {
		translation, err := h.GameRepo.GetTranslationByID(req.TransID)
		if err == nil {
			//Отправка уведомления
			notif := &notification.Notification{
				UserID:  translation.AuthorId,
				Title:   "Файл был отправлен на глубокую проверку!",
				Message: "Ваш перевод был отправлен на более глубокую проверку",
			}
			h.NotifRepo.Create(notif)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Результат сканирования принят"})
}
