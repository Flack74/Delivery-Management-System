package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		if param.StatusCode >= 400 {
			log.Printf("[%s] %s %s %d %s\n",
				param.TimeStamp.Format(time.RFC3339),
				param.Method,
				param.Path,
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
			log.Printf("CORS request from origin: [REDACTED] for %s %s", c.Request.Method, c.Request.URL.Path)
		}
		
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			log.Printf("CORS preflight request handled for path: %s", c.Request.URL.Path)
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
