package tests

import (
	"testing"

	"delivery-management/internal/models"
	"delivery-management/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestOrderStatusTransitions(t *testing.T) {
	// Test valid transitions
	assert.True(t, models.StatusCreated.CanTransitionTo(models.StatusDispatched))
	assert.True(t, models.StatusCreated.CanTransitionTo(models.StatusCancelled))
	assert.True(t, models.StatusDispatched.CanTransitionTo(models.StatusInTransit))
	assert.True(t, models.StatusDispatched.CanTransitionTo(models.StatusCancelled))
	assert.True(t, models.StatusInTransit.CanTransitionTo(models.StatusDelivered))

	// Test invalid transitions
	assert.False(t, models.StatusDelivered.CanTransitionTo(models.StatusCreated))
	assert.False(t, models.StatusCancelled.CanTransitionTo(models.StatusDispatched))
	assert.False(t, models.StatusInTransit.CanTransitionTo(models.StatusCreated))
}

func TestOrderStatusValidation(t *testing.T) {
	// Test valid statuses
	assert.True(t, models.StatusCreated.IsValid())
	assert.True(t, models.StatusDispatched.IsValid())
	assert.True(t, models.StatusInTransit.IsValid())
	assert.True(t, models.StatusDelivered.IsValid())
	assert.True(t, models.StatusCancelled.IsValid())

	// Test invalid status
	assert.False(t, models.OrderStatus("invalid").IsValid())
}

func TestOrderStatusFinal(t *testing.T) {
	// Test final statuses
	assert.True(t, models.StatusDelivered.IsFinal())
	assert.True(t, models.StatusCancelled.IsFinal())

	// Test non-final statuses
	assert.False(t, models.StatusCreated.IsFinal())
	assert.False(t, models.StatusDispatched.IsFinal())
	assert.False(t, models.StatusInTransit.IsFinal())
}

func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"

	// Test password hashing
	hashedPassword, err := utils.HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)

	// Test password verification
	assert.True(t, utils.CheckPasswordHash(password, hashedPassword))
	assert.False(t, utils.CheckPasswordHash("wrongpassword", hashedPassword))
}

func TestUserResponseConversion(t *testing.T) {
	user := &models.User{
		ID:    1,
		Email: "test@example.com",
		Role:  models.RoleCustomer,
	}

	response := user.ToResponse()
	assert.Equal(t, user.ID, response.ID)
	assert.Equal(t, user.Email, response.Email)
	assert.Equal(t, user.Role, response.Role)
}

func TestOrderResponseConversion(t *testing.T) {
	order := &models.Order{
		ID:          1,
		CustomerID:  1,
		Status:      models.StatusCreated,
		Items:       "Test items",
		Description: "Test description",
		Address:     "Test address",
	}

	response := order.ToResponse()
	assert.Equal(t, order.ID, response.ID)
	assert.Equal(t, order.CustomerID, response.CustomerID)
	assert.Equal(t, order.Status, response.Status)
	assert.Equal(t, order.Items, response.Items)
	assert.Equal(t, order.Description, response.Description)
	assert.Equal(t, order.Address, response.Address)
}
