package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"delivery-management/internal/cache"
	"delivery-management/internal/config"
	"delivery-management/internal/db"
	"delivery-management/internal/models"
	"delivery-management/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestConcurrentOrderProcessing(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{Host: "localhost", Port: 5432, User: "postgres", Password: "password", Name: "test_db", SSLMode: "disable"},
		Redis:    config.RedisConfig{Host: "localhost", Port: 6379},
		JWT:      config.JWTConfig{Secret: "test-secret"},
	}

	database, err := db.NewDatabase(cfg)
	if err != nil {
		t.Skipf("Database not available for integration test: %v", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			t.Logf("Failed to close database: %v", err)
		}
	}()
	
	redisCache, err := cache.NewCache(cfg)
	if err != nil {
		t.Skip("Redis not available for integration test")
	}
	defer redisCache.Close()
	orderService := services.NewOrderService(database, redisCache)
	defer orderService.Stop()

	var wg sync.WaitGroup
	orderCount := 10

	for i := 0; i < orderCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			req := &models.CreateOrderRequest{
				Items:   "Test items",
				Address: "Test address",
			}
			_, err := orderService.CreateOrder(context.Background(), req, uint(id))
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()
}

func TestRedisIntegration(t *testing.T) {
	cfg := &config.Config{
		Redis: config.RedisConfig{Host: "localhost", Port: 6379},
	}

	redisCache, err := cache.NewCache(cfg)
	assert.NoError(t, err)
	defer redisCache.Close()

	ctx := context.Background()

	// Test set/get
	err = redisCache.Set(ctx, "test-key", "test-value", time.Minute)
	assert.NoError(t, err)

	value, err := redisCache.Get(ctx, "test-key")
	assert.NoError(t, err)
	assert.Equal(t, "test-value", value)

	// Test pub/sub
	err = redisCache.Publish(ctx, "test-channel", "test-message")
	assert.NoError(t, err)
}

func TestDatabaseIntegration(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{Host: "localhost", Port: 5432, User: "postgres", Password: "password", Name: "test_db", SSLMode: "disable"},
	}

	database, err := db.NewDatabase(cfg)
	if err != nil {
		t.Skipf("Database not available for integration test: %v", err)
		return
	}
	defer func() {
		if database != nil {
			database.Close()
		}
	}()

	// Test health check
	err = database.HealthCheck(context.Background())
	assert.NoError(t, err)

	// Test user creation
	user := &models.User{Email: "test@example.com", Password: "hashed", Role: models.RoleCustomer}
	err = database.Create(user).Error
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
}
