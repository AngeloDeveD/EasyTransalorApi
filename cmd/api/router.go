package main

import (
	_ "myapi/docs"
	"myapi/internal/auth"
	"myapi/internal/chat"
	"myapi/internal/config"
	"myapi/internal/database"
	"myapi/internal/files"
	"myapi/internal/game"
	"myapi/internal/notification"
	"myapi/internal/ratelimit"
	"myapi/internal/scanner"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func setupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware(cfg.CORSAllowedOrigins))

	limiter := ratelimit.NewMemoryLimiter()
	r.Use(limiter.Middleware(ratelimit.Config{
		Enabled:  cfg.RateLimitEnabled,
		Requests: cfg.RateLimitGlobalRequests,
		Window:   cfg.RateLimitGlobalWindow,
	}, ratelimit.KeyByClientIP("global")))

	db, err := database.ConnectDB(cfg)
	if err != nil {
		panic("Ошибка подключения к БД: " + err.Error())
	}

	// --- Инициализация репозиториев ---
	authRepo := auth.NewSqlUserRepo(db)
	notifRepo := notification.NewSqlNotificationRepo(db)
	fileRepo := files.NewLocalFileRepo()
	gameRepo := game.NewSqlGameRepo(db)
	chatRepo := chat.NewSqliteChatRepo(db)

	// --- Инициализация сервисов ---
	jwtManager := auth.NewJWTManager(cfg.JWTSecret)
	scannerClient := scanner.NewClient(cfg.ScannerURL, cfg.InternalKey, cfg.ScannerFileRoot)

	// --- Настройка маршрутов ---

	checkedAuthMiddleware := auth.AuthMiddlewareWithUserCheck(jwtManager, authRepo)

	// Авторизация
	autoHandler := auth.NewAuthHandler(authRepo, jwtManager)
	auth.SetupAuthRoutes(r, autoHandler, limiter.Middleware(ratelimit.Config{
		Enabled:  cfg.RateLimitEnabled,
		Requests: cfg.RateLimitAuthRequests,
		Window:   cfg.RateLimitAuthWindow,
	}, ratelimit.KeyByClientIP("auth")))

	// Уведомления
	notifHandler := notification.NewNotificationHandler(notifRepo)
	notification.SetupNotificationRoutes(r, notifHandler, checkedAuthMiddleware, auth.ModeratorMiddleware())

	// Админка
	adminHandler := auth.NewAdminHandler(authRepo, notifRepo)
	auth.SetupAdminRoutes(r, adminHandler, checkedAuthMiddleware, auth.ModeratorMiddleware(), auth.AdminMiddleware())

	// Игры
	gameHandler := game.NewGameHandler(gameRepo, fileRepo, scannerClient)
	game.SetupGameRoutes(r, gameHandler, checkedAuthMiddleware, auth.ModeratorMiddleware(), limiter.Middleware(ratelimit.Config{
		Enabled:  cfg.RateLimitEnabled,
		Requests: cfg.RateLimitWriteRequests,
		Window:   cfg.RateLimitWriteWindow,
	}, ratelimit.KeyByUserOrIP("game-write")))

	// Чат
	chatHub := chat.NewHub()
	chatHandler := chat.NewChatHandler(chatHub, chatRepo, []byte(cfg.EncryptKey), cfg.CORSAllowedOrigins)
	chat.SetupChatRoutes(r, chatHandler, checkedAuthMiddleware)

	// Модерация
	moderationHandler := game.NewModerationHandler(gameRepo, authRepo, notifRepo)
	game.SetupModerationRoutes(r, moderationHandler, checkedAuthMiddleware, auth.ModeratorMiddleware())

	// Внутренние роуты (для Python-сканера)
	internalHandler := game.NewInternalHandler(gameRepo, authRepo, fileRepo, notifRepo)
	internalGroup := r.Group("/api/internal", auth.APIKeyMiddleware(cfg.InternalKey))
	{
		internalGroup.POST("/scan-result", internalHandler.ReceiveScanResult)
	}

	// Фоновые процессы
	game.StartCleaner(gameRepo)

	// Swagger UI
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/api/tos", func(c *gin.Context) {
		c.JSON(http.StatusTeapot, gin.H{"message": "Пока ничего нету. Я - пакетик"})
	})

	// Healthcheck
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "API работает!"})
	})

	return r
}
