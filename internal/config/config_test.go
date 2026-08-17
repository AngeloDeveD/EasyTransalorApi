package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// При пустом окружении Load подставляет дефолтные значения.
func TestLoad_Defaults(t *testing.T) {
	// Обнуляем все переменные окружения, которые читает Load.
	for _, k := range []string{
		"APP_PORT", "DB_TYPE", "DB_NAME", "DB_HOST", "DB_PORT",
		"DB_USER", "DB_PASS", "JWT_SECRET", "EncryptKey", "ENCRYPT_KEY", "InternalKey", "INTERNAL_KEY",
		"SCANNER_URL", "SCANNER_FILE_ROOT", "CORS_ALLOWED_ORIGINS",
		"RATE_LIMIT_ENABLED", "RATE_LIMIT_GLOBAL_REQUESTS", "RATE_LIMIT_GLOBAL_WINDOW",
		"RATE_LIMIT_AUTH_REQUESTS", "RATE_LIMIT_AUTH_WINDOW",
		"RATE_LIMIT_WRITE_REQUESTS", "RATE_LIMIT_WRITE_WINDOW",
	} {
		t.Setenv(k, "")
	}

	cfg := Load()

	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "sqlite", cfg.DBType)
	assert.Equal(t, "app.db", cfg.DBName)
	assert.Equal(t, "localhost", cfg.DBHost)
	assert.Equal(t, "5432", cfg.DBPort)
	assert.Equal(t, "postgres", cfg.DBUser)
	assert.Equal(t, "secret", cfg.DBPass)
	assert.Equal(t, "super_secret_dev_key_123", cfg.JWTSecret)
	assert.Len(t, cfg.EncryptKey, 32)
	assert.NotEmpty(t, cfg.InternalKey)
	assert.Equal(t, "http://localhost:8000/scan", cfg.ScannerURL)
	assert.Equal(t, "/app/uploads", cfg.ScannerFileRoot)
	assert.Equal(t, []string{"http://localhost:3000", "http://localhost:8080"}, cfg.CORSAllowedOrigins)
	assert.True(t, cfg.RateLimitEnabled)
	assert.Equal(t, 120, cfg.RateLimitGlobalRequests)
	assert.Equal(t, time.Minute, cfg.RateLimitGlobalWindow)
	assert.Equal(t, 10, cfg.RateLimitAuthRequests)
	assert.Equal(t, time.Minute, cfg.RateLimitAuthWindow)
	assert.Equal(t, 1, cfg.RateLimitWriteRequests)
	assert.Equal(t, 10*time.Second, cfg.RateLimitWriteWindow)
}

// Значения из окружения имеют приоритет над дефолтами.
func TestLoad_Overrides(t *testing.T) {
	t.Setenv("APP_PORT", "9999")
	t.Setenv("DB_TYPE", "postgres")
	t.Setenv("DB_NAME", "custom.db")
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("DB_PORT", "6543")
	t.Setenv("DB_USER", "admin")
	t.Setenv("DB_PASS", "hunter2")
	t.Setenv("JWT_SECRET", "my-jwt")
	t.Setenv("ENCRYPT_KEY", "abcdefghijklmnopqrstuvwxyz012345")
	t.Setenv("INTERNAL_KEY", "my-internal")
	t.Setenv("SCANNER_URL", "http://scanner:8000/scan")
	t.Setenv("SCANNER_FILE_ROOT", "/data/uploads")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com, https://admin.example.com")
	t.Setenv("RATE_LIMIT_ENABLED", "false")
	t.Setenv("RATE_LIMIT_GLOBAL_REQUESTS", "50")
	t.Setenv("RATE_LIMIT_GLOBAL_WINDOW", "30s")
	t.Setenv("RATE_LIMIT_AUTH_REQUESTS", "3")
	t.Setenv("RATE_LIMIT_AUTH_WINDOW", "2m")
	t.Setenv("RATE_LIMIT_WRITE_REQUESTS", "2")
	t.Setenv("RATE_LIMIT_WRITE_WINDOW", "15s")

	cfg := Load()

	assert.Equal(t, "9999", cfg.Port)
	assert.Equal(t, "postgres", cfg.DBType)
	assert.Equal(t, "custom.db", cfg.DBName)
	assert.Equal(t, "db.example.com", cfg.DBHost)
	assert.Equal(t, "6543", cfg.DBPort)
	assert.Equal(t, "admin", cfg.DBUser)
	assert.Equal(t, "hunter2", cfg.DBPass)
	assert.Equal(t, "my-jwt", cfg.JWTSecret)
	assert.Equal(t, "abcdefghijklmnopqrstuvwxyz012345", cfg.EncryptKey)
	assert.Equal(t, "my-internal", cfg.InternalKey)
	assert.Equal(t, "http://scanner:8000/scan", cfg.ScannerURL)
	assert.Equal(t, "/data/uploads", cfg.ScannerFileRoot)
	assert.Equal(t, []string{"https://app.example.com", "https://admin.example.com"}, cfg.CORSAllowedOrigins)
	assert.False(t, cfg.RateLimitEnabled)
	assert.Equal(t, 50, cfg.RateLimitGlobalRequests)
	assert.Equal(t, 30*time.Second, cfg.RateLimitGlobalWindow)
	assert.Equal(t, 3, cfg.RateLimitAuthRequests)
	assert.Equal(t, 2*time.Minute, cfg.RateLimitAuthWindow)
	assert.Equal(t, 2, cfg.RateLimitWriteRequests)
	assert.Equal(t, 15*time.Second, cfg.RateLimitWriteWindow)
}
