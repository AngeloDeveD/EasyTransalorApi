package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMemoryLimiter_AllowsRequestsInsideLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := NewMemoryLimiter()
	r := gin.New()
	r.Use(limiter.Middleware(Config{
		Enabled:  true,
		Requests: 2,
		Window:   time.Minute,
	}, KeyByClientIP("test")))
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	assert.Equal(t, http.StatusOK, performRequest(r, http.MethodGet, "/ping").Code)
	assert.Equal(t, http.StatusOK, performRequest(r, http.MethodGet, "/ping").Code)
}

func TestMemoryLimiter_BlocksRequestsOverLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC)
	limiter := NewMemoryLimiter()
	limiter.now = func() time.Time { return now }

	r := gin.New()
	r.Use(limiter.Middleware(Config{
		Enabled:  true,
		Requests: 1,
		Window:   time.Minute,
	}, KeyByClientIP("test")))
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	assert.Equal(t, http.StatusOK, performRequest(r, http.MethodGet, "/ping").Code)

	w := performRequest(r, http.MethodGet, "/ping")
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, "60", w.Header().Get("Retry-After"))
}

func TestMemoryLimiter_ResetsAfterWindow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 8, 17, 10, 0, 0, 0, time.UTC)
	limiter := NewMemoryLimiter()
	limiter.now = func() time.Time { return now }

	r := gin.New()
	r.Use(limiter.Middleware(Config{
		Enabled:  true,
		Requests: 1,
		Window:   time.Minute,
	}, KeyByClientIP("test")))
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	assert.Equal(t, http.StatusOK, performRequest(r, http.MethodGet, "/ping").Code)
	assert.Equal(t, http.StatusTooManyRequests, performRequest(r, http.MethodGet, "/ping").Code)

	now = now.Add(time.Minute)
	assert.Equal(t, http.StatusOK, performRequest(r, http.MethodGet, "/ping").Code)
}

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.RemoteAddr = "192.0.2.10:1234"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
