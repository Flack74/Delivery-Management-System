package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"delivery-management/internal/cache"
	"delivery-management/internal/config"
	"delivery-management/internal/db"
	"delivery-management/internal/models"
	"delivery-management/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderService_CreateOrder(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "86k9M0whXiogO2z5F8",
			Name:     "delivery_management",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
	}

	database, err := db.NewDatabase(cfg)
	require.NoError(t, err)
	defer database.Close()

	redisCache, err := cache.NewCache(cfg)
	require.NoError(t, err)
	defer redisCache.Close()

	orderService := services.NewOrderService(database, redisCache)
	defer orderService.Stop()

	// Create test user
	user := &models.User{
		Email:    fmt.Sprintf("test_%d@example.com", time.Now().UnixNano()),
		Password: "hashedpassword",
		Role:     models.RoleCustomer,
	}
	err = database.WithContext(context.Background()).Create(user).Error
	require.NoError(t, err)

	req := &models.CreateOrderRequest{
		Items:       "Test items",
		Description: "Test description",
		Address:     "Test address",
	}

	order, err := orderService.CreateOrder(context.Background(), req, user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, order)
	assert.Equal(t, models.StatusCreated, order.Status)
	assert.Equal(t, user.ID, order.CustomerID)
}

func TestOrderService_GetOrdersByCustomer(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "86k9M0whXiogO2z5F8",
			Name:     "delivery_management",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
	}

	database, err := db.NewDatabase(cfg)
	require.NoError(t, err)
	defer database.Close()

	redisCache, err := cache.NewCache(cfg)
	require.NoError(t, err)
	defer redisCache.Close()

	orderService := services.NewOrderService(database, redisCache)
	defer orderService.Stop()

	// Create test user
	user := &models.User{
		Email:    fmt.Sprintf("test2_%d@example.com", time.Now().UnixNano()),
		Password: "hashedpassword",
		Role:     models.RoleCustomer,
	}
	err = database.WithContext(context.Background()).Create(user).Error
	require.NoError(t, err)

	// Create test order
	order := &models.Order{
		CustomerID:  user.ID,
		Status:      models.StatusCreated,
		Items:       "Test items",
		Description: "Test description",
		Address:     "Test address",
	}
	err = database.WithContext(context.Background()).Create(order).Error
	require.NoError(t, err)

	orders, err := orderService.GetOrdersByCustomer(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Len(t, orders, 1)
	assert.Equal(t, order.ID, orders[0].ID)
}

func TestOrderService_CancelOrder(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "86k9M0whXiogO2z5F8",
			Name:     "delivery_management",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
	}

	database, err := db.NewDatabase(cfg)
	require.NoError(t, err)
	defer database.Close()

	redisCache, err := cache.NewCache(cfg)
	require.NoError(t, err)
	defer redisCache.Close()

	orderService := services.NewOrderService(database, redisCache)
	defer orderService.Stop()

	// Create test user
	user := &models.User{
		Email:    fmt.Sprintf("test3_%d@example.com", time.Now().UnixNano()),
		Password: "hashedpassword",
		Role:     models.RoleCustomer,
	}
	err = database.WithContext(context.Background()).Create(user).Error
	require.NoError(t, err)

	// Create test order
	order := &models.Order{
		CustomerID:  user.ID,
		Status:      models.StatusCreated,
		Items:       "Test items",
		Description: "Test description",
		Address:     "Test address",
	}
	err = database.WithContext(context.Background()).Create(order).Error
	require.NoError(t, err)

	err = orderService.CancelOrder(context.Background(), order.ID, user.ID, false)
	assert.NoError(t, err)

	// Verify order is cancelled
	var updatedOrder models.Order
	err = database.WithContext(context.Background()).First(&updatedOrder, order.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.StatusCancelled, updatedOrder.Status)
}

func TestOrderService_UpdateOrderStatus(t *testing.T) {
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "86k9M0whXiogO2z5F8",
			Name:     "delivery_management",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
	}

	database, err := db.NewDatabase(cfg)
	require.NoError(t, err)
	defer database.Close()

	redisCache, err := cache.NewCache(cfg)
	require.NoError(t, err)
	defer redisCache.Close()

	orderService := services.NewOrderService(database, redisCache)
	defer orderService.Stop()

	// Create test user
	user := &models.User{
		Email:    fmt.Sprintf("test4_%d@example.com", time.Now().UnixNano()),
		Password: "hashedpassword",
		Role:     models.RoleAdmin,
	}
	err = database.WithContext(context.Background()).Create(user).Error
	require.NoError(t, err)

	// Create test order
	order := &models.Order{
		CustomerID:  user.ID,
		Status:      models.StatusCreated,
		Items:       "Test items",
		Description: "Test description",
		Address:     "Test address",
	}
	err = database.WithContext(context.Background()).Create(order).Error
	require.NoError(t, err)

	req := &models.UpdateOrderStatusRequest{
		Status: models.StatusDispatched,
	}

	err = orderService.UpdateOrderStatus(context.Background(), order.ID, req)
	assert.NoError(t, err)

	// Verify order status is updated
	var updatedOrder models.Order
	err = database.WithContext(context.Background()).First(&updatedOrder, order.ID).Error
	require.NoError(t, err)
	assert.Equal(t, models.StatusDispatched, updatedOrder.Status)
}

func TestOrderProcessor_DatabasePersistence(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "86k9M0whXiogO2z5F8",
			Name:     "delivery_management",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
	}

	database, err := db.NewDatabase(cfg)
	require.NoError(t, err)
	defer database.Close()

	redisCache, err := cache.NewCache(cfg)
	require.NoError(t, err)
	defer redisCache.Close()

	orderService := services.NewOrderService(database, redisCache)
	defer orderService.Stop()

	// Create test user
	user := &models.User{
		Email:    fmt.Sprintf("test5_%d@example.com", time.Now().UnixNano()),
		Password: "hashedpassword",
		Role:     models.RoleCustomer,
	}
	err = database.WithContext(context.Background()).Create(user).Error
	require.NoError(t, err)

	req := &models.CreateOrderRequest{
		Items:       "Test items for processing",
		Description: "Test concurrent processing",
		Address:     "Test address",
	}

	order, err := orderService.CreateOrder(context.Background(), req, user.ID)
	require.NoError(t, err)

	// Wait for first status transition (60s + buffer)
	time.Sleep(65 * time.Second)

	// Verify order status was updated in database
	var updatedOrder models.Order
	err = database.WithContext(context.Background()).First(&updatedOrder, order.ID).Error
	require.NoError(t, err)
	
	// Order should have progressed from "created" status
	assert.NotEqual(t, models.StatusCreated, updatedOrder.Status)
	assert.Contains(t, []models.OrderStatus{
		models.StatusDispatched,
		models.StatusInTransit,
		models.StatusDelivered,
	}, updatedOrder.Status)
}