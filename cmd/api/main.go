package main

import (
	"myapi/internal/game"

	"net/http"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	//GET
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API работает!"})
	})

	game.SetupGameRoutes(r)

	return r
}

func main(){

	r := setupRouter()

	r.Run(":8080")
}