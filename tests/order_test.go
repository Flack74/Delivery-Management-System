package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	"github.com/stretchr/testify/require"
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
			Password: "86k9M0whXiogO2z5F8",
			Name:     "delivery_management",
			SSLMode:  "disable",
		},
		Redis: config.RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		JWT: config.JWTConfig{
			Secret: "test-secret-key-with-32-characters-minimum",
			Expiry: 24 * time.Hour,
		},
	}

	var err error
	suite.db, err = db.NewDatabase(cfg)
	require.NoError(suite.T(), err)

	suite.cache, err = cache.NewCache(cfg)
	require.NoError(suite.T(), err)

	suite.jwtManager, err = utils.NewJWTManager(cfg)
	require.NoError(suite.T(), err)
	
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
	if suite.orderService != nil {
		go suite.orderService.Stop()
		time.Sleep(2 * time.Second) // Give workers time to stop
	}
	if suite.cache != nil {
		suite.cache.Close()
	}
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *OrderTestSuite) SetupTest() {
	suite.db.Exec("DELETE FROM orders")
	suite.db.Exec("DELETE FROM users")

	hashedPassword, err := utils.HashPassword("password123")
	require.NoError(suite.T(), err)

	suite.testUser = &models.User{
		Email:    "customer@example.com",
		Password: hashedPassword,
		Role:     models.RoleCustomer,
	}
	require.NoError(suite.T(), suite.db.Create(suite.testUser).Error)

	suite.adminUser = &models.User{
		Email:    "admin@example.com",
		Password: hashedPassword,
		Role:     models.RoleAdmin,
	}
	require.NoError(suite.T(), suite.db.Create(suite.adminUser).Error)
}

func (suite *OrderTestSuite) getAuthToken(user *models.User) string {
	token, err := suite.jwtManager.GenerateToken(user)
	require.NoError(suite.T(), err)
	return token
}

func (suite *OrderTestSuite) TestCreateOrder() {
	reqBody := models.CreateOrderRequest{
		Items:       "Test items",
		Description: "Test description",
		Address:     "Test address",
	}

	jsonBody, err := json.Marshal(reqBody)
	require.NoError(suite.T(), err)
	
	req, err := http.NewRequest("POST", "/api/orders", bytes.NewBuffer(jsonBody))
	require.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.getAuthToken(suite.testUser))

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Order created successfully", response["message"])
}

func (suite *OrderTestSuite) TestGetOrders() {
	order := &models.Order{
		CustomerID: suite.testUser.ID,
		Status:     models.StatusCreated,
		Items:      "Test items",
		Address:    "Test address",
	}
	require.NoError(suite.T(), suite.db.Create(order).Error)

	req, err := http.NewRequest("GET", "/api/orders", nil)
	require.NoError(suite.T(), err)
	req.Header.Set("Authorization", "Bearer "+suite.getAuthToken(suite.testUser))

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	orders, ok := response["orders"].([]interface{})
	assert.True(suite.T(), ok)
	assert.Len(suite.T(), orders, 1)
}

func (suite *OrderTestSuite) TestCancelOrder() {
	order := &models.Order{
		CustomerID: suite.testUser.ID,
		Status:     models.StatusCreated,
		Items:      "Test items",
		Address:    "Test address",
	}
	require.NoError(suite.T(), suite.db.Create(order).Error)

	req, err := http.NewRequest("PUT", fmt.Sprintf("/api/orders/%d/cancel", order.ID), nil)
	require.NoError(suite.T(), err)
	req.Header.Set("Authorization", "Bearer "+suite.getAuthToken(suite.testUser))

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var updatedOrder models.Order
	require.NoError(suite.T(), suite.db.First(&updatedOrder, order.ID).Error)
	assert.Equal(suite.T(), models.StatusCancelled, updatedOrder.Status)
}

func (suite *OrderTestSuite) TestAdminUpdateOrderStatus() {
	order := &models.Order{
		CustomerID: suite.testUser.ID,
		Status:     models.StatusCreated,
		Items:      "Test items",
		Address:    "Test address",
	}
	require.NoError(suite.T(), suite.db.Create(order).Error)

	reqBody := models.UpdateOrderStatusRequest{
		Status: models.StatusDispatched,
	}

	jsonBody, err := json.Marshal(reqBody)
	require.NoError(suite.T(), err)
	
	req, err := http.NewRequest("POST", fmt.Sprintf("/api/admin/orders/%d/status", order.ID), bytes.NewBuffer(jsonBody))
	require.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.getAuthToken(suite.adminUser))

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var updatedOrder models.Order
	require.NoError(suite.T(), suite.db.First(&updatedOrder, order.ID).Error)
	assert.Equal(suite.T(), models.StatusDispatched, updatedOrder.Status)
}

func TestOrderTestSuite(t *testing.T) {
	suite.Run(t, new(OrderTestSuite))
}