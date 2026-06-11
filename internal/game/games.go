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
func addGame(c *gin.Context){
	//Привязка текстовых данных из формф
	var req CreateGameRequest

	if err := c.ShouldBind(&req); err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Получение всеё формы
	form, err := c.MultipartForm()
	if err != nil{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл обязателен"})
	}

	//Получение файлов
	files := form.File["images"]

	if len(files) == 0{
		c.JSON(http.StatusBadRequest, gin.H{"error": "Загрузите хотя бы одно фото"})
		return
	}

	if len(files) > 2{
		c.JSON(http.StatusBadRequest, gin.H{"error" : "Максимум 2 изображения"})
	}

	allowExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}

	var imagesUrls []string

	for _, file := range files {
		//Валидация изоброажений
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowExts[ext]{
			c.JSON(http.StatusBadRequest, gin.H{"error" : fmt.Sprintf("Файл %s имеет  недопустимый формат", file.Filename)})
			return
		}

		if file.Size > 5<<20{
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Файл %s слишком большой", file.Filename)})
			return
		}

		newFileName := uuid.New().String() + ext
		savePath := filepath.Join("uploads", newFileName)

		if err := c.SaveUploadedFile(file, savePath); err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить файл"})
			return
		}

		imagesUrls = append(imagesUrls, fmt.Sprintf("/static/%s", newFileName))
	}

	info := GameCard{
		ID: int(uuid.New().ID()),
		Title: req.Title,
		IconUrl: imagesUrls[0],
	}

	gameCard = append(gameCard, info)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Успешно создано",
		"title": req.Title,
		"image_urls": imagesUrls,
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