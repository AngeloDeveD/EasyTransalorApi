package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M){
	gin.SetMode(gin.TestMode)
	gin.DefaultWriter = io.Discard
	os.Exit(m.Run())
}

func TestDef(t *testing.T){

	router := setupRouter()

	req, _ := http.NewRequest("GET", "/", nil)

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	//Проверяем
	assert.Equal(t, http.StatusOK, w.Code)

	//Проверка тело объекта равно ожидаемоме URL
	expected := `{"message": "API работает!"}`
	assert.JSONEq(t, expected, w.Body.String())
}
