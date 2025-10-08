package handlers

import (
	"net/http"
	"strings"

	"delivery-management/internal/models"
	"delivery-management/internal/utils"
	"github.com/gin-gonic/gin"
)

type RefreshHandler struct {
	jwtManager *utils.JWTManager
}

func NewRefreshHandler(jwtManager *utils.JWTManager) *RefreshHandler {
	return &RefreshHandler{
		jwtManager: jwtManager,
	}
}

func (h *RefreshHandler) RefreshToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Authorization header required",
		})
		return
	}

	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "Invalid authorization header format",
		})
		return
	}

	newToken, err := h.jwtManager.RefreshToken(tokenParts[1])
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "Token refresh failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": newToken,
	})
}
