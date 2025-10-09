package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"delivery-management/internal/cache"
	"delivery-management/internal/config"
	"delivery-management/internal/db"
	"delivery-management/internal/handlers"
	"delivery-management/internal/routes"
	"delivery-management/internal/services"
	"delivery-management/internal/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Initialize database
	database, err := db.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Initialize Redis cache
	redisCache, err := cache.NewCache(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Redis cache: %v", err)
	}
	defer func() {
		if err := redisCache.Close(); err != nil {
			log.Printf("Failed to close Redis cache: %v", err)
		}
	}()

	// Initialize JWT manager
	jwtManager, err := utils.NewJWTManager(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize JWT manager: %v", err)
	}

	// Initialize services
	userService := services.NewUserService(database, jwtManager)
	orderService := services.NewOrderService(database, redisCache)
	if orderService == nil {
		log.Fatalf("Failed to initialize order service")
	}

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	orderHandler := handlers.NewOrderHandler(orderService)

	// Setup routes
	router := routes.SetupRoutes(userHandler, orderHandler, jwtManager, cfg)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server on port %d: %v", cfg.Server.Port, err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	// Stop the order service after the server has shut down
	orderService.Stop()

	log.Println("Server exited")
}
