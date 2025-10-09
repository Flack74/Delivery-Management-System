package tests

import (
	"testing"

	"delivery-management/internal/models"
	"delivery-management/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestInputValidation(t *testing.T) {
	// Test XSS prevention
	maliciousInput := "<script>alert('xss')</script>"
	sanitized := utils.SanitizeInput(maliciousInput)
	assert.NotContains(t, sanitized, "<script>")
	assert.Contains(t, sanitized, "&lt;script&gt;")

	// Test SQL injection detection
	sqlInjection := "'; DROP TABLE users; --"
	assert.True(t, utils.DetectSQLInjection(sqlInjection))

	// Test email validation
	assert.True(t, utils.ValidateEmail("test@example.com"))
	assert.False(t, utils.ValidateEmail("invalid-email"))
	assert.False(t, utils.ValidateEmail(""))

	// Test password validation
	assert.True(t, utils.ValidatePassword("password123"))
	assert.False(t, utils.ValidatePassword("short"))
	assert.False(t, utils.ValidatePassword(""))
}

func TestErrorHandling(t *testing.T) {
	// Test APIError creation
	err := models.NewBadRequestError("test error")
	assert.Equal(t, 400, err.Code)
	assert.Equal(t, "test error", err.Message)

	err = models.NewUnauthorizedError("unauthorized")
	assert.Equal(t, 401, err.Code)

	err = models.NewNotFoundError("not found")
	assert.Equal(t, 404, err.Code)

	err = models.NewInternalError("internal error")
	assert.Equal(t, 500, err.Code)
}

func TestOrderStatusEdgeCases(t *testing.T) {
	// Test invalid status transitions
	assert.False(t, models.StatusDelivered.CanTransitionTo(models.StatusCreated))
	assert.False(t, models.StatusCancelled.CanTransitionTo(models.StatusDispatched))
	assert.False(t, models.StatusInTransit.CanTransitionTo(models.StatusCancelled))

	// Test final status detection
	assert.True(t, models.StatusDelivered.IsFinal())
	assert.True(t, models.StatusCancelled.IsFinal())
	assert.False(t, models.StatusCreated.IsFinal())

	// Test invalid status validation
	invalidStatus := models.OrderStatus("invalid")
	assert.False(t, invalidStatus.IsValid())
}

func TestConcurrencyEdgeCases(t *testing.T) {
	// Test empty string handling
	assert.Equal(t, "", utils.SanitizeInput(""))
	
	// Test very long input
	longInput := make([]byte, 2000)
	for i := range longInput {
		longInput[i] = 'a'
	}
	sanitized := utils.SanitizeInput(string(longInput))
	assert.LessOrEqual(t, len(sanitized), 1000)
}