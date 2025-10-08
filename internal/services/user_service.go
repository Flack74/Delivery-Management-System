package services

import (
	"context"
	"fmt"

	"delivery-management/internal/db"
	"delivery-management/internal/models"
	"delivery-management/internal/utils"
	"gorm.io/gorm"
)

type UserService struct {
	db         *db.Database
	jwtManager *utils.JWTManager
}

func NewUserService(database *db.Database, jwtManager *utils.JWTManager) *UserService {
	return &UserService{
		db:         database,
		jwtManager: jwtManager,
	}
}

func (s *UserService) Register(ctx context.Context, req *models.CreateUserRequest) (*models.UserResponse, error) {
	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = models.RoleCustomer
	}

	// Create user
	user := &models.User{
		Email:    req.Email,
		Password: hashedPassword,
		Role:     role,
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user.ToResponse(), nil
}

func (s *UserService) Login(ctx context.Context, req *models.LoginRequest) (*models.LoginResponse, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := s.jwtManager.GenerateToken(&user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &models.LoginResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

func (s *UserService) GetByID(ctx context.Context, id uint) (*models.UserResponse, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user.ToResponse(), nil
}
