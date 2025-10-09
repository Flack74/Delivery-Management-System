package middleware

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	DefaultLimit  = 100
	DefaultWindow = time.Minute
)

type RateLimiter struct {
	rate   int64
	period int64
}

type bucket struct {
	tokens   int64
	lastSeen int64
}

var (
	buckets    sync.Map
	bucketSize int64 = 10000
)

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rate:   int64(limit),
		period: int64(window),
	}
}

func (rl *RateLimiter) Allow(clientIP string) bool {
	now := time.Now().UnixNano()

	// Cleanup with proper synchronization
	if atomic.LoadInt64(&bucketSize) > 10000 {
		go func() {
			buckets.Range(func(key, value interface{}) bool {
				b := value.(*bucket)
				if now-atomic.LoadInt64(&b.lastSeen) > int64(5*time.Minute) {
					buckets.Delete(key)
					atomic.AddInt64(&bucketSize, -1)
				}
				return atomic.LoadInt64(&bucketSize) > 5000
			})
		}()
	}

	value, loaded := buckets.LoadOrStore(clientIP, &bucket{
		tokens:   rl.rate,
		lastSeen: now,
	})
	if !loaded {
		atomic.AddInt64(&bucketSize, 1)
	}

	b := value.(*bucket)
	last := atomic.LoadInt64(&b.lastSeen)
	elapsed := now - last

	if elapsed > rl.period {
		atomic.StoreInt64(&b.tokens, rl.rate)
		atomic.StoreInt64(&b.lastSeen, now)
		return atomic.AddInt64(&b.tokens, -1) >= 0
	}

	return atomic.AddInt64(&b.tokens, -1) >= 0
}

var (
	limiter = NewRateLimiter(DefaultLimit, DefaultWindow)
	rateLimitResponse = gin.H{
		"error": "Rate limit exceeded",
		"code":  "RATE_LIMIT_EXCEEDED",
	}
)

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if !limiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, rateLimitResponse)
			c.Abort()
			return
		}
		c.Next()
	}
}
