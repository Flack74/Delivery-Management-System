package routes

import (
	"net/http"

	"delivery-management/internal/handlers"
	"delivery-management/internal/middleware"
	"delivery-management/internal/models"
	"delivery-management/internal/utils"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	userHandler *handlers.UserHandler,
	orderHandler *handlers.OrderHandler,
	jwtManager *utils.JWTManager,
) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "delivery-management",
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Authentication routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			// User profile
			protected.GET("/profile", userHandler.GetProfile)

			// Order routes
			orders := protected.Group("/orders")
			{
				orders.POST("", orderHandler.CreateOrder)
				orders.GET("", orderHandler.GetOrders)
				orders.GET("/:id", orderHandler.GetOrder)
				orders.PUT("/:id/cancel", orderHandler.CancelOrder)
				orders.GET("/:id/status", orderHandler.GetOrderStatus)
			}

			// Admin-only routes
			admin := protected.Group("/admin")
			admin.Use(middleware.RoleMiddleware(models.RoleAdmin))
			{
				admin.GET("/orders", orderHandler.GetOrders)
				admin.PUT("/orders/:id", orderHandler.UpdateOrderStatus)
				admin.POST("/orders/:id/status", orderHandler.UpdateOrderStatus)
			}
		}
	}

	return router
}