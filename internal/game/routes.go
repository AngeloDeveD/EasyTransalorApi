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
		public.GET("/translations/:transid/files", handler.GetTranslationFiles)
	}

	privateMiddlewares := append([]gin.HandlerFunc{authHandler}, writeMiddlewares...)
	private := router.Group("", privateMiddlewares...)
	{
		private.POST("/games/add", handler.AddGame)
		private.POST("games/translate/:gameid", handler.AddTranslationInfo)
		private.POST("/api/files/hash-check", handler.HashCheckArchive)
		private.GET("/translations/:transid/status", handler.GetTranslationStatus)
		private.GET("/api/me/translations", handler.GetMyTranslations)
		private.DELETE("/translations/:transid", handler.DeleteMyTranslation)
	}

	adminMiddlewares := append([]gin.HandlerFunc{authHandler, adminHandler}, writeMiddlewares...)
	adminOnly := router.Group("", adminMiddlewares...)
	{
		adminOnly.DELETE("/games/:gameid", handler.DeleteGame)
		adminOnly.DELETE("/games/translate/:transid", handler.DeleteTranslation)
	}

}
