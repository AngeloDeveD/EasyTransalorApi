package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// При пустом окружении Load подставляет дефолтные значения.
func TestLoad_Defaults(t *testing.T) {
	// Обнуляем все переменные окружения, которые читает Load.
	for _, k := range []string{
		"APP_PORT", "DB_TYPE", "DB_NAME", "DB_HOST", "DB_PORT",
		"DB_USER", "DB_PASS", "JWT_SECRET", "EncryptKey", "InternalKey",
		"SCANNER_URL", "SCANNER_FILE_ROOT",
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
	t.Setenv("EncryptKey", "abcdefghijklmnopqrstuvwxyz012345")
	t.Setenv("InternalKey", "my-internal")
	t.Setenv("SCANNER_URL", "http://scanner:8000/scan")
	t.Setenv("SCANNER_FILE_ROOT", "/data/uploads")

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
}
