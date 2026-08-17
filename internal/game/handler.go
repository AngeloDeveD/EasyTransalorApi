package game

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"myapi/internal/files"
)

// ScanTrigger отправляет загруженный архив на антивирусную проверку.
// Интерфейс держим в пакете game, чтобы не завязываться на пакет scanner
// напрямую (и легко подменять заглушкой в тестах).
type ScanTrigger interface {
	Scan(transID int, fileURL string)
}

// Структура хэндлера
type GameHandler struct {
	Repo     GameRepository
	FileRepo files.FileRepository
	Scanner  ScanTrigger // может быть nil (напр. в тестах) — тогда проверка не запускается
}

// Конструктор для создания хэндлера
func NewGameHandler(repo GameRepository, fileRepo files.FileRepository, scanner ScanTrigger) *GameHandler {
	return &GameHandler{Repo: repo, FileRepo: fileRepo, Scanner: scanner}
}

func normalizeGameFiles(games []GameInfo) {
	for i := range games {
		normalizeGameFile(&games[i])
	}
}

func normalizeGameFile(game *GameInfo) {
	if game.TranslateCards == nil {
		game.TranslateCards = []TranslateCard{}
	}
	for i := range game.TranslateCards {
		if game.TranslateCards[i].GameFiles == nil {
			game.TranslateCards[i].GameFiles = []DetailedGameFiles{}
		}
	}
}

// Проверка id перевода и игры
// Проверяет на отсутвие символов и на то, чтобы в id были только числа
func CheckGameIdForNumValue(gameId string) (int, error) {
	if len(gameId) == 0 {
		return -1, errors.New("Id игры не был получен")
	}
	gameid_int, err := strconv.Atoi(gameId)

	if err != nil {
		return -1, errors.New("Полученный id не является числом")
	}

	return gameid_int, nil
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

	normalizeGameFiles(games)
	c.JSON(http.StatusAccepted, games)
}

//GET /games/:gameid
/*Получение полной информации об игре по id*/
func (h *GameHandler) GetGameById(c *gin.Context) {

	//получение параметров с url
	gameid, err := CheckGameIdForNumValue(c.Param("gameid"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.Repo.GetGameInfoById(gameid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	normalizeGameFile(&game)
	c.JSON(http.StatusAccepted, game)
}

// GET /download/:transid
/*Устанавливаем архив файла с переводом*/
func (h *GameHandler) DownloadGameTranslation(c *gin.Context) {
	transid, err := CheckGameIdForNumValue(c.Param("transid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID перевода"})
		return
	}

	card, err := h.Repo.GetTranslationByID(transid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Перевод не найден"})
		return
	}

	if card.Status != "approved" {
		role, exist := c.Get("role")
		if !exist || role != "moderation" && role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ закрыт. Файл находится на модерации."})
			return
		}
	}

	gameInfo, err := h.Repo.GetGameInfoById(card.GameInfoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить данные об игре"})
		return
	}

	fileUrl := card.UrlToDownload
	author := card.AuthorName

	if fileUrl != "" {
		// Просим репозиторий дать нам путь к файлу
		filePath, filename, err := h.FileRepo.GetArchivePath(gameInfo.Title, author, fileUrl)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		fmt.Println("Попытка отдать файл:", filePath)
		// Официант (Handler) сам отдает файл клиенту
		c.FileAttachment(filePath, filename)
		return
	}
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

	//Проверка на поддерживаемые форматы
	if err := h.FileRepo.IsAllowedImageFormat(big_pic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.FileRepo.IsAllowedImageFormat(small_pic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Проверка размера изображений
	if err := h.FileRepo.IsAllowedImageSize(small_pic.Size, big_pic.Size); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Сохранения картинок
	image_url, err := h.FileRepo.SaveImages(big_pic, small_pic)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//Получение ссылки на маленькую иконку
	image_small_url := image_url[0]
	image_big_url := image_url[1]

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

	userID, exist := c.Get("userID")
	if !exist {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка получения текстовых данных"})
		return
	}

	//получение параметров с url
	gameid, err := CheckGameIdForNumValue(c.Param("gameid"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Получение форм с названием file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка получения файла"})
		return
	}

	foundGame := h.Repo.CheckCreatedGame(gameid)

	if foundGame != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": foundGame.Error()})
		return
	}

	if err := h.FileRepo.IsAllowedArchiveFormat(file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Проверка на максимальный размер файла до 5 гб
	if err := h.FileRepo.IsAllowedArchiveSize(file.Size); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Для url-скачивания: /static/files/"gameid"
	file_url, err := h.FileRepo.SaveArchive(gameid, file)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//Сначала узнаём размер файла в мегабайтах, а потом округляем до одного символа пары символов после точки
	sizeInMb := float64(file.Size) / (1024 * 1024)
	roundedSize := math.Round(sizeInMb*100) / 100

	trasnalteInfo := TranslateCard{
		ID:            int(uuid.New().ID()),
		AuthorName:    req.AuthorName,
		AuthorId:      userID.(int),
		Source:        req.Source,
		Version:       req.Version,
		PercentReady:  req.PercentReady,
		UrlToDownload: file_url,
		FileSize:      roundedSize,
		Status:        "pending_scan",
		GameFiles:     []DetailedGameFiles{},
	}

	saveTranslationInfo := h.Repo.AddTranslation(gameid, trasnalteInfo)

	if saveTranslationInfo != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": saveTranslationInfo.Error()})
		return
	}

	// Отправляем архив на антивирусную проверку в фоне.
	// Пока статус остаётся "pending_scan"; вердикт придёт на /api/internal/scan-result.
	if h.Scanner != nil {
		h.Scanner.Scan(trasnalteInfo.ID, file_url)
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

//DELETE /games/:gameid
/*Удаление игры*/
func (h *GameHandler) DeleteGame(c *gin.Context) {
	gameIdStr := c.Param("gameid")
	gameId, err := strconv.Atoi(gameIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Полученный id не является числом"})
		return
	}

	//Получение информации об игре из бд
	gameInfo, err := h.Repo.GetGameInfoById(gameId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Игра не найдена"})
		return
	}

	//Удаление игры из бд
	if err := h.Repo.DeleteGame(gameId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления игры"})
		return
	}

	_ = h.FileRepo.DeleteGameFiles(gameId)

	//Удаление иконок
	if gameInfo.IconUrl != "" {
		_ = h.FileRepo.DeleteImageFiles(gameInfo.IconUrl)
	}

	c.JSON(http.StatusOK, gin.H{"error": "Игра вместе с переводами была удалена"})
}

//DELETE /games/translate/:transid
/*Удаление перевода из игры*/
func (h *GameHandler) DeleteTranslation(c *gin.Context) {

	translationId, err := CheckGameIdForNumValue(c.Param("transid"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := h.Repo.GetTranslationByID(translationId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Перевод не найден"})
		return
	}

	//Удаление записи с бд
	if err := h.Repo.DeleteTranslation(card.GameInfoID, translationId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//Удаление файла с диска
	fileUrl := card.UrlToDownload

	if fileUrl != "" {
		_ = h.FileRepo.DeleteFile(fileUrl)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Перевод успешно удалён"})
}
