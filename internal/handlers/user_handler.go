package handlers

import (
	"html"
	"net/http"

	"delivery-management/internal/models"
	"delivery-management/internal/services"
	"delivery-management/internal/utils"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Header("X-Request-ID", c.GetString("request_id"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   html.EscapeString("Invalid request body"),
			Message: html.EscapeString("Request validation failed"),
		})
		return
	}
	
	// Validate struct
	if err := utils.ValidateStruct(&req); err != nil {
		c.Header("X-Request-ID", c.GetString("request_id"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   html.EscapeString("Validation failed"),
			Message: html.EscapeString(err.Error()),
		})
		return
	}

	user, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		c.Header("X-Request-ID", c.GetString("request_id"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   html.EscapeString("Registration failed"),
			Message: html.EscapeString("Unable to create user account"),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Header("X-Request-ID", c.GetString("request_id"))
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   html.EscapeString("Invalid request body"),
			Message: html.EscapeString("Request validation failed"),
		})
		return
	}

	response, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		c.Header("X-Request-ID", c.GetString("request_id"))
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   html.EscapeString("Login failed"),
			Message: html.EscapeString("Invalid credentials"),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: html.EscapeString("User ID not found in context"),
		})
		return
	}

	id, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: html.EscapeString("Invalid user ID type"),
		})
		return
	}

	user, err := h.userService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error:   html.EscapeString("User not found"),
			Message: html.EscapeString("User profile not available"),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}