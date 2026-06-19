package game

import "github.com/gin-gonic/gin"

func SetupGameRoutes(router *gin.Engine, handler *GameHandler) {

	router.Static("/static", "./uploads")

	router.GET("/cards", handler.GetCards)
	router.GET("/games", handler.GetGames)
	router.GET("/games/:gameid", handler.GetGameById)
	router.POST("/games/add", handler.AddGame)
	router.POST("games/translate/:gameid", handler.AddTranslationInfo)
}
