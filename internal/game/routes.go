package game

import "github.com/gin-gonic/gin"

func SetupGameRoutes(router *gin.Engine, handler *GameHandler) {

	router.Static("/static", "./uploads")

	/* /games */
	//Надо переписать под хэндер
	// router.GET("/games", getGame)
	// router.POST("/games/add", addGame)
	// router.POST("games/translate/:gameid", addTranslate)
	/* /cards */

	router.GET("/cards", handler.GetCards)
}
