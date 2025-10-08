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

var GlobalMetrics = &Metrics{
	RequestCount: make(map[string]int64),
	ResponseTime: make(map[string]time.Duration),
	ErrorCount:   make(map[string]int64),
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		method := c.Request.Method

		GlobalMetrics.mu.Lock()
		GlobalMetrics.ActiveRequests++
		GlobalMetrics.mu.Unlock()

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		GlobalMetrics.mu.Lock()
		defer GlobalMetrics.mu.Unlock()

		key := method + " " + path
		GlobalMetrics.RequestCount[key]++
		GlobalMetrics.ResponseTime[key] = duration
		GlobalMetrics.ActiveRequests--

		if statusCode >= 400 {
			GlobalMetrics.ErrorCount[key]++
		}
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
