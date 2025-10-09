package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"delivery-management/internal/config"
	"delivery-management/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	*gorm.DB
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
		PrepareStmt: true,
	}

	dsn := cfg.GetDatabaseDSN()
	log.Printf("Connecting to database: %s", cfg.DatabaseDSN()) // Use masked version for logging
	
	var db *gorm.DB
	var err error
	
	// Retry connection up to 5 times
	for i := 0; i < 5; i++ {
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err == nil {
			break
		}
		log.Printf("Database connection attempt %d failed: %v", i+1, err)
		if i < 4 {
			time.Sleep(time.Duration(i+1) * 2 * time.Second)
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after 5 attempts: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool for better performance
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetMaxOpenConns(200)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &Database{DB: db}

	// Run migrations with timeout
	migrationCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := database.migrateWithContext(migrationCtx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Database connected and migrated successfully")
	return database, nil
}

func (d *Database) migrate() error {
	return d.AutoMigrate(
		&models.User{},
		&models.Order{},
	)
}

func (d *Database) migrateWithContext(ctx context.Context) error {
	return d.WithContext(ctx).AutoMigrate(
		&models.User{},
		&models.Order{},
	)
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *Database) HealthCheck(ctx context.Context) error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
