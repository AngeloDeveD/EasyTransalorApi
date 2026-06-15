package game

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Структура хэндлера
type GameHandler struct {
	Repo GameRepository
}

// Конструктор для создания хэндлера
func NewGameHanlder(repo GameRepository) *GameHandler {
	return &GameHandler{Repo: repo}
}

// Хэндлер
func (h *GameHandler) GetCards(c *gin.Context) {
	cards, err := h.Repo.GetAllCards()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить карточки"})
	}

	c.JSON(http.StatusAccepted, cards)
}

//ну и тд под get запросы
