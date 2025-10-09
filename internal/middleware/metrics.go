package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Metrics struct {
	RequestCount   map[string]int64
	ResponseTime   map[string]time.Duration
	ErrorCount     map[string]int64
	ActiveRequests int64
	mu             sync.RWMutex
}

var (
	GlobalMetrics *Metrics
	metricsOnce   sync.Once
)

func getMetrics() *Metrics {
	metricsOnce.Do(func() {
		GlobalMetrics = &Metrics{
			RequestCount: make(map[string]int64),
			ResponseTime: make(map[string]time.Duration),
			ErrorCount:   make(map[string]int64),
		}
	})
	return GlobalMetrics
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		method := c.Request.Method

		metrics := getMetrics()
		metrics.mu.Lock()
		metrics.ActiveRequests++
		// Cleanup if too many entries
		if len(metrics.RequestCount) > 1000 {
			metrics.RequestCount = make(map[string]int64)
			metrics.ResponseTime = make(map[string]time.Duration)
			metrics.ErrorCount = make(map[string]int64)
		}
		metrics.mu.Unlock()

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		key := method + " " + path
		metrics.mu.Lock()
		metrics.RequestCount[key]++
		metrics.ResponseTime[key] = duration
		metrics.ActiveRequests--
		if statusCode >= 400 {
			metrics.ErrorCount[key]++
		}
		metrics.mu.Unlock()
	}
}

func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"request_count":   m.RequestCount,
		"response_times":  m.ResponseTime,
		"error_count":     m.ErrorCount,
		"active_requests": m.ActiveRequests,
	}
}
