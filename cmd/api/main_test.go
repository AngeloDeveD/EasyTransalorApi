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

//Имена временной тестовой БД (чтобы не портить рабочую app.db)
const testDBName = "test.db"

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard

	//Направляем тесты на отдельную БД через переменную окружения
	os.Setenv("DB_NAME", testDBName)

	code := m.Run()

	//Очистка тестовой БД и WAL-файлов
	os.Remove(testDBName)
	os.Remove(testDBName + "-wal")
	os.Remove(testDBName + "-shm")

	os.Exit(code)
}

func TestDef(t *testing.T) {

	cfg := config.Load()
	router := setupRouter(cfg)

	req, _ := http.NewRequest("GET", "/", nil)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	//Проверяем
	assert.Equal(t, http.StatusOK, w.Code)

	//Проверка тело объекта равно ожидаемоме URL
	expected := `{"message": "API работает!"}`
	assert.JSONEq(t, expected, w.Body.String())
}
