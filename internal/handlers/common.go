package handlers

import (
	"strconv"

	"delivery-management/internal/models"
	"github.com/gin-gonic/gin"
)

func getUserContext(c *gin.Context) (uint, models.UserRole, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, "", models.NewUnauthorizedError("User context not found")
	}

	id, ok := userID.(uint)
	if !ok {
		return 0, "", models.NewInternalError("Invalid user context")
	}

	userRole, _ := c.Get("user_role")
	role, _ := userRole.(models.UserRole)

	return id, role, nil
}

func parseOrderID(c *gin.Context) (uint, error) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		return 0, models.NewBadRequestError("Order ID must be a valid number")
	}
	return uint(orderID), nil
}