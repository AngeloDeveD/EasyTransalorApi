package ratelimit

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Enabled  bool
	Requests int
	Window   time.Duration
}

type bucket struct {
	count int
	reset time.Time
}

type MemoryLimiter struct {
	mu      sync.Mutex
	buckets map[string]bucket
	now     func() time.Time
}

func NewMemoryLimiter() *MemoryLimiter {
	return &MemoryLimiter{
		buckets: make(map[string]bucket),
		now:     time.Now,
	}
}

func (l *MemoryLimiter) Middleware(cfg Config, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled || cfg.Requests <= 0 || cfg.Window <= 0 {
			c.Next()
			return
		}

		key := keyFunc(c)
		if key == "" {
			key = c.ClientIP()
		}

		allowed, remaining, retryAfter := l.allow(key, cfg.Requests, cfg.Window)
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Слишком много запросов. Попробуйте позже.",
				"retryAfter":  int(retryAfter.Seconds()),
				"rateLimited": true,
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.Requests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Next()
	}
}

func (l *MemoryLimiter) allow(key string, maxRequests int, window time.Duration) (bool, int, time.Duration) {
	now := l.now()

	l.mu.Lock()
	defer l.mu.Unlock()

	b, exists := l.buckets[key]
	if !exists || !now.Before(b.reset) {
		l.buckets[key] = bucket{count: 1, reset: now.Add(window)}
		return true, maxRequests - 1, 0
	}

	if b.count >= maxRequests {
		return false, 0, b.reset.Sub(now)
	}

	b.count++
	l.buckets[key] = b
	return true, maxRequests - b.count, 0
}

func KeyByClientIP(prefix string) func(*gin.Context) string {
	return func(c *gin.Context) string {
		return prefix + ":ip:" + c.ClientIP()
	}
}

func KeyByUserOrIP(prefix string) func(*gin.Context) string {
	return func(c *gin.Context) string {
		if userID, exists := c.Get("userID"); exists {
			return prefix + ":user:" + fmt.Sprint(userID) + ":" + c.FullPath()
		}
		return prefix + ":ip:" + c.ClientIP() + ":" + c.FullPath()
	}
}
