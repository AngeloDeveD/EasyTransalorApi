package game

import "github.com/gin-gonic/gin"

func SetupGameRoutes(router *gin.Engine, handler *GameHandler, authHandler gin.HandlerFunc, adminHandler gin.HandlerFunc) {

	public := router.Group("")
	{
		public.GET("/cards", handler.GetCards)
		public.GET("/games", handler.GetGames)
		public.GET("/games/:gameid", handler.GetGameById)
	}

	private := router.Group("", authHandler)
	{
		private.POST("/games/add", handler.AddGame)
		private.POST("games/translate/:gameid", handler.AddTranslationInfo)
		private.GET("/download/:gameid/:translid", handler.DownloadGameTranslation)
	}

	adminOnly := router.Group("", authHandler, adminHandler)
	{
		adminOnly.DELETE("/games/:gameid", handler.DeleteGame)
		adminOnly.DELETE("/games/translate/:gameid/:transid", handler.DeleteTranslation)
	}

}
