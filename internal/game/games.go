package game

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GameCard struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	IconUrl string `json:"iconUrl"`
	GameId  int `json:"gameId"`
}

type GameInfo struct {
	ID             int             `json:"id"`
	Title          string          `json:"title"`
	IconUrl        string          `json:"iconUrl"`
	TranslateCards []TranslateCard `json:"translateCards"`
}

type TranslateCard struct {
	ID            int     `json:"id"`
	AuthorName    string  `json:"authorName"`
	Source        string  `json:"source"`
	Version       float32 `json:"version"`
	PercentReady  float32 `json:"percentReady"`
	UrlToDownload string  `json:"urlToDownload"`
	FileSize float32 `json:"fileSize"`
}

var gameCard = []GameCard{
	{
		ID: 1,
		Title: "Игра номер 1",
		IconUrl: "source",
		GameId: 1,
	},
}

var gameInfo = []GameInfo{
	{
		ID:      1,
		Title:   "Игра номер 1",
		IconUrl: "Url1",
		TranslateCards: []TranslateCard{
			{
				ID:            1,
				AuthorName:    "Васька 1",
				Source:        "url",
				Version:       1.0,
				PercentReady:  0.0,
				UrlToDownload: "url",
				FileSize: 0.0,
			},
		},
	},
	{
		ID:      1,
		Title:   "Игра номер 1",
		IconUrl: "Url1",
		TranslateCards: []TranslateCard{
			{
				ID:            1,
				AuthorName:    "Васька 1",
				Source:        "url",
				Version:       1.0,
				PercentReady:  0.0,
				UrlToDownload: "url",
				FileSize: 0.0,
			},
		},
	},
}

type CreateGameRequest struct {
	Title string `json:"title"`
}

//GET: получение информации об игре (С информацией о переводах)
func getGame(c *gin.Context){
	c.JSON(http.StatusAccepted, gameInfo)
}

//GET: получение карточки игры (Без информации о переводах)
func getCard(c *gin.Context){
	c.JSON(http.StatusOK, gameCard)
}

//POST: добавление игры
//TODO: Написать тесты для проверки: разных форматов файлов (поддерживамые или нет), для отправки только одного изобрадение и для отправлки слишком большого файла
func addGame(c *gin.Context){
	//Привязка текстовых данных из формф
	var req CreateGameRequest

	if err := c.ShouldBind(&req); err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Получение большой и маленькой картинки
	big_pic, err := c.FormFile("big_pic")
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Большое изображение должно быть загружено!"})
		return
	}

	small_pic, err := c.FormFile("small_pic")
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Маленькое изображение должно быть загружено!"})
		return
	}

	//Подерживаемые формамы
	allowExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}

	//Генерация уникального имени
	ext := strings.ToLower(filepath.Ext(big_pic.Filename))
	newPic := uuid.New().String() + ext

	//Проверка на формат файлов
	if(!allowExts[ext]){
		c.JSON(http.StatusBadRequest, gin.H{"error": "Недопустимый формат файла!"})
		return
	}

	//Проверка на размер файлоы (максимум -> 5 мб)
	if big_pic.Size > 5<<20 || small_pic.Size > 5<<20{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл слишком большой (макс 5 мб)"})
		return
	}

	//Пути сохранения для картинок
	savePathBig := filepath.Join("uploads/Icons/Big", newPic)
	savePathSmall := filepath.Join("uploads/Icons/Small", newPic)

	//Сохранение файлов на диск
	if err := c.SaveUploadedFile(big_pic, savePathBig); err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения большой иконки"})
		return
	}

	if err := c.SaveUploadedFile(small_pic, savePathSmall); err != nil{
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения маленькой иконки"})
		return
	}

	image_small_url := fmt.Sprintf("/static/Icons/Small/%s", newPic)
	image_big_url := fmt.Sprintf("/static/Icons/Big/%s", newPic)

	info := GameCard{
		ID: int(uuid.New().ID()),
		Title: req.Title,
		IconUrl: image_small_url,
	}

	gameCard = append(gameCard, info)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Успешно создано",
		"title": req.Title,
		"big_image_url": image_big_url,
		"small_image_url": image_small_url,
	})
}

func SetupGameRoutes(r *gin.Engine) {
	r.Static("/static", "./uploads")

	/* /games */
	r.POST("/games/add", addGame)
	r.GET("/games", getGame)

	/* /cards */
	r.GET("/cards", getCard)
}