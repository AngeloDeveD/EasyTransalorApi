package main

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set("requestID", requestID)
		c.Writer.Header().Set(requestIDHeader, requestID)
		c.Next()
	}
}

func requestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := c.Writer.Status()
		attrs := []slog.Attr{
			slog.String("request_id", requestIDFromContext(c)),
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.Int("status", status),
			slog.Duration("latency", time.Since(start)),
			slog.String("client_ip", c.ClientIP()),
			slog.String("user_agent", c.Request.UserAgent()),
		}

		if userID, exists := c.Get("userID"); exists {
			attrs = append(attrs, slog.Any("user_id", userID))
		}

		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("gin_errors", c.Errors.String()))
		}

		if status >= http.StatusInternalServerError {
			slog.LogAttrs(c.Request.Context(), slog.LevelError, "request failed", attrs...)
			return
		}

		slog.LogAttrs(c.Request.Context(), slog.LevelInfo, "request completed", attrs...)
	}
}

func recoveryLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				requestID := requestIDFromContext(c)
				slog.ErrorContext(
					c.Request.Context(),
					"panic recovered",
					slog.String("request_id", requestID),
					slog.String("method", c.Request.Method),
					slog.String("path", c.FullPath()),
					slog.String("client_ip", c.ClientIP()),
					slog.Any("panic", recovered),
					slog.String("stack", string(debug.Stack())),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":     "Внутренняя ошибка сервера",
					"requestId": requestID,
				})
			}
		}()

		c.Next()
	}
}

func requestIDFromContext(c *gin.Context) string {
	if requestID, exists := c.Get("requestID"); exists {
		if value, ok := requestID.(string); ok {
			return value
		}
	}
	return c.Writer.Header().Get(requestIDHeader)
}
