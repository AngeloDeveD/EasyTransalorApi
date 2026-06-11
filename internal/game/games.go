package game

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
			},
		},
	},
}

func getGame(c *gin.Context){
	c.JSON(http.StatusAccepted, gameInfo)
}

func getCard(c *gin.Context){
	c.JSON(http.StatusOK, gameCard)
}

func SetupGameRoutes(router *gin.Engine) {
	router.GET("/games", getGame)
	router.GET("/cards", getCard)
}