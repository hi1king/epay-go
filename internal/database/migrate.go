// internal/database/migrate.go
package database

import (
	"log"

	"github.com/example/epay-go/internal/model"
)

// Migrate 自动迁移数据库表
func Migrate() error {
	log.Println("Running database migrations...")

	err := DB.AutoMigrate(
		&model.Admin{},
		&model.Merchant{},
		&model.Channel{},
		&model.Order{},
		&model.Settlement{},
		&model.BalanceRecord{},
		&model.Config{},
		&model.Refund{},
	)

	if err != nil {
		return err
	}

	log.Println("Database migrations completed")
	return nil
}
