package utils

import (
	"html"
	"regexp"
	"strings"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	sqlRegex   = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|script)`)
)

// SanitizeString removes potentially dangerous characters and HTML
func SanitizeString(input string) string {
	// Remove HTML tags and escape HTML entities
	sanitized := html.EscapeString(strings.TrimSpace(input))
	// Remove null bytes
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")
	return sanitized
}

// ValidateEmail checks if email format is valid
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email) && len(email) <= 254
}

// ValidatePassword checks password strength
func ValidatePassword(password string) bool {
	return len(password) >= 8 && len(password) <= 128
}

// DetectSQLInjection checks for basic SQL injection patterns
func DetectSQLInjection(input string) bool {
	return sqlRegex.MatchString(input)
}

// SanitizeInput performs comprehensive input sanitization
func SanitizeInput(input string) string {
	if input == "" {
		return ""
	}
	
	// Basic sanitization
	sanitized := SanitizeString(input)
	
	// Limit length
	if len(sanitized) > 1000 {
		sanitized = sanitized[:1000]
	}
	
	return sanitized
}