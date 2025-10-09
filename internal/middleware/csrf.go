package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"delivery-management/internal/config"
	"github.com/gin-gonic/gin"
)

var (
	tokenStore = make(map[string]time.Time)
	storeMutex = sync.RWMutex{}
)

// CSRFMiddleware provides CSRF protection for state-changing requests
func CSRFMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF if disabled
		if !cfg.CSRF.Enabled {
			c.Next()
			return
		}

		// Skip CSRF for safe methods
		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		// Check exemption
		if isExempt(c) {
			c.Next()
			return
		}

		token := c.GetHeader("X-CSRF-Token")
		if token == "" || !validateCSRFToken(token) {
			csrfError(c)
			return
		}

		c.Next()
	}
}

// GenerateCSRFToken creates a cryptographically secure CSRF token
func GenerateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	n, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	if n != 32 {
		return "", fmt.Errorf("insufficient random bytes generated")
	}
	token := hex.EncodeToString(bytes)
	
	storeMutex.Lock()
	tokenStore[token] = time.Now().Add(time.Hour)
	storeMutex.Unlock()
	
	return token, nil
}

// isSafeMethod checks if HTTP method is safe from CSRF
func isSafeMethod(method string) bool {
	return method == "GET" || method == "HEAD" || method == "OPTIONS"
}

// isExempt checks if request is exempt from CSRF protection
func isExempt(c *gin.Context) bool {
	exempt, exists := c.Get("csrf_exempt")
	if !exists {
		return false
	}
	isExempt, ok := exempt.(bool)
	return ok && isExempt
}

// csrfError sends CSRF error response
func csrfError(c *gin.Context) {
	c.JSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF token"})
	c.Abort()
}

// validateCSRFToken validates the provided token
func validateCSRFToken(token string) bool {
	storeMutex.RLock()
	expiry, exists := tokenStore[token]
	storeMutex.RUnlock()
	
	if !exists || time.Now().After(expiry) {
		return false
	}
	
	// Remove used token
	storeMutex.Lock()
	delete(tokenStore, token)
	storeMutex.Unlock()
	
	return true
}