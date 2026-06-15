package main

import (
	"myapi/internal/game"

	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	r := gin.Default()

	repo := &game.InMemoryGameRepo{}

	handler := game.NewGameHanlder(repo)

	game.SetupGameRoutes(r, handler)

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API работает!"})
	})

	r.Run(":8080")
}

//TODO: переписать запросы под новые handler и routes. Добавить поддержку sqlite
