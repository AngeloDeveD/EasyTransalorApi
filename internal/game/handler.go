package game

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

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
type PublicGameInfo struct {
	ID           int                        `json:"id"`
	Title        string                     `json:"title"`
	IconUrl      string                     `json:"iconUrl"`
	Translations []PublicTranslationSummary `json:"translations"`
}

type PublicTranslationSummary struct {
	ID           int       `json:"id"`
	AuthorName   string    `json:"authorName"`
	Source       string    `json:"source"`
	Version      float64   `json:"version"`
	PercentReady float64   `json:"percentReady"`
	FileSize     float64   `json:"fileSize"`
	CreatedAt    time.Time `json:"createdAt"`
	DownloadUrl  string    `json:"downloadUrl"`
}

type ArchiveHashCheckRequest struct {
	ArchiveHash string `json:"archiveHash" binding:"required"`
}

type TranslationStatusResponse struct {
	ID          int              `json:"id"`
	Status      string           `json:"status"`
	ScanDetails string           `json:"scanDetails"`
	Files       []PublicGameFile `json:"files"`
}

type MyTranslationResponse struct {
	ID           int       `json:"id"`
	GameInfoID   int       `json:"gameInfoId"`
	AuthorName   string    `json:"authorName"`
	Source       string    `json:"source"`
	Version      float64   `json:"version"`
	PercentReady float64   `json:"percentReady"`
	FileSize     float64   `json:"fileSize"`
	Status       string    `json:"status"`
	ScanDetails  string    `json:"scanDetails"`
	CreatedAt    time.Time `json:"createdAt"`
	DownloadUrl  string    `json:"downloadUrl,omitempty"`
}

type PublicGameFile struct {
	FileName string `json:"fileName"`
	Size     string `json:"size"`
}
type GameHandler struct {
	Repo     GameRepository
	FileRepo files.FileRepository
	Scanner  ScanTrigger // может быть nil (напр. в тестах) — тогда проверка не запускается
}

// Конструктор для создания хэндлера
func NewGameHandler(repo GameRepository, fileRepo files.FileRepository, scanner ScanTrigger) *GameHandler {
	return &GameHandler{Repo: repo, FileRepo: fileRepo, Scanner: scanner}
}

func toPublicGameInfo(game GameInfo) PublicGameInfo {
	translations := make([]PublicTranslationSummary, 0, len(game.TranslateCards))
	for _, card := range game.TranslateCards {
		if card.Status != "approved" {
			continue
		}
		translations = append(translations, PublicTranslationSummary{
			ID:           card.ID,
			AuthorName:   card.AuthorName,
			Source:       card.Source,
			Version:      card.Version,
			PercentReady: card.PercentReady,
			FileSize:     card.FileSize,
			CreatedAt:    card.CreatedAt,
			DownloadUrl:  "/download/" + strconv.Itoa(card.ID),
		})
	}

	return PublicGameInfo{
		ID:           game.ID,
		Title:        game.Title,
		IconUrl:      game.IconUrl,
		Translations: translations,
	}
}

func toPublicGameFiles(files []DetailedGameFiles) []PublicGameFile {
	publicFiles := make([]PublicGameFile, 0, len(files))
	for _, file := range files {
		publicFiles = append(publicFiles, PublicGameFile{
			FileName: file.FileName,
			Size:     file.Size,
		})
	}
	return publicFiles
}
func toMyTranslationResponse(card TranslateCard) MyTranslationResponse {
	resp := MyTranslationResponse{
		ID:           card.ID,
		GameInfoID:   card.GameInfoID,
		AuthorName:   card.AuthorName,
		Source:       card.Source,
		Version:      card.Version,
		PercentReady: card.PercentReady,
		FileSize:     card.FileSize,
		Status:       card.Status,
		ScanDetails:  card.ScanDetails,
		CreatedAt:    card.CreatedAt,
	}
	if card.Status == "approved" {
		resp.DownloadUrl = "/download/" + strconv.Itoa(card.ID)
	}
	return resp
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

	publicGames := make([]PublicGameInfo, 0, len(games))
	for _, game := range games {
		publicGames = append(publicGames, toPublicGameInfo(game))
	}

	c.JSON(http.StatusAccepted, publicGames)
}

//GET /games/:gameid
/*Получение полной информации об игре по id*/
func (h *GameHandler) GetGameById(c *gin.Context) {
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

	c.JSON(http.StatusAccepted, toPublicGameInfo(game))
}

// GET /translations/:transid/files
func (h *GameHandler) GetTranslationFiles(c *gin.Context) {
	transid, err := CheckGameIdForNumValue(c.Param("transid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := h.Repo.GetTranslationByID(transid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Перевод не найден"})
		return
	}
	if card.Status != "approved" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Список файлов доступен только для одобренных переводов"})
		return
	}
	if card.GameFiles == nil {
		card.GameFiles = []DetailedGameFiles{}
	}

	c.JSON(http.StatusOK, toPublicGameFiles(card.GameFiles))
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

func (h *GameHandler) HashCheckArchive(c *gin.Context) {
	var req ArchiveHashCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите archiveHash"})
		return
	}

	archiveHash := strings.ToLower(strings.TrimSpace(req.ArchiveHash))
	if len(archiveHash) != 64 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "archiveHash должен быть SHA-256 строкой длиной 64 символа"})
		return
	}

	exists, err := h.Repo.ArchiveHashExists(archiveHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось проверить hash архива"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

func (h *GameHandler) GetTranslationStatus(c *gin.Context) {
	transid, err := CheckGameIdForNumValue(c.Param("transid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := h.Repo.GetTranslationByID(transid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Перевод не найден"})
		return
	}

	userID, _ := c.Get("userID")
	role, _ := c.Get("role")
	if card.AuthorId != userID.(int) && role != "moderator" && role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
		return
	}

	files := []PublicGameFile{}
	if card.Status == "approved" {
		files = toPublicGameFiles(card.GameFiles)
	}

	c.JSON(http.StatusOK, TranslationStatusResponse{
		ID:          card.ID,
		Status:      card.Status,
		ScanDetails: card.ScanDetails,
		Files:       files,
	})
}

func (h *GameHandler) GetMyTranslations(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	cards, err := h.Repo.GetTranslationsByAuthorID(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить переводы пользователя"})
		return
	}

	items := make([]MyTranslationResponse, 0, len(cards))
	for _, card := range cards {
		items = append(items, toMyTranslationResponse(card))
	}

	c.JSON(http.StatusOK, items)
}

func (h *GameHandler) DeleteMyTranslation(c *gin.Context) {
	transid, err := CheckGameIdForNumValue(c.Param("transid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := h.Repo.GetTranslationByID(transid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Перевод не найден"})
		return
	}

	userID, _ := c.Get("userID")
	if card.AuthorId != userID.(int) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Можно удалить только свой перевод"})
		return
	}
	if card.Status != "pending_scan" && card.Status != "rejected" && card.Status != "error" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Можно удалить только перевод на проверке или отклонённый перевод"})
		return
	}

	if err := h.Repo.DeleteTranslation(card.GameInfoID, card.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if card.UrlToDownload != "" {
		_ = h.FileRepo.DeleteFile(card.UrlToDownload)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Перевод удалён"})
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

	archiveHash, err := files.CalculateSHA256(files.StaticURLToPath(file_url))
	if err != nil {
		_ = h.FileRepo.DeleteFile(file_url)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось рассчитать hash архива"})
		return
	}

	duplicateExists, err := h.Repo.ArchiveHashExists(archiveHash)
	if err != nil {
		_ = h.FileRepo.DeleteFile(file_url)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось проверить дубликат архива"})
		return
	}
	if duplicateExists {
		_ = h.FileRepo.DeleteFile(file_url)
		c.JSON(http.StatusConflict, gin.H{"error": "Такой архив перевода уже был загружен"})
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
		ArchiveHash:   archiveHash,
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
