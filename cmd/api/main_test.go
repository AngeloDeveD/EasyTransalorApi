package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"myapi/internal/config"
)

const testDBName = "test.db"

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard

	os.Setenv("DB_NAME", testDBName)
	os.Setenv("JWT_SECRET", "test_secret_for_main") // НОВОЕ: добавили секрет для JWT

	code := m.Run()

	os.Remove(testDBName)
	os.Remove(testDBName + "-wal")
	os.Remove(testDBName + "-shm")
	os.RemoveAll("uploads") // На всякий случай чистим папку uploads

	os.Exit(code)
}

func TestDef(t *testing.T) {
	cfg := config.Load()
	router := setupRouter(cfg)

	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	expected := `{"message": "API работает!"}`
	assert.JSONEq(t, expected, w.Body.String())
}
