package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"delivery-management/internal/cache"
	"delivery-management/internal/config"
	"delivery-management/internal/db"
	"delivery-management/internal/handlers"
	"delivery-management/internal/middleware"
	"delivery-management/internal/models"
	"delivery-management/internal/services"
	"delivery-management/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type OrderTestSuite struct {
	suite.Suite
	db           *db.Database
	cache        *cache.Cache
	handler      *handlers.OrderHandler
	userHandler  *handlers.UserHandler
	router       *gin.Engine
	jwtManager   *utils.JWTManager
	orderService *services.OrderService
	testUser     *models.User
	adminUser    *models.User
}

func (suite *OrderTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			Name:     "delivery_management_test",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		JWT: config.JWTConfig{
			Secret: "test-secret",
		},
	}

	var err error
	suite.db, err = db.NewDatabase(cfg)
	suite.Require().NoError(err)

	suite.cache, err = cache.NewCache(cfg)
	suite.Require().NoError(err)

	suite.jwtManager = utils.NewJWTManager(cfg)
	userService := services.NewUserService(suite.db, suite.jwtManager)
	suite.orderService = services.NewOrderService(suite.db, suite.cache)

	suite.handler = handlers.NewOrderHandler(suite.orderService)
	suite.userHandler = handlers.NewUserHandler(userService)

	suite.router = gin.New()
	suite.setupRoutes()
}

func (suite *OrderTestSuite) setupRoutes() {
	protected := suite.router.Group("/api")
	protected.Use(middleware.AuthMiddleware(suite.jwtManager))

	protected.POST("/orders", suite.handler.CreateOrder)
	protected.GET("/orders", suite.handler.GetOrders)
	protected.GET("/orders/:id", suite.handler.GetOrder)
	protected.PUT("/orders/:id/cancel", suite.handler.CancelOrder)

	admin := protected.Group("/admin")
	admin.Use(middleware.RoleMiddleware(models.RoleAdmin))
	admin.POST("/orders/:id/status", suite.handler.UpdateOrderStatus)
}

func (suite *OrderTestSuite) TearDownSuite() {
	suite.orderService.Stop()
	suite.cache.Close()
	suite.db.Close()
}

func (suite *OrderTestSuite) SetupTest() {
	// Clean up database before each test
	suite.db.Exec("DELETE FROM orders")
	suite.db.Exec("DELETE FROM users")

	// Create test users
	hashedPassword, _ := utils.HashPassword("password123")

	suite.testUser = &models.User{
		Email:    "customer@example.com",
		Password: hashedPassword,
		Role:     models.RoleCustomer,
	}
	suite.db.Create(suite.testUser)

	suite.adminUser = &models.User{
		Email:    "admin@example.com",
		Password: hashedPassword,
		Role:     models.RoleAdmin,
	}
	suite.db.Create(suite.adminUser)
}

func (suite *OrderTestSuite) getAuthToken(user *models.User) string {
	token, _ := suite.jwtManager.GenerateToken(user)
	return token
}

func (suite *OrderTestSuite) TestCreateOrder() {
	reqBody := models.CreateOrderRequest{
		Items:       "Test items",
		Description: "Test description",
		Address:     "Test address",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.getAuthToken(suite.testUser))

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Order created successfully", response["message"])
}

func (suite *OrderTestSuite) TestGetOrders() {
	// Create a test order
	order := &models.Order{
		CustomerID: suite.testUser.ID,
		Status:     models.StatusCreated,
		Items:      "Test items",
		Address:    "Test address",
	}
	suite.db.Create(order)

	req, _ := http.NewRequest("GET", "/api/orders", nil)
	req.Header.Set("Authorization", "Bearer "+suite.getAuthToken(suite.testUser))

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	orders, ok := response["orders"].([]interface{})
	assert.True(suite.T(), ok)
	assert.Len(suite.T(), orders, 1)
}

func (suite *OrderTestSuite) TestCancelOrder() {
	// Create a test order
	order := &models.Order{
		CustomerID: suite.testUser.ID,
		Status:     models.StatusCreated,
		Items:      "Test items",
		Address:    "Test address",
	}
	suite.db.Create(order)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/api/orders/%d/cancel", order.ID), nil)
	req.Header.Set("Authorization", "Bearer "+suite.getAuthToken(suite.testUser))

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify order was cancelled
	var updatedOrder models.Order
	suite.db.First(&updatedOrder, order.ID)
	assert.Equal(suite.T(), models.StatusCancelled, updatedOrder.Status)
}

func (suite *OrderTestSuite) TestAdminUpdateOrderStatus() {
	// Create a test order
	order := &models.Order{
		CustomerID: suite.testUser.ID,
		Status:     models.StatusCreated,
		Items:      "Test items",
		Address:    "Test address",
	}
	suite.db.Create(order)

	reqBody := models.UpdateOrderStatusRequest{
		Status: models.StatusDispatched,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/admin/orders/%d/status", order.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.getAuthToken(suite.adminUser))

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify order status was updated
	var updatedOrder models.Order
	suite.db.First(&updatedOrder, order.ID)
	assert.Equal(suite.T(), models.StatusDispatched, updatedOrder.Status)
}

func (suite *OrderTestSuite) TestOrderStatusTransition() {
	// Test invalid status transition
	order := &models.Order{
		CustomerID: suite.testUser.ID,
		Status:     models.StatusDelivered,
		Items:      "Test items",
		Address:    "Test address",
	}
	suite.db.Create(order)

	reqBody := models.UpdateOrderStatusRequest{
		Status: models.StatusCreated,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/admin/orders/%d/status", order.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.getAuthToken(suite.adminUser))

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func TestOrderTestSuite(t *testing.T) {
	suite.Run(t, new(OrderTestSuite))
}
