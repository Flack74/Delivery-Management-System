package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"delivery-management/internal/config"
	"delivery-management/internal/db"
	"delivery-management/internal/handlers"
	"delivery-management/internal/models"
	"delivery-management/internal/services"
	"delivery-management/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserTestSuite struct {
	suite.Suite
	db         *db.Database
	handler    *handlers.UserHandler
	router     *gin.Engine
	jwtManager *utils.JWTManager
}

func (suite *UserTestSuite) SetupSuite() {
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
		JWT: config.JWTConfig{
			Secret: "test-secret",
		},
	}

	var err error
	suite.db, err = db.NewDatabase(cfg)
	suite.Require().NoError(err)

	suite.jwtManager = utils.NewJWTManager(cfg)
	userService := services.NewUserService(suite.db, suite.jwtManager)
	suite.handler = handlers.NewUserHandler(userService)

	suite.router = gin.New()
	suite.router.POST("/register", suite.handler.Register)
	suite.router.POST("/login", suite.handler.Login)
}

func (suite *UserTestSuite) TearDownSuite() {
	suite.db.Close()
}

func (suite *UserTestSuite) SetupTest() {
	// Clean up database before each test
	suite.db.Exec("DELETE FROM orders")
	suite.db.Exec("DELETE FROM users")
}

func (suite *UserTestSuite) TestUserRegistration() {
	reqBody := models.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     models.RoleCustomer,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "User registered successfully", response["message"])
}

func (suite *UserTestSuite) TestUserLogin() {
	// First register a user
	user := &models.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     models.RoleCustomer,
	}
	hashedPassword, _ := utils.HashPassword("password123")
	user.Password = hashedPassword
	suite.db.Create(user)

	reqBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), response.Token)
	assert.Equal(suite.T(), "test@example.com", response.User.Email)
}

func (suite *UserTestSuite) TestDuplicateUserRegistration() {
	// Create a user first
	user := &models.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     models.RoleCustomer,
	}
	suite.db.Create(user)

	reqBody := models.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     models.RoleCustomer,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}
