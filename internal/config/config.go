package config

import "os"

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

	JWTSecret string
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

	return &Config{
		Port:      port,
		DBType:    dbType,
		DBName:    dbName,
		DBHost:    dbHost,
		DBPort:    dbPort,
		DBUser:    dbUser,
		DBPass:    dbPass,
		JWTSecret: jwtSecret,
	}
}
