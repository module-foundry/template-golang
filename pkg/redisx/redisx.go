// Package redisx wraps the Redis client used by the application.
package redisx

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config describes the Redis connection.
type Config struct {
	Addr     string
	Password string
	DB       int
}

// New creates a Redis client and verifies connectivity.
func New(ctx context.Context, cfg Config) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}

// Healthcheck verifies connectivity with a short timeout.
func Healthcheck(ctx context.Context, client *redis.Client) error {
	if client == nil {
		return fmt.Errorf("redis client is nil")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return client.Ping(ctx).Err()
}
