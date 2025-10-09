package middleware

import (
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// sanitizeLogInput removes newlines and control characters to prevent log injection
func sanitizeLogInput(input string) string {
	input = strings.ReplaceAll(input, "\n", "")
	input = strings.ReplaceAll(input, "\r", "")
	for i := 0; i < 32; i++ {
		if i != 9 { // Keep tab character
			input = strings.ReplaceAll(input, string(rune(i)), "")
		}
	}
	return input
}

func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		if param.StatusCode >= 400 {
			method := "UNKNOWN"
			path := "UNKNOWN"
			if param.Method != "" {
				method = sanitizeLogInput(param.Method)
			}
			if param.Path != "" {
				path = sanitizeLogInput(param.Path)
			}
			log.Printf("[%s] %s %s %d %s\n",
				param.TimeStamp.Format(time.RFC3339),
				method,
				path,
				param.StatusCode,
				param.Latency,
			)
		}
		return ""
	})
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			log.Printf("CORS request from origin: %s for %s %s", sanitizeLogInput(origin), sanitizeLogInput(c.Request.Method), sanitizeLogInput(c.Request.URL.Path))
		}

		allowedOrigin := "*"
		if origin != "" {
			// In production, this should be configurable
			allowedOrigin = sanitizeLogInput(origin)
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			log.Printf("CORS preflight request handled for path: %s", sanitizeLogInput(c.Request.URL.Path))
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
