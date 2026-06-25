package game

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Структура хэндлера
type GameHandler struct {
	Repo GameRepository
}

// Конструктор для создания хэндлера
func NewGameHandler(repo GameRepository) *GameHandler {
	return &GameHandler{Repo: repo}
}

// GET /cards
/*Получение карточки игры (Без информации о переводах)*/
func (h *GameHandler) GetCards(c *gin.Context) {
	cards, err := h.Repo.GetAllCards()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить карточки"})
		return
	}

	c.JSON(http.StatusAccepted, cards)
}

// GET /games
/*Получение информации об игре (С информацией о переводах)*/
func (h *GameHandler) GetGames(c *gin.Context) {
	games, err := h.Repo.GetAllGamesInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить игры"})
		return
	}

	c.JSON(http.StatusAccepted, games)
}

//GET /games/:gameid
/*Получение полной информации об игре по id*/
func (h *GameHandler) GetGameById(c *gin.Context) {

	//получение параметров с url
	gameId_str := c.Param("gameid")

	//Прверка на пустой gameid
	if len(gameId_str) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Id игры не был получен"})
		return
	}

	//Преобразование id в int64
	gameid, err := strconv.ParseInt(gameId_str, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Полученный id не является числом"})
		return
	}

	game, err := h.Repo.GetGameInfoById(gameid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, game)
}

//POST /games/add
/*Добавление карточки игры*/
func (h *GameHandler) AddGame(c *gin.Context) {
	//Привязка текстовых данных из форм
	var req CreateGameRequest

	//Если текстовые поля были не отпавлены
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Получение большой и маленькой картинки
	big_pic, err := c.FormFile("big_pic")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Большое изображение должно быть загружено!"})
		return
	}

	small_pic, err := c.FormFile("small_pic")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Маленькое изображение должно быть загружено!"})
		return
	}

	//Подерживаемые формамы
	allowExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}

	//Генерация уникального имени
	ext := strings.ToLower(filepath.Ext(big_pic.Filename))
	newPic := uuid.New().String() + ext

	//Проверка на формат файлов
	if !allowExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый формат файла!"})
		return
	}

	//Проверка на размер файлоы (максимум -> 5 мб)
	if big_pic.Size > 5<<20 || small_pic.Size > 5<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл слишком большой (макс 5 мб)"})
		return
	}

	//Пути сохранения для картинок
	savePathBig := filepath.Join("uploads/Icons/Big", newPic)
	savePathSmall := filepath.Join("uploads/Icons/Small", newPic)

	//Сохранение файлов на диск
	if err := c.SaveUploadedFile(big_pic, savePathBig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения большой иконки"})
		return
	}

	if err := c.SaveUploadedFile(small_pic, savePathSmall); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения маленькой иконки"})
		return
	}

	image_small_url := fmt.Sprintf("/static/Icons/Small/%s", newPic)
	image_big_url := fmt.Sprintf("/static/Icons/Big/%s", newPic)

	//Создание id для игры
	gameID := int(uuid.New().ID())

	//Создание данных о карточке игры
	card := GameCard{
		ID:      int(uuid.New().ID()),
		Title:   req.Title,
		IconUrl: image_small_url,
		GameId:  gameID,
	}

	//Создание полной информации об игре
	info := GameInfo{
		ID:             gameID,
		Title:          req.Title,
		IconUrl:        image_big_url,
		TranslateCards: []TranslateCard{},
	}

	err = h.Repo.CreateNewGame(card, info)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения данных"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":         "Успешно создано",
		"title":           req.Title,
		"big_image_url":   image_big_url,
		"small_image_url": image_small_url,
		"gameId":          gameID,
	})
}

//POST /games/translate/:gameid
/*Добавляет архив с переводом к игре*/
func (h *GameHandler) AddTranslationInfo(c *gin.Context) {
	var req CreateTraslateRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка получения текстовых данных"})
		return
	}

	//получение параметров с url
	gameId_str := c.Param("gameid")

	//Прверка на пустой gameid
	if len(gameId_str) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Id игры не был получен"})
		return
	}

	//Преобразование id в int64
	gameid, err := strconv.ParseInt(gameId_str, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Полученный id не является числом"})
		return
	}

	//Получение форм с названием file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка почучения файла"})
		return
	}

	//Подерживаемые формамы файлов
	allowFileExts := map[string]bool{".zip": true, ".7zip": true, ".rar": true}

	foundGame := h.Repo.CheckCreatedGame(gameid)

	if foundGame != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": foundGame.Error()})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	newFile := uuid.New().String() + ext

	if !allowFileExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый формат файла!"})
		return
	}

	//Проверка на максимальный размер файла до 5 гб
	if file.Size > 5<<30 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл слишком большой (макс 5 гб)"})
		return
	}

	// uploads/files/"gameid"
	folderPath := filepath.Join("uploads", "files", strconv.FormatInt(gameid, 10))

	// Убедимся, что папка существует! Иначе SaveUploadedFile упадет с ошибкой
	if err := os.MkdirAll(folderPath, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать директорию"})
		return
	}

	//Создание полного пути для сохранения файла
	savePathFile := filepath.Join(folderPath, newFile)

	if err := c.SaveUploadedFile(file, savePathFile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить перевод"})
		return
	}

	//Для url-скачивания: /static/files/"gameid"
	file_url := filepath.Join("/static", "files", strconv.FormatInt(gameid, 10), newFile)

	//Для windows
	file_url = strings.ReplaceAll(file_url, `\`, "/")

	//Сначала узнаём размер файла в мегабайтах, а потом округляем до одного символа пары символов после точки
	sizeInMb := float64(file.Size) / (1024 * 1024)
	roundedSize := math.Round(sizeInMb*100) / 100

	trasnalteInfo := TranslateCard{
		ID:            int(uuid.New().ID()),
		AuthorName:    req.AuthorName,
		Source:        req.Source,
		Version:       req.Version,
		PercentReady:  req.PercentReady,
		UrlToDownload: file_url,
		FileSize:      roundedSize,
	}

	saveTranslationInfo := h.Repo.AddTranslation(gameid, trasnalteInfo)

	if saveTranslationInfo != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": saveTranslationInfo.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Успешно создано",
		"urlToDownload": trasnalteInfo.UrlToDownload,
		"FileSize":      trasnalteInfo.FileSize,
		"AuthorName":    trasnalteInfo.AuthorName,
		"Source":        trasnalteInfo.Source,
		"PercentReady":  trasnalteInfo.PercentReady,
		"id":            trasnalteInfo.ID,
	})
}
