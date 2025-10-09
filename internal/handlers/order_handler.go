package handlers

import (
	"html"
	"log"
	"net/http"

	"delivery-management/internal/models"
	"delivery-management/internal/services"
	"delivery-management/internal/utils"
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

// handleError centralizes error handling and prevents XSS
func (h *OrderHandler) handleError(c *gin.Context, err error, defaultMsg string, defaultCode int) {
	if apiErr, ok := err.(*models.APIError); ok {
		c.JSON(apiErr.Code, models.ErrorResponse{
			Error:   html.EscapeString(apiErr.Error()),
			Message: html.EscapeString(apiErr.Message),
		})
		return
	}
	log.Printf("Order handler error [%s]: %v", html.EscapeString(c.Request.URL.Path), err)
	c.Header("X-Request-ID", c.GetString("request_id"))
	c.JSON(defaultCode, models.ErrorResponse{
		Error:   html.EscapeString(defaultMsg),
		Message: "Operation failed",
	})
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Invalid request body",
			Message: "Request validation failed",
		})
		return
	}
	
	// Validate struct
	if err := utils.ValidateStruct(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}

	customerID, _, err := getUserContext(c)
	if err != nil {
		h.handleError(c, err, "Authentication failed", http.StatusUnauthorized)
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), &req, customerID)
	if err != nil {
		h.handleError(c, err, "Failed to create order", http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"order":   order,
	})
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	customerID, role, err := getUserContext(c)
	if err != nil {
		h.handleError(c, err, "Authentication failed", http.StatusUnauthorized)
		return
	}

	var orders []*models.OrderResponse
	if role == models.RoleAdmin {
		orders, err = h.orderService.GetAllOrders(c.Request.Context())
		if err != nil {
			h.handleError(c, err, "Failed to retrieve all orders", http.StatusInternalServerError)
			return
		}
	} else {
		orders, err = h.orderService.GetOrdersByCustomer(c.Request.Context(), customerID)
		if err != nil {
			h.handleError(c, err, "Failed to retrieve customer orders", http.StatusInternalServerError)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
	})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID, err := parseOrderID(c)
	if err != nil {
		h.handleError(c, err, "Invalid order ID", http.StatusBadRequest)
		return
	}

	customerID, role, err := getUserContext(c)
	if err != nil {
		h.handleError(c, err, "Authentication failed", http.StatusUnauthorized)
		return
	}

	isAdmin := role == models.RoleAdmin
	order, err := h.orderService.GetOrderByID(c.Request.Context(), orderID, customerID, isAdmin)
	if err != nil {
		h.handleError(c, err, "Order not found", http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderID, err := parseOrderID(c)
	if err != nil {
		h.handleError(c, err, "Invalid order ID", http.StatusBadRequest)
		return
	}

	customerID, role, err := getUserContext(c)
	if err != nil {
		h.handleError(c, err, "Authentication failed", http.StatusUnauthorized)
		return
	}

	isAdmin := role == models.RoleAdmin
	if err := h.orderService.CancelOrder(c.Request.Context(), orderID, customerID, isAdmin); err != nil {
		h.handleError(c, err, "Failed to cancel order", http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled successfully",
	})
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderID, err := parseOrderID(c)
	if err != nil {
		h.handleError(c, err, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleError(c, err, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.orderService.UpdateOrderStatus(c.Request.Context(), orderID, &req); err != nil {
		h.handleError(c, err, "Failed to update order status", http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
	})
}

func (h *OrderHandler) GetOrderStatus(c *gin.Context) {
	orderID, err := parseOrderID(c)
	if err != nil {
		h.handleError(c, err, "Invalid order ID", http.StatusBadRequest)
		return
	}

	customerID, role, err := getUserContext(c)
	if err != nil {
		h.handleError(c, err, "Authentication failed", http.StatusUnauthorized)
		return
	}

	isAdmin := role == models.RoleAdmin
	order, err := h.orderService.GetOrderByID(c.Request.Context(), orderID, customerID, isAdmin)
	if err != nil {
		h.handleError(c, err, "Order not found", http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id":   order.ID,
		"status":     order.Status,
		"updated_at": order.UpdatedAt,
	})
}