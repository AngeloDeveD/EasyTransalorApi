package game

import (
	"github.com/gin-gonic/gin"
)

func SetupGameRoutes(router *gin.Engine, handler *GameHandler, authHandler gin.HandlerFunc, adminHandler gin.HandlerFunc, writeMiddlewares ...gin.HandlerFunc) {

	public := router.Group("")
	{
		public.GET("/cards", handler.GetCards)
		public.GET("/games", handler.GetGames)
		public.GET("/games/:gameid", handler.GetGameById)
		public.GET("/download/:transid", handler.DownloadGameTranslation)
	}

	privateMiddlewares := append([]gin.HandlerFunc{authHandler}, writeMiddlewares...)
	private := router.Group("", privateMiddlewares...)
	{
		private.POST("/games/add", handler.AddGame)
		private.POST("games/translate/:gameid", handler.AddTranslationInfo)
	}

	adminMiddlewares := append([]gin.HandlerFunc{authHandler, adminHandler}, writeMiddlewares...)
	adminOnly := router.Group("", adminMiddlewares...)
	{
		adminOnly.DELETE("/games/:gameid", handler.DeleteGame)
		adminOnly.DELETE("/games/translate/:transid", handler.DeleteTranslation)
	}

}
