package game

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGames(t *testing.T){
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	SetupGameRoutes(router)

	req, _ := http.NewRequest("GET", "/games", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	assert.Contains(t, w.Body.String(), "Игра номер 1")
}

func TestCards(t *testing.T){
	gin.SetMode(gin.TestMode)

	router := gin.New()
	SetupGameRoutes(router)

	req, _ := http.NewRequest("GET", "/cards", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	expected := `[{
		"id": 1, 
		"title": "Игра номер 1",
		"iconUrl": "source",
		"gameId": 1
	}]`

	assert.JSONEq(t, expected, w.Body.String())
}