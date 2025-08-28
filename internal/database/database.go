package database

import (
	"fmt"
	"os"
	"time"
	"elk-stack-user/internal/model"
	"elk-stack-user/internal/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewConfig() *Config {
	return &Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "user_service"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

func Connect(config *Config) (*gorm.DB, error) {
	dsn := config.GetDSN()
	
	// GORM logger konfigürasyonu
	gormLogLevel := gormlogger.Info
	if os.Getenv("ENV") == "development" {
		gormLogLevel = gormlogger.Info
	}
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormLogLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Connection pool ayarları
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Connection pool ayarları
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Logger.Info("Database connected successfully",
		logger.String("host", config.Host),
		logger.String("port", config.Port),
		logger.String("database", config.DBName),
		logger.String("user", config.User),
	)
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	logger.Logger.Info("Starting database migration...")
	
	err := db.AutoMigrate(&model.User{}, &model.LoginAttempt{}, &model.BanRecord{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	logger.Logger.Info("Database migration completed successfully")
	return nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
