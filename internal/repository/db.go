package repository

import (
	"fmt"
	"log"
	"time"

	"github.com/example/epay-go/internal/config"
	"github.com/example/epay-go/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func InitDB() error {
	cfg := config.Get().Database

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	var gormConfig *gorm.Config
	if config.Get().Server.Mode == "debug" {
		gormConfig = &gorm.Config{Logger: logger.Default.LogMode(logger.Info)}
	} else {
		gormConfig = &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connected successfully")
	return nil
}

func GetDB() *gorm.DB {
	return db
}

func CloseDB() error {
	if db == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func MigrateDB() error {
	log.Println("Running database migrations...")

	if err := db.AutoMigrate(
		&model.Admin{},
		&model.Merchant{},
		&model.Channel{},
		&model.Order{},
		&model.Settlement{},
		&model.BalanceRecord{},
		&model.Config{},
		&model.Refund{},
	); err != nil {
		return err
	}

	log.Println("Database migrations completed")
	return nil
}
