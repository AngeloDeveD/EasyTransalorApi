package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIDMiddlewareAddsHeader(t *testing.T) {
	r := gin.New()
	r.Use(requestIDMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get(requestIDHeader))
}

func TestRequestIDMiddlewarePreservesIncomingHeader(t *testing.T) {
	r := gin.New()
	r.Use(requestIDMiddleware())
	r.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(requestIDHeader, "test-request-id")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "test-request-id", w.Header().Get(requestIDHeader))
}

func TestRecoveryLoggerMiddlewareReturnsRequestID(t *testing.T) {
	r := gin.New()
	r.Use(requestIDMiddleware())
	r.Use(requestLoggerMiddleware())
	r.Use(recoveryLoggerMiddleware())
	r.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set(requestIDHeader, "panic-request-id")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var body map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "panic-request-id", body["requestId"])
	assert.Equal(t, "panic-request-id", w.Header().Get(requestIDHeader))
	assert.NotEmpty(t, body["error"])
}
