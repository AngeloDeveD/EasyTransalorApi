package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"myapi/internal/files"
)

var heavyFileBytes = make([]byte, 5*1024*1024+1)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	repo := NewInMemoryGameRepo()
	// FileRepo — реальный локальный репозиторий (тесты сохраняют файлы);
	// Scanner=nil — в тестах антивирусную проверку не запускаем.
	handler := NewGameHandler(repo, files.NewLocalFileRepo(), nil)

	// Заглушки для middleware
	dummyAuthMiddleware := func(c *gin.Context) {
		c.Set("userID", 1)
		c.Set("role", "author")
		c.Next()
	}
	dummyAdminMiddleware := func(c *gin.Context) {
		c.Next() // В тестах игр admim middleware всегда пропускает
	}

	// Передаем 4 аргумента
	SetupGameRoutes(r, handler, dummyAuthMiddleware, dummyAdminMiddleware)
	return r
}

// Получение информации об играх
func TestGames(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/games", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	var games []PublicGameInfo
	err := json.Unmarshal(w.Body.Bytes(), &games)
	assert.NoError(t, err)

	assert.NotEmpty(t, games)
	assert.Equal(t, 1, games[0].ID)
	assert.Equal(t, "Игра номер 1", games[0].Title)
	assert.NotContains(t, w.Body.String(), "archiveHash")
	assert.NotContains(t, w.Body.String(), "urlToDownload")
	assert.NotContains(t, w.Body.String(), "scanDetails")
}

// Получение карточек игр
func TestCards(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/cards", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	expected := `[{
        "id": 1,
        "title": "Игра номер 1",
        "iconUrl": "source",
        "gameId": 1
    }]`

	assert.JSONEq(t, expected, w.Body.String())
}

// Для добавления игры
func TestAddGames(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	_ = multiWriter.WriteField("Title", "Тестовая игра")

	bigImage, err := multiWriter.CreateFormFile("big_pic", "big_pic.jpg")
	assert.NoError(t, err)
	_, err = bigImage.Write([]byte{0xff, 0xd8, 0xff, 0xdb, 0x00, 0x43, 0x00})
	assert.NoError(t, err)

	smallImage, err := multiWriter.CreateFormFile("small_pic", "small_pic.png")
	assert.NoError(t, err)
	_, err = smallImage.Write([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
	assert.NoError(t, err)

	multiWriter.Close()

	req, err := http.NewRequest(http.MethodPost, "/games/add", &requestBody)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Тестовая игра")
	assert.Contains(t, w.Body.String(), "big_image_url")
	assert.Contains(t, w.Body.String(), "small_image_url")

	// Используем t.Cleanup для надежной очистки
	t.Cleanup(func() {
		os.RemoveAll("uploads")
	})
}

// Добавление игры, но картинки много весят
func TestAddGames_MultipleFiles_HeavyFiles(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	_ = multiWriter.WriteField("Title", "Тестовая игра")

	bigImage, err := multiWriter.CreateFormFile("big_pic", "big_pic_heavy.jpg")
	assert.NoError(t, err)
	bigImage.Write(heavyFileBytes)
	_, err = bigImage.Write([]byte{0xff, 0xd8, 0xff, 0xdb, 0x00, 0x43, 0x00})
	assert.NoError(t, err)

	smallImage, err := multiWriter.CreateFormFile("small_pic", "small_pic_heavy.png")
	assert.NoError(t, err)
	smallImage.Write(heavyFileBytes)
	_, err = smallImage.Write([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
	assert.NoError(t, err)
	multiWriter.Close()

	req, _ := http.NewRequest(http.MethodPost, "/games/add", &requestBody)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "размер изображения большой")
}

// Добавление игры, но вместо картинок загружаются неизвестные файлы
func TestAddGames_MultipleFiles_UnsupportFormatFiles(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	_ = multiWriter.WriteField("Title", "Тестовая игра")

	bigImage, err := multiWriter.CreateFormFile("big_pic", "big_pic_heavy.exe")
	assert.NoError(t, err)
	_, err = bigImage.Write([]byte{0xff, 0xd8, 0xff, 0xdb, 0x00, 0x43, 0x00})
	assert.NoError(t, err)

	smallImage, err := multiWriter.CreateFormFile("small_pic", "small_pic_heavy.dll")
	assert.NoError(t, err)
	_, err = smallImage.Write([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
	assert.NoError(t, err)

	multiWriter.Close()

	req, _ := http.NewRequest(http.MethodPost, "/games/add", &requestBody)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "неподдерживаемый формат изображения")
}

// Добавление перевода к игре
func TestAddTranslate(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	_ = multiWriter.WriteField("authorName", "Какой то автор")
	_ = multiWriter.WriteField("source", "source_url")
	_ = multiWriter.WriteField("version", "0.1")
	_ = multiWriter.WriteField("percentReady", "10")

	zipFile, err := multiWriter.CreateFormFile("file", "translate.zip")
	assert.NoError(t, err)
	_, err = io.WriteString(zipFile, "PK\x03\x04fake-archive-content-1")
	assert.NoError(t, err)

	multiWriter.Close()

	gameId := 1
	url := fmt.Sprintf("/games/translate/%d", gameId)

	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Успешно создано")

	t.Cleanup(func() {
		os.RemoveAll("uploads")
	})
}

// Добавление перевода к игре, id является string (исправлено имя функции)
func TestAddTranslate_WithStringId(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	_ = multiWriter.WriteField("authorName", "Какой то автор")
	_ = multiWriter.WriteField("source", "source_url")
	_ = multiWriter.WriteField("version", "0.1")
	_ = multiWriter.WriteField("percentReady", "10")

	zipFile, err := multiWriter.CreateFormFile("file", "translate.zip")
	assert.NoError(t, err)
	_, err = io.WriteString(zipFile, "PK\x03\x04fake-archive-content-1")
	assert.NoError(t, err)

	multiWriter.Close()

	gameId := "TestId1"
	url := fmt.Sprintf("/games/translate/%s", gameId)

	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Полученный id не является числом")
}

// Добавление перевода к игре, но с несуществующим id игры
func TestAddTranslate_InvalidGameId(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	_ = multiWriter.WriteField("authorName", "Какой то автор")
	_ = multiWriter.WriteField("source", "source_url")
	_ = multiWriter.WriteField("version", "0.1")
	_ = multiWriter.WriteField("percentReady", "10")

	zipFile, err := multiWriter.CreateFormFile("file", "translate.zip")
	assert.NoError(t, err)
	_, err = io.WriteString(zipFile, "PK\x03\x04fake-archive-content-1")
	assert.NoError(t, err)

	multiWriter.Close()

	gameId := 9999999
	url := fmt.Sprintf("/games/translate/%d", gameId)

	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Игра была не найдена")
}

// Добавление перевода к игре, но с неизвестным форматом файла (исправлена опечатка в имени)
func TestAddTranslate_InvalidFileFormat(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	_ = multiWriter.WriteField("authorName", "Какой то автор")
	_ = multiWriter.WriteField("source", "source_url")
	_ = multiWriter.WriteField("version", "0.1")
	_ = multiWriter.WriteField("percentReady", "10")

	zipFile, err := multiWriter.CreateFormFile("file", "translate.exe")
	assert.NoError(t, err)
	_, err = io.WriteString(zipFile, "PK\x03\x04fake-archive-content-1")
	assert.NoError(t, err)

	multiWriter.Close()

	gameId := 1
	url := fmt.Sprintf("/games/translate/%d", gameId)

	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "неподдерживаемый формат файла!")
}

// Получение игры по строковому ID (не числу)
func TestGetGameById_InvalidStringId(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/games/NotANumber", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Полученный id не является числом")
}

// Получение игры по несуществующему ID
func TestGetGameById_NonExistentId(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/games/9999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Игра не найдена")
}

// Добавление игры без указания Title
func TestAddGame_MissingTitle(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	// Намеренно не пишем multiWriter.WriteField("Title", ...)

	bigImage, _ := multiWriter.CreateFormFile("big_pic", "big_pic.jpg")
	io.WriteString(bigImage, "fake")

	smallImage, _ := multiWriter.CreateFormFile("small_pic", "small_pic.png")
	io.WriteString(smallImage, "fake")

	multiWriter.Close()

	req, _ := http.NewRequest(http.MethodPost, "/games/add", &requestBody)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Удаление игры
func TestDeleteGame(t *testing.T) {
	r := setupTestRouter()

	// В InMemoryGameRepo у нас есть игра с ID 1
	req, _ := http.NewRequest(http.MethodDelete, "/games/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "была удалена")
}

// Удаление несуществующей игры
func TestDeleteGame_NonExistent(t *testing.T) {
	r := setupTestRouter()

	req, _ := http.NewRequest(http.MethodDelete, "/games/9999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
