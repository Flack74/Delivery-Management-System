package middleware

import (
	"net/http"

	"delivery-management/internal/models"
	"github.com/gin-gonic/gin"
)

func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			switch e := err.Err.(type) {
			case *models.APIError:
				c.JSON(e.Code, e)
			default:
				apiErr := models.NewInternalError("Internal server error")
				c.JSON(http.StatusInternalServerError, apiErr)
			}
		}
	}
}
