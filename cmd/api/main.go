package main

import (
	"log"
	"myapi/internal/auth"
	"myapi/internal/config"
	"myapi/internal/database"
	"myapi/internal/game"

	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	r := setupRouter(cfg)

	r.Run(":" + cfg.Port)
}

// setupRouter настраивает роутер, подключает БД и регистрирует маршруты.
// Вынесен отдельно, чтобы переиспользовать его в тестах (main_test.go).
func setupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	db, err := database.ConnectSqlite(cfg.DBName)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}

	//Настройка авторизацииgameRepo.NewJWTManager(cfg.JWTSecret)
	authRepo := auth.NewSqlUserRepo(db)
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	autoHandler := auth.NewAuthHandler(authRepo, jwtManager)
	auth.SetupAuthRoutes(r, autoHandler)

	//Настройка игр
	gameRepo := game.NewSqlGameRepo(db)
	gameHandler := game.NewGameHandler(gameRepo)

	game.SetupGameRoutes(r, gameHandler, auth.AuthMiddleware(jwtManager))

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API работает!"})
	})

	return r
}
