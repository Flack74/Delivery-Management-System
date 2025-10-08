package handlers

import (
	"net/http"
	"strconv"

	"delivery-management/internal/models"
	"delivery-management/internal/services"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "User ID not found in context",
		})
		return
	}

	customerID, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "Invalid user ID type",
		})
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), &req, customerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Failed to create order",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"order":   order,
	})
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	userRole, _ := c.Get("user_role")
	role, _ := userRole.(models.UserRole)

	var orders []*models.OrderResponse
	var err error

	if role == models.RoleAdmin {
		orders, err = h.orderService.GetAllOrders(c.Request.Context())
	} else {
		userID, _ := c.Get("user_id")
		customerID, _ := userID.(uint)
		orders, err = h.orderService.GetOrdersByCustomer(c.Request.Context(), customerID)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "Failed to get orders",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
	})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid order ID",
		})
		return
	}

	userID, _ := c.Get("user_id")
	customerID, _ := userID.(uint)

	userRole, _ := c.Get("user_role")
	role, _ := userRole.(models.UserRole)
	isAdmin := role == models.RoleAdmin

	order, err := h.orderService.GetOrderByID(c.Request.Context(), uint(orderID), customerID, isAdmin)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Order not found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid order ID",
		})
		return
	}

	userID, _ := c.Get("user_id")
	customerID, _ := userID.(uint)

	userRole, _ := c.Get("user_role")
	role, _ := userRole.(models.UserRole)
	isAdmin := role == models.RoleAdmin

	if err := h.orderService.CancelOrder(c.Request.Context(), uint(orderID), customerID, isAdmin); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Failed to cancel order",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled successfully",
	})
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid order ID",
		})
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	if err := h.orderService.UpdateOrderStatus(c.Request.Context(), uint(orderID), &req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Failed to update order status",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
	})
}

func (h *OrderHandler) GetOrderStatus(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "Invalid order ID",
		})
		return
	}

	userID, _ := c.Get("user_id")
	customerID, _ := userID.(uint)

	userRole, _ := c.Get("user_role")
	role, _ := userRole.(models.UserRole)
	isAdmin := role == models.RoleAdmin

	order, err := h.orderService.GetOrderByID(c.Request.Context(), uint(orderID), customerID, isAdmin)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   "Order not found",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id":   order.ID,
		"status":     order.Status,
		"updated_at": order.UpdatedAt,
	})
}
