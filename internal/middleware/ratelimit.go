package middleware

import (
	"net/http"
	"sync"
	"time"

	"delivery-management/internal/models"
	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	clients map[string]*ClientInfo
	mu      sync.RWMutex
	limit   int
	window  time.Duration
}

type ClientInfo struct {
	requests  int
	resetTime time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*ClientInfo),
		limit:   limit,
		window:  window,
	}

	// Cleanup expired entries every minute
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, info := range rl.clients {
			if now.After(info.resetTime) {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientIP]

	if !exists || now.After(client.resetTime) {
		rl.clients[clientIP] = &ClientInfo{
			requests:  1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	if client.requests >= rl.limit {
		return false
	}

	client.requests++
	return true
}

var defaultRateLimiter = NewRateLimiter(100, time.Minute)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !defaultRateLimiter.Allow(clientIP) {
			apiErr := models.NewAPIError(
				http.StatusTooManyRequests,
				"Rate limit exceeded",
				"Too many requests from this IP address",
				"RATE_LIMIT_EXCEEDED",
			)
			c.JSON(http.StatusTooManyRequests, apiErr)
			c.Abort()
			return
		}

		c.Next()
	}
}
