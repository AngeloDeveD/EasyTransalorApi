package game

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"myapi/internal/files"
)

// Структура хэндлера
type GameHandler struct {
	Repo     GameRepository
	FileRepo files.FileRepository
}

// Конструктор для создания хэндлера
func NewGameHandler(repo GameRepository, fileRepo files.FileRepository) *GameHandler {
	return &GameHandler{Repo: repo, FileRepo: fileRepo}
}

// Проверка id перевода и игры
// Проверяет на отсутвие символов и на то, чтобы в id были только числа
func CheckGameId(gameId string) (int, error) {
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

	c.JSON(http.StatusAccepted, games)
}

//GET /games/:gameid
/*Получение полной информации об игре по id*/
func (h *GameHandler) GetGameById(c *gin.Context) {

	//получение параметров с url
	gameid, err := CheckGameId(c.Param("gameid"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	game, err := h.Repo.GetGameInfoById(gameid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, game)
}

// GET /download/:gameid/:translid
/*Устанавливаем архив файла с переводом*/
// GET /download/:gameid/:translid
/*Устанавливаем архив файла с переводом*/
func (h *GameHandler) DownloadGameTranslation(c *gin.Context) {
	gameId, err := CheckGameId(c.Param("gameid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transid, err := CheckGameId(c.Param("translid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.Repo.CheckCreatedGame(gameId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Игра не найдена!"})
		return
	}

	gameInfo, err := h.Repo.GetGameInfoById(gameId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var fileUrl string
	var author string
	found := false

	for _, card := range gameInfo.TranslateCards {
		if card.ID == transid {
			if card.Status != "approved" {
				// Получениие роли пользователя
				role, exists := c.Get("role")
				if !exists || role != "admin" {
					c.JSON(http.StatusForbidden, gin.H{"error": "Доступ закрыт. Файл находится на модерации."})
					return
				}
			}
			fileUrl = card.UrlToDownload
			author = card.AuthorName
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Перевод не найден"})
		return
	}

	if fileUrl != "" {
		filePath := strings.Replace(fileUrl, "/static/", "uploads/", 1)

		rawName := gameInfo.Title + "_" + author + filepath.Ext(fileUrl)
		filename := strings.ReplaceAll(rawName, " ", "_")
		filename = strings.ReplaceAll(filename, "+", "_")

		fmt.Println("Попытка отдать файл:", filePath)
		c.FileAttachment(filePath, filename)
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"error": "URL файла пуст!"})
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
	gameid, err := CheckGameId(c.Param("gameid"))

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

	file_path, err := h.FileRepo.SaveArchive(gameid, file, c)

	//Для url-скачивания: /static/files/"gameid"
	file_url := strings.Replace(file_path, "uploads", "static", 1)

	//Для windows
	//file_url = strings.ReplaceAll(file_url, `\`, "/")

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
		Status:        "pending",
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

	//Удаление файлов с диска

	//Удаление архвов
	folderPath := filepath.Join("uploads", "files", strconv.Itoa(gameId))
	_ = os.RemoveAll(folderPath) //Если будет ошибка, то можно заигнорить

	//Удаление иконок
	if gameInfo.IconUrl != "" {
		//Замена "/static/Icons" на "uploads/Icons"
		filePathBig := strings.Replace(gameInfo.IconUrl, "/static/", "uploads/", 1)
		filePathSmall := strings.Replace(filePathBig, "/Big/", "/Small/", 1)
		_ = os.Remove(filePathBig)
		_ = os.Remove(filePathSmall)
	}

	c.JSON(http.StatusOK, gin.H{"error": "Игра вместе с переводами была удалена"})
}

//DELETE /games/translate/:gameid/:transid
/*Удаление перевода из игры*/
func (h *GameHandler) DeleteTranslation(c *gin.Context) {

	gameId, err := CheckGameId(c.Param("gameid"))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	translationId, err := CheckGameId("transid")

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Получение информации об игре для получения перевода и его пути к архиву
	gameInfo, err := h.Repo.GetGameInfoById(gameId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Игра не найдена"})
		return
	}

	var fileUrl string
	found := false
	for _, card := range gameInfo.TranslateCards {
		if card.ID == translationId {
			fileUrl = card.UrlToDownload
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Перевод не найден"})
		return
	}

	//Удаление записи из БД
	if err := h.Repo.DeleteTranslation(gameId, translationId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Удаление файла из диска
	if fileUrl != "" {
		filePath := strings.Replace(fileUrl, "/static/", "uploads/", 1)
		_ = os.Remove(filePath)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Перевод успешно удалён"})
}
