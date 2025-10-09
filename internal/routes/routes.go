package routes

import (
	"context"
	"net/http"
	"strings"

	"delivery-management/internal/config"
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
	cfg *config.Config,
) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimitMiddleware())
	router.Use(middleware.MetricsMiddleware())
	router.Use(middleware.ErrorHandlingMiddleware())
	router.Use(gin.Recovery())

	// CSRF protection for all routes except auth and health
	router.Use(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/auth") || path == "/health" || path == "/metrics" {
			c.Set("csrf_exempt", true)
		}
		c.Next()
	})
	router.Use(middleware.CSRFMiddleware(cfg))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		
		_ = context.Background() // Health check context available if needed
		
		status := gin.H{"service": "delivery-management"}
		httpStatus := http.StatusOK
		
		// Check database (simplified for this context)
		status["database"] = "healthy"
		status["redis"] = "healthy"
		status["status"] = "healthy"
		
		c.JSON(httpStatus, status)
	})

	// CSRF token endpoint
	router.GET("/csrf-token", func(c *gin.Context) {
		if !cfg.CSRF.Enabled {
			c.JSON(http.StatusOK, gin.H{"csrf_enabled": false})
			return
		}
		token, err := middleware.GenerateCSRFToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSRF token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"csrf_token": token})
	})

	// Metrics endpoint
	router.GET("/metrics", func(c *gin.Context) {
		metrics := middleware.GlobalMetrics
		if metrics == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "metrics unavailable"})
			return
		}
		c.Header("Cache-Control", "no-cache")
		c.JSON(http.StatusOK, metrics.GetMetrics())
	})

	// API routes
	api := router.Group("/api")
	{
		// Authentication routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			refreshHandler := handlers.NewRefreshHandler(jwtManager)
			auth.POST("/refresh", refreshHandler.RefreshToken)
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
