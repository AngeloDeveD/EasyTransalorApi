package game

import "github.com/gin-gonic/gin"

func SetupGameRoutes(router *gin.Engine, handler *GameHandler, authHandler gin.HandlerFunc) {

	router.Static("/static", "./uploads")

	public := router.Group("")
	{
		public.GET("/cards", handler.GetCards)
		public.GET("/games", handler.GetGames)
		router.GET("/games/:gameid", handler.GetGameById)
	}

	private := router.Group("", authHandler)
	{
		private.POST("/games/add", handler.AddGame)
		private.POST("games/translate/:gameid", handler.AddTranslationInfo)
	}

}
