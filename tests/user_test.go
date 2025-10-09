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
	"github.com/stretchr/testify/require"
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
	require.NoError(suite.T(), err)

	suite.jwtManager, err = utils.NewJWTManager(cfg)
	require.NoError(suite.T(), err)
	
	userService := services.NewUserService(suite.db, suite.jwtManager)
	suite.handler = handlers.NewUserHandler(userService)

	suite.router = gin.New()
	suite.router.POST("/register", suite.handler.Register)
	suite.router.POST("/login", suite.handler.Login)
}

func (suite *UserTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *UserTestSuite) SetupTest() {
	suite.db.Exec("DELETE FROM orders")
	suite.db.Exec("DELETE FROM users")
}

func (suite *UserTestSuite) TestUserRegistration() {
	reqBody := models.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     models.RoleCustomer,
	}

	jsonBody, err := json.Marshal(reqBody)
	require.NoError(suite.T(), err)
	
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	require.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "User registered successfully", response["message"])
}

func (suite *UserTestSuite) TestUserLogin() {
	hashedPassword, err := utils.HashPassword("password123")
	require.NoError(suite.T(), err)
	
	user := &models.User{
		Email:    "test@example.com",
		Password: hashedPassword,
		Role:     models.RoleCustomer,
	}
	require.NoError(suite.T(), suite.db.Create(user).Error)

	reqBody := models.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	jsonBody, err := json.Marshal(reqBody)
	require.NoError(suite.T(), err)
	
	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	require.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.LoginResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), response.Token)
	assert.Equal(suite.T(), "test@example.com", response.User.Email)
}

func (suite *UserTestSuite) TestDuplicateUserRegistration() {
	user := &models.User{
		Email:    "test@example.com",
		Password: "hashedpassword",
		Role:     models.RoleCustomer,
	}
	require.NoError(suite.T(), suite.db.Create(user).Error)

	reqBody := models.CreateUserRequest{
		Email:    "test@example.com",
		Password: "password123",
		Role:     models.RoleCustomer,
	}

	jsonBody, err := json.Marshal(reqBody)
	require.NoError(suite.T(), err)
	
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	require.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, new(UserTestSuite))
}