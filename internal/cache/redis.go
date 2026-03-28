// internal/cache/redis.go
package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/example/epay-go/internal/config"
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func Init() error {
	cfg := config.Get().Redis

	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RDB.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect redis: %w", err)
	}

	log.Println("Redis connected successfully")
	return nil
}

func Get() *redis.Client {
	return RDB
}

func Close() error {
	return RDB.Close()
}
