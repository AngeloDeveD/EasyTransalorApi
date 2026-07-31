package files

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeFileHeader собирает *multipart.FileHeader из содержимого — так же,
// как это делает Gin при разборе входящей формы.
func makeFileHeader(t *testing.T, field, filename, content string) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, err := w.CreateFormFile(field, filename)
	require.NoError(t, err)
	_, err = io.WriteString(fw, content)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(32<<20))

	_, fh, err := req.FormFile(field)
	require.NoError(t, err)
	return fh
}

func TestIsAllowedArchiveFormat(t *testing.T) {
	repo := NewLocalFileRepo()

	cases := []struct {
		filename string
		ok       bool
	}{
		{"translate.zip", true},
		{"translate.rar", true},
		{"translate.7zip", true},
		{"TRANSLATE.ZIP", true}, // регистр расширения не важен
		{"virus.exe", false},
		{"lib.dll", false},
		{"noext", false},
	}

	for _, c := range cases {
		err := repo.IsAllowedArchiveFormat(&multipart.FileHeader{Filename: c.filename})
		if c.ok {
			assert.NoError(t, err, c.filename)
		} else {
			assert.EqualError(t, err, "неподдерживаемый формат файла!", c.filename)
		}
	}
}

func TestIsAllowedImageFormat(t *testing.T) {
	repo := NewLocalFileRepo()

	cases := []struct {
		filename string
		ok       bool
	}{
		{"pic.jpg", true},
		{"pic.jpeg", true},
		{"pic.png", true},
		{"pic.webp", true},
		{"pic.PNG", true},
		{"pic.gif", false},
		{"pic.exe", false},
	}

	for _, c := range cases {
		err := repo.IsAllowedImageFormat(&multipart.FileHeader{Filename: c.filename})
		if c.ok {
			assert.NoError(t, err, c.filename)
		} else {
			assert.EqualError(t, err, "неподдерживаемый формат изображения", c.filename)
		}
	}
}

func TestIsAllowedArchiveSize(t *testing.T) {
	repo := NewLocalFileRepo()

	assert.NoError(t, repo.IsAllowedArchiveSize(1<<20))     // 1 МБ — ок
	assert.NoError(t, repo.IsAllowedArchiveSize(5<<30))     // ровно 5 ГБ — ещё ок
	assert.EqualError(t, repo.IsAllowedArchiveSize(5<<30+1), // 5 ГБ + 1 байт — превышение
		"файл слишком большой!")
}

func TestIsAllowedImageSize(t *testing.T) {
	repo := NewLocalFileRepo()

	assert.NoError(t, repo.IsAllowedImageSize(1<<20, 1<<20))
	assert.EqualError(t, repo.IsAllowedImageSize(5<<20+1, 1<<20), "размер изображения большой")
	assert.EqualError(t, repo.IsAllowedImageSize(1<<20, 5<<20+1), "размер изображения большой")
}

func TestGetArchivePath_Empty(t *testing.T) {
	repo := NewLocalFileRepo()
	_, _, err := repo.GetArchivePath("Игра", "Автор", "")
	assert.EqualError(t, err, "URL файла пуст!")
}

func TestGetArchivePath_Valid(t *testing.T) {
	repo := NewLocalFileRepo()

	path, filename, err := repo.GetArchivePath("Моя Игра", "Автор Крутой", "/static/files/1/abc.zip")
	assert.NoError(t, err)
	// URL превращается в путь на диске
	assert.Equal(t, "uploads/files/1/abc.zip", path)
	// пробелы и + в имени заменяются на _
	assert.Equal(t, "Моя_Игра_Автор_Крутой.zip", filename)
}

func TestSaveArchive_RoundTrip(t *testing.T) {
	// Работаем во временном каталоге, чтобы не мусорить в репозитории.
	dir := t.TempDir()
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(old)

	repo := NewLocalFileRepo()
	fh := makeFileHeader(t, "file", "translate.zip", "PK\x03\x04fake-archive")

	url, err := repo.SaveArchive(7, fh)
	require.NoError(t, err)

	// file_url в формате static/files/{gameid}/{uuid}.zip (forward-slash, без ведущего /)
	assert.True(t, strings.HasPrefix(url, "static/files/7/"), url)
	assert.True(t, strings.HasSuffix(url, ".zip"), url)

	// Файл реально лежит на диске по соответствующему uploads-пути
	diskPath := filepath.FromSlash(strings.Replace(url, "static", "uploads", 1))
	_, err = os.Stat(diskPath)
	assert.NoError(t, err)
}

func TestSaveImages_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	old, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer os.Chdir(old)

	repo := NewLocalFileRepo()
	big := makeFileHeader(t, "big_pic", "big.jpg", "big-content")
	small := makeFileHeader(t, "small_pic", "small.png", "small-content")

	// SaveImages сам создаёт каталоги uploads/Icons/Big и uploads/Icons/Small.
	urls, err := repo.SaveImages(big, small)
	require.NoError(t, err)
	require.Len(t, urls, 2)

	// [0] — маленькая иконка, [1] — большая
	assert.True(t, strings.Contains(urls[0], "static/Icons/Small/"), urls[0])
	assert.True(t, strings.Contains(urls[1], "static/Icons/Big/"), urls[1])
}
