package database

import (
	"fmt"
	"log"
	"time"

	"github.com/tunecent/backend/internal/config"
	"github.com/tunecent/backend/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	*gorm.DB
}

func New(cfg *config.Config) (*DB, error) {
	dsn := cfg.GetDSN()

	// Configure GORM logger
	var gormLogger logger.Interface
	if cfg.Server.Env == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB for connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established")

	return &DB{db}, nil
}

func (db *DB) Migrate() error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.User{},
		&models.MusicMetadata{},
		&models.Campaign{},
		&models.Contribution{},
		&models.RoyaltyPayment{},
		&models.RoyaltyDistribution{},
		&models.UsageDetection{},
		&models.Analytics{},
	)

	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health check
func (db *DB) Ping() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
