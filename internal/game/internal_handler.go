package game

import (
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
	TransID int                 `json:"transId" binding:"required"`
	Status  string              `json:"status" binding:"required"`
	Details string              `json:"details"`
	Threats []string            `json:"threats"`
	Error   string              `json:"error"`
	Files   []DetailedGameFiles `json:"files"`
}

func (r ScanResultRequest) scanDetails() string {
	if r.Details != "" {
		return r.Details
	}
	if len(r.Threats) > 0 {
		return strings.Join(r.Threats, "; ")
	}
	return r.Error
}

// POST /api/internal/scan-result
func (h *InternalHandler) ReceiveScanResult(c *gin.Context) {
	var req ScanResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	details := req.scanDetails()

	switch req.Status {
	case "rejected":
		if err := h.GameRepo.UpdateScanResult(req.TransID, req.Status, details); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Достаем перевод, чтобы узнать автора и путь к файлу
		translation, err := h.GameRepo.GetTranslationByID(req.TransID)
		if err == nil {
			//Отправка уведомления
			notif := &notification.Notification{
				UserID:  translation.AuthorId,
				Title:   "В архиве обнаружен вирус!",
				Message: "Ваш перевод отклонен антивирусной проверкой. Детали: " + details,
			}
			h.NotifRepo.Create(notif)

			// Автоматическа выдача варн пользователю
			_ = h.UserRepo.WarnUser(translation.AuthorId, "Загрузка файла с вирусом: "+details)

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
	case "error":
		if err := h.GameRepo.UpdateScanResult(req.TransID, req.Status, details); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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
	case "approved":
		translation, err := h.GameRepo.GetTranslationByID(req.TransID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := h.GameRepo.UpdateFileInfo(translation.ID, req.Files); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := h.GameRepo.UpdateScanResult(req.TransID, req.Status, details); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	default:
		if err := h.GameRepo.UpdateScanResult(req.TransID, req.Status, details); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Результат сканирования принят"})
}
