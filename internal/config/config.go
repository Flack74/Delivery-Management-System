package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Server   ServerConfig
	Log      LogConfig
	CSRF     CSRFConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type ServerConfig struct {
	Port int
	Mode string
}

type LogConfig struct {
	Level string
}

type CSRFConfig struct {
	Enabled bool
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Database configuration
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "")
	cfg.Database.Name = getEnv("DB_NAME", "delivery_management")
	cfg.Database.SSLMode = getEnv("DB_SSL_MODE", "disable")

	port, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT configuration: %w", err)
	}
	cfg.Database.Port = port

	// Redis configuration
	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")

	redisPort, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT configuration: %w", err)
	}
	cfg.Redis.Port = redisPort

	// JWT configuration
	cfg.JWT.Secret = getEnv("JWT_SECRET", "")
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required and cannot be empty")
	}
	if len(cfg.JWT.Secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters long for security")
	}

	expiry, err := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY configuration: %w", err)
	}
	cfg.JWT.Expiry = expiry

	// Server configuration
	serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT configuration: %w", err)
	}
	cfg.Server.Port = serverPort
	cfg.Server.Mode = getEnv("GIN_MODE", "debug")

	// Log configuration
	cfg.Log.Level = getEnv("LOG_LEVEL", "info")

	// CSRF configuration
	cfg.CSRF.Enabled = getEnv("CSRF_PROTECTION", "true") == "true"

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=*** dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User,
		c.Database.Name, c.Database.SSLMode)
}

func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password,
		c.Database.Name, c.Database.SSLMode)
}

// GetSafeDatabaseDSN returns a sanitized DSN for logging purposes
func (c *Config) GetSafeDatabaseDSN() string {
	return c.DatabaseDSN()
}


