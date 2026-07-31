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

	JWTSecret  string
	EncryptKey string

	InternalKey string

	// Адрес python-сканера, куда Go шлёт POST /scan после загрузки архива
	ScannerURL string

	// Корень, по которому загрузки видны контейнеру-сканеру.
	// В docker-compose это /app/uploads (общий том ./uploads примонтирован в оба контейнера).
	ScannerFileRoot string
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
	encryptKey := os.Getenv("EncryptKey")
	if encryptKey == "" {
		encryptKey = "12345678901234567890123456789012"
	}
	internalKey := os.Getenv("InternalKey")
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

	return &Config{
		Port:            port,
		DBType:          dbType,
		DBName:          dbName,
		DBHost:          dbHost,
		DBPort:          dbPort,
		DBUser:          dbUser,
		DBPass:          dbPass,
		JWTSecret:       jwtSecret,
		EncryptKey:      encryptKey,
		InternalKey:     internalKey,
		ScannerURL:      scannerURL,
		ScannerFileRoot: scannerFileRoot,
	}
}
