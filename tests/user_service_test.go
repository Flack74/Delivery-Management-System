package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"delivery-management/internal/config"
	"delivery-management/internal/db"
	"delivery-management/internal/models"
	"delivery-management/internal/services"
	"delivery-management/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestConfig() *config.Config {
	return &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "86k9M0whXiogO2z5F8",
			Name:     "delivery_management",
		},
		JWT: config.JWTConfig{
			Secret: "test-secret-key-with-32-characters-minimum",
			Expiry: 24 * time.Hour,
		},
	}
}

func setupUserService(t *testing.T) (*services.UserService, func()) {
	cfg := createTestConfig()

	database, err := db.NewDatabase(cfg)
	require.NoError(t, err)

	jwtManager, err := utils.NewJWTManager(cfg)
	require.NoError(t, err)

	userService := services.NewUserService(database, jwtManager)

	cleanup := func() {
		database.Close()
	}

	return userService, cleanup
}

func TestUserService_Register(t *testing.T) {
	userService, cleanup := setupUserService(t)
	defer cleanup()

	req := &models.CreateUserRequest{
		Email:    fmt.Sprintf("newuser_%d@example.com", time.Now().UnixNano()),
		Password: "password123",
		Role:     models.RoleCustomer,
	}

	user, err := userService.Register(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, req.Email, user.Email)
	assert.Equal(t, req.Role, user.Role)
	assert.NotZero(t, user.ID)
}

func TestUserService_Login(t *testing.T) {
	userService, cleanup := setupUserService(t)
	defer cleanup()

	email := fmt.Sprintf("logintest_%d@example.com", time.Now().UnixNano())
	registerReq := &models.CreateUserRequest{
		Email:    email,
		Password: "password123",
		Role:     models.RoleCustomer,
	}

	_, err := userService.Register(context.Background(), registerReq)
	require.NoError(t, err)

	loginReq := &models.LoginRequest{
		Email:    email,
		Password: "password123",
	}

	response, err := userService.Login(context.Background(), loginReq)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, loginReq.Email, response.User.Email)
}

func TestUserService_Login_InvalidCredentials(t *testing.T) {
	userService, cleanup := setupUserService(t)
	defer cleanup()

	loginReq := &models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "wrongpassword",
	}

	response, err := userService.Login(context.Background(), loginReq)
	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestUserService_GetByID(t *testing.T) {
	userService, cleanup := setupUserService(t)
	defer cleanup()

	registerReq := &models.CreateUserRequest{
		Email:    fmt.Sprintf("getbyid_%d@example.com", time.Now().UnixNano()),
		Password: "password123",
		Role:     models.RoleAdmin,
	}

	registeredUser, err := userService.Register(context.Background(), registerReq)
	require.NoError(t, err)

	user, err := userService.GetByID(context.Background(), registeredUser.ID)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, registeredUser.ID, user.ID)
	assert.Equal(t, registeredUser.Email, user.Email)
	assert.Equal(t, registeredUser.Role, user.Role)
}

func TestUserService_DuplicateEmail(t *testing.T) {
	userService, cleanup := setupUserService(t)
	defer cleanup()

	req := &models.CreateUserRequest{
		Email:    fmt.Sprintf("duplicate_%d@example.com", time.Now().UnixNano()),
		Password: "password123",
		Role:     models.RoleCustomer,
	}

	_, err := userService.Register(context.Background(), req)
	require.NoError(t, err)

	_, err = userService.Register(context.Background(), req)
	assert.Error(t, err)
}