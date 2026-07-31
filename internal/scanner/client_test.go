package scanner

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// filePath переводит публичный file_url в путь внутри контейнера сканера.
func TestFilePath(t *testing.T) {
	cases := []struct {
		name     string
		fileRoot string
		fileURL  string
		want     string
	}{
		{
			name:     "обычный url без ведущего слэша",
			fileRoot: "/app/uploads",
			fileURL:  "static/files/1/uuid.zip",
			want:     "/app/uploads/files/1/uuid.zip",
		},
		{
			name:     "url с ведущим слэшем",
			fileRoot: "/app/uploads",
			fileURL:  "/static/files/2/x.rar",
			want:     "/app/uploads/files/2/x.rar",
		},
		{
			name:     "корень с завершающим слэшем не удваивает разделитель",
			fileRoot: "/app/uploads/",
			fileURL:  "static/files/3/y.7zip",
			want:     "/app/uploads/files/3/y.7zip",
		},
		{
			name:     "url без префикса static остаётся как есть",
			fileRoot: "/app/uploads",
			fileURL:  "files/4/z.zip",
			want:     "/app/uploads/files/4/z.zip",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cl := NewClient("http://scanner:8000/scan", "key", c.fileRoot)
			assert.Equal(t, c.want, cl.filePath(c.fileURL))
		})
	}
}

// send успешно отправляет корректный JSON и заголовок X-Internal-Key.
func TestSend_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "secret-key", r.Header.Get("X-Internal-Key"))

		raw, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var req scanRequest
		require.NoError(t, json.Unmarshal(raw, &req))
		assert.Equal(t, 42, req.TransID)
		assert.Equal(t, "/app/uploads/files/1/uuid.zip", req.FilePath)

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cl := NewClient(srv.URL, "secret-key", "/app/uploads")
	err := cl.send(42, "static/files/1/uuid.zip")
	assert.NoError(t, err)
}

// send возвращает ошибку, если сканер ответил не 200.
func TestSend_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	cl := NewClient(srv.URL, "key", "/app/uploads")
	err := cl.send(1, "static/files/1/x.zip")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

// Scan запускает горутину, которая реально шлёт POST на сканер.
func TestScan_FiresRequest(t *testing.T) {
	hit := make(chan int, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req scanRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		hit <- req.TransID
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cl := NewClient(srv.URL, "key", "/app/uploads")
	cl.Scan(7, "static/files/9/a.zip")

	select {
	case tid := <-hit:
		assert.Equal(t, 7, tid)
	case <-time.After(3 * time.Second):
		t.Fatal("Scan не отправил запрос за отведённое время")
	}
}
