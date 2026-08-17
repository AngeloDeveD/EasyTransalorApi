package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port   string
	DBType string // "sqlite" или "postgres"

	// Для SQLite
	DBName string

	// Для PostgreSQL
	DBHost string
	DBPort string
	DBUser string
	DBPass string

	JWTSecret  string
	EncryptKey string

	InternalKey string

	// Адрес python-сканера, куда Go шлёт POST /scan после загрузки архива
	ScannerURL string

	// Корень, по которому загрузки видны контейнеру-сканеру.
	// В docker-compose это /app/uploads (общий том ./uploads примонтирован в оба контейнера).
	ScannerFileRoot string

	// Список доверенных browser origins для CORS, например http://localhost:3000,https://example.com.
	CORSAllowedOrigins []string

	RateLimitEnabled        bool
	RateLimitGlobalRequests int
	RateLimitGlobalWindow   time.Duration
	RateLimitAuthRequests   int
	RateLimitAuthWindow     time.Duration
	RateLimitWriteRequests  int
	RateLimitWriteWindow    time.Duration
}

func Load() *Config {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite"
	} // По умолчанию используем SQLite

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "app.db"
	}

	// Дефолтные значения для локального постгреса
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPass := os.Getenv("DB_PASS")
	if dbPass == "" {
		dbPass = "secret"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super_secret_dev_key_123"
	}
	encryptKey := firstEnv("EncryptKey", "ENCRYPT_KEY")
	if encryptKey == "" {
		encryptKey = "12345678901234567890123456789012"
	}
	internalKey := firstEnv("InternalKey", "INTERNAL_KEY")
	if internalKey == "" {
		internalKey = `super_secret_cloud_key_998`
	}

	scannerURL := os.Getenv("SCANNER_URL")
	if scannerURL == "" {
		scannerURL = "http://localhost:8000/scan"
	}

	scannerFileRoot := os.Getenv("SCANNER_FILE_ROOT")
	if scannerFileRoot == "" {
		scannerFileRoot = "/app/uploads"
	}

	corsAllowedOrigins := parseCSVEnv("CORS_ALLOWED_ORIGINS", []string{
		"http://localhost:3000",
		"http://localhost:8080",
	})

	rateLimitEnabled := parseBoolEnv("RATE_LIMIT_ENABLED", true)
	rateLimitGlobalRequests := parseIntEnv("RATE_LIMIT_GLOBAL_REQUESTS", 120)
	rateLimitGlobalWindow := parseDurationEnv("RATE_LIMIT_GLOBAL_WINDOW", time.Minute)
	rateLimitAuthRequests := parseIntEnv("RATE_LIMIT_AUTH_REQUESTS", 10)
	rateLimitAuthWindow := parseDurationEnv("RATE_LIMIT_AUTH_WINDOW", time.Minute)
	rateLimitWriteRequests := parseIntEnv("RATE_LIMIT_WRITE_REQUESTS", 1)
	rateLimitWriteWindow := parseDurationEnv("RATE_LIMIT_WRITE_WINDOW", 10*time.Second)

	return &Config{
		Port:                    port,
		DBType:                  dbType,
		DBName:                  dbName,
		DBHost:                  dbHost,
		DBPort:                  dbPort,
		DBUser:                  dbUser,
		DBPass:                  dbPass,
		JWTSecret:               jwtSecret,
		EncryptKey:              encryptKey,
		InternalKey:             internalKey,
		ScannerURL:              scannerURL,
		ScannerFileRoot:         scannerFileRoot,
		CORSAllowedOrigins:      corsAllowedOrigins,
		RateLimitEnabled:        rateLimitEnabled,
		RateLimitGlobalRequests: rateLimitGlobalRequests,
		RateLimitGlobalWindow:   rateLimitGlobalWindow,
		RateLimitAuthRequests:   rateLimitAuthRequests,
		RateLimitAuthWindow:     rateLimitAuthWindow,
		RateLimitWriteRequests:  rateLimitWriteRequests,
		RateLimitWriteWindow:    rateLimitWriteWindow,
	}
}

func parseCSVEnv(key string, defaults []string) []string {
	raw := os.Getenv(key)
	if raw == "" {
		return defaults
	}

	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func parseBoolEnv(key string, defaultValue bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return defaultValue
	}
	return value
}

func parseIntEnv(key string, defaultValue int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return defaultValue
	}
	return value
}

func parseDurationEnv(key string, defaultValue time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return defaultValue
	}
	return value
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		value := os.Getenv(key)
		if value != "" {
			return value
		}
	}
	return ""
}
