package game

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Для имитации больших файлов
var heavyFileBytes = make([]byte, 5*1024*1024+1)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode) //Переход в тестовыую зону (не будут отображаться лишие ползунки и прочая юслез тема)
	r := gin.New()
	SetupGameRoutes(r) //Маршруты связанные с играми
	return r
}

// Получение информации об игрых
func TestGames(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/games", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	assert.Equal(t, 1, gameInfo[0].ID)
	assert.Equal(t, "Игра номер 1", gameInfo[0].Title)
}

// Получение карточек игр
func TestCards(t *testing.T) {
	router := setupTestRouter()

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

// Для добавления игры
func TestAddGames(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	//Текстовое поле
	_ = multiWriter.WriteField("Title", "Тестовая игра")

	//Загшрузка большой картинки
	bigImage, err := multiWriter.CreateFormFile("big_pic", "big_pic.jpg")
	assert.NoError(t, err)
	_, err = io.WriteString(bigImage, "fake-image-content-1")
	assert.NoError(t, err)

	//Загрузка маленькой картинки
	smallImage, err := multiWriter.CreateFormFile("small_pic", "small_pic.png")
	assert.NoError(t, err)
	_, err = io.WriteString(smallImage, "fake-image-content-2")
	assert.NoError(t, err)

	multiWriter.Close()

	//Создание запроса
	req, err := http.NewRequest(http.MethodPost, "/games/add", &requestBody)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Тестовая игра")

	assert.Contains(t, w.Body.String(), "big_image_url")
	assert.Contains(t, w.Body.String(), "small_image_url")

	// Очистка
	os.RemoveAll("uploads")
}

// Добавление игры, но картинки много весят
func TestAddGames_MultipleFiles_HeavyFiles(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	_ = multiWriter.WriteField("Title", "Тестовая игра")

	// Отправляем 3 файла (а по условию максимум 2)

	//Загшрузка большой картинки с большим размеров
	bigImage, err := multiWriter.CreateFormFile("big_pic", "big_pic_heavy.jpg")
	assert.NoError(t, err)
	bigImage.Write(heavyFileBytes)
	_, err = io.WriteString(bigImage, "fake-image-content-1")
	assert.NoError(t, err)

	//Загрузка маленькой картинки с большим размеров
	smallImage, err := multiWriter.CreateFormFile("small_pic", "small_pic_heavy.png")
	assert.NoError(t, err)
	smallImage.Write(heavyFileBytes)
	_, err = io.WriteString(smallImage, "fake-image-content-2")
	assert.NoError(t, err)
	multiWriter.Close()

	req, _ := http.NewRequest(http.MethodPost, "/games/add", &requestBody)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Ожидаем 400 ошибку с сообщением о лимите
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Файл слишком большой (макс 5 мб)")
}

// Добавление игры, но вместо картинок загружаются неизвестные файлы
func TestAddGames_MultipleFiles_UnsupportFormatFiles(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	_ = multiWriter.WriteField("Title", "Тестовая игра")

	// Отправляем 3 файла (а по условию максимум 2)

	//Загрузка большой картинки с формате .exe
	bigImage, err := multiWriter.CreateFormFile("big_pic", "big_pic_heavy.exe")
	assert.NoError(t, err)
	_, err = io.WriteString(bigImage, "fake-image-content-1")
	assert.NoError(t, err)

	//Загрузка маленькой картинки в формате .dll
	smallImage, err := multiWriter.CreateFormFile("small_pic", "small_pic_heavy.dll")
	assert.NoError(t, err)
	_, err = io.WriteString(smallImage, "fake-image-content-2")
	assert.NoError(t, err)

	multiWriter.Close()

	req, _ := http.NewRequest(http.MethodPost, "/games/add", &requestBody)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Ожидаем 400 ошибку с сообщением о лимите
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Недопустимый формат файла!")
}

// Добавление перевода к игре
func TestAddTranslate(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	//Текстовое поле
	_ = multiWriter.WriteField("authorName", "Какой то автор")
	_ = multiWriter.WriteField("source", "source_url")
	_ = multiWriter.WriteField("version", "0.1")
	_ = multiWriter.WriteField("percentReady", "10")

	//Загшрузка большой картинки
	zipFile, err := multiWriter.CreateFormFile("file", "translate.zip")
	assert.NoError(t, err)
	_, err = io.WriteString(zipFile, "fake-archive-content-1")
	assert.NoError(t, err)

	multiWriter.Close()

	//Тестовый id существующей игры
	gameId := 1

	url := fmt.Sprintf("/games/translate/%d", gameId)

	//Создание запроса
	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Успешно создано")

	// Очистка
	os.RemoveAll("uploads")
}

// Добавление перевода к игре, id является string
func TestAddTranslate__WithStringId(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	//Текстовое поле
	_ = multiWriter.WriteField("authorName", "Какой то автор")
	_ = multiWriter.WriteField("source", "source_url")
	_ = multiWriter.WriteField("version", "0.1")
	_ = multiWriter.WriteField("percentReady", "10")

	//Загшрузка большой картинки
	zipFile, err := multiWriter.CreateFormFile("file", "translate.zip")
	assert.NoError(t, err)
	_, err = io.WriteString(zipFile, "fake-archive-content-1")
	assert.NoError(t, err)

	multiWriter.Close()

	gameId := "TestId1"

	url := fmt.Sprintf("/games/translate/%s", gameId)

	//Создание запроса
	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Полученный id не является числом")
}

// Добавление перевода к игре, но с несуществующим id игры
func TestAddTranslate__InvalidGameId(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	//Текстовое поле
	_ = multiWriter.WriteField("authorName", "Какой то автор")
	_ = multiWriter.WriteField("source", "source_url")
	_ = multiWriter.WriteField("version", "0.1")
	_ = multiWriter.WriteField("percentReady", "10")

	//Загшрузка большой картинки
	zipFile, err := multiWriter.CreateFormFile("file", "translate.zip")
	assert.NoError(t, err)
	_, err = io.WriteString(zipFile, "fake-archive-content-1")
	assert.NoError(t, err)

	multiWriter.Close()

	gameId := 9999999

	url := fmt.Sprintf("/games/translate/%d", gameId)

	//Создание запроса
	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Игра с таким id не найдена")
}

// Добавление перевода к игре, но с неизвестынм форматом файла
func TestAddTranslate__InvaliFileFormat(t *testing.T) {
	r := setupTestRouter()

	var requestBody bytes.Buffer
	multiWriter := multipart.NewWriter(&requestBody)

	//Текстовое поле
	_ = multiWriter.WriteField("authorName", "Какой то автор")
	_ = multiWriter.WriteField("source", "source_url")
	_ = multiWriter.WriteField("version", "0.1")
	_ = multiWriter.WriteField("percentReady", "10")

	//Загшрузка большой картинки
	zipFile, err := multiWriter.CreateFormFile("file", "translate.exe")
	assert.NoError(t, err)
	_, err = io.WriteString(zipFile, "fake-archive-content-1")
	assert.NoError(t, err)

	multiWriter.Close()

	gameId := 1

	url := fmt.Sprintf("/games/translate/%d", gameId)

	//Создание запроса
	req, err := http.NewRequest(http.MethodPost, url, &requestBody)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", multiWriter.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Недопустимый формат файла!")
}
