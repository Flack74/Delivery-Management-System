package tests

import (
	"context"
	"fmt"
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
	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}
	cfg := &config.Config{
		Database: config.DatabaseConfig{Host: "localhost", Port: 5432, User: "postgres", Password: "86k9M0whXiogO2z5F8", Name: "delivery_management", SSLMode: "disable"},
		Redis:    config.RedisConfig{Host: "localhost", Port: 6379},
		JWT:      config.JWTConfig{Secret: "test-secret-key-with-32-characters-minimum"},
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

	// Create test users
	var userIDs []uint
	for i := 0; i < 10; i++ {
		user := &models.User{
			Email:    fmt.Sprintf("testuser%d_%d@example.com", i, time.Now().UnixNano()),
			Password: "hashed",
			Role:     models.RoleCustomer,
		}
		err := database.Create(user).Error
		if err != nil {
			t.Skipf("Failed to create test user: %v", err)
		}
		userIDs = append(userIDs, user.ID)
	}

	var wg sync.WaitGroup

	for i, userID := range userIDs {
		wg.Add(1)
		go func(id int, uid uint) {
			defer wg.Done()
			req := &models.CreateOrderRequest{
				Items:   "Test items",
				Address: "Test address",
			}
			_, err := orderService.CreateOrder(context.Background(), req, uid)
			assert.NoError(t, err)
		}(i, userID)
	}

	wg.Wait()
}

func TestRedisIntegration(t *testing.T) {
	cfg := &config.Config{
		Redis: config.RedisConfig{Host: "localhost", Port: 6379},
	}

	redisCache, err := cache.NewCache(cfg)
	if err != nil {
		t.Skipf("Redis not available for integration test: %v", err)
	}
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
		Database: config.DatabaseConfig{Host: "localhost", Port: 5432, User: "postgres", Password: "86k9M0whXiogO2z5F8", Name: "delivery_management", SSLMode: "disable"},
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
	user := &models.User{Email: fmt.Sprintf("dbtest_%d@example.com", time.Now().UnixNano()), Password: "hashed", Role: models.RoleCustomer}
	err = database.Create(user).Error
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
}
