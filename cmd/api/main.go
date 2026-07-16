package main

import (
	"fmt"
	"log"
	"myapi/internal/auth"
	"myapi/internal/config"
	"myapi/internal/database"
	"myapi/internal/game"
	"os"

	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	//Обработка cli для назначения admin
	if len(os.Args) > 2 && os.Args[1] == "--make-admin" {
		nickname := os.Args[2]
		makeAdmin(cfg, nickname)
		return //Завершение программы после выполнения
	}

	r := setupRouter(cfg)

	r.Run(":" + cfg.Port)
}

// makeAdmin выдаёт права администратора по никнейму, который был передан при запуске программы
// Пример: app.exe --make-admin yapidoras2012
func makeAdmin(cfg *config.Config, nickname string) {
	db, err := database.ConnectSqlite(cfg.DBName)
	if err != nil {
		log.Fatalf("Ошибка БД: %v", err)
	}

	//Поиск пользователя
	var user auth.User
	result := db.Where("nickname = ?", nickname).First(&user)

	if result.Error != nil {
		log.Fatalf("Пользователь '%s' не найден!", nickname)
	}

	//Обновление роли
	user.Role = "admin"
	db.Save(&user)

	fmt.Printf("Пользователь '%s' успешно назначен администратором!\n", nickname)
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

	game.SetupGameRoutes(r, gameHandler, auth.AuthMiddleware(jwtManager), auth.AdminMiddleware())

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API работает!"})
	})

	return r
}
