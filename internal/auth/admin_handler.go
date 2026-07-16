package auth

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	Repo UserRepository
}

func NewAdminHandler(repo UserRepository) *AdminHandler {
	return &AdminHandler{Repo: repo}
}

// Структура для принятия варна
type WarnRequest struct {
	Reason string `json:"reason" binding:"required"`
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

	//TODO: GetUserByID, далее проверка на 3 варна, и если true -> BlockUser()

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

	c.JSON(http.StatusOK, gin.H{"message": "Последний варн снят"})
}
