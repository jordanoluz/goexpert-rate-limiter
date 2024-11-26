package persistence_strategy

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const BlockedValue = "BLOCKED"

type PersistenceStrategy interface {
	Increment(ctx context.Context, key string, expiration time.Duration) (int, error)
	Block(ctx context.Context, key string, duration time.Duration) error
	IsBlocked(ctx context.Context, key string) (bool, error)
}

type RedisStrategy struct {
	Client *redis.Client
}

func NewRedisStrategy() (PersistenceStrategy, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis client: %w", err)
	}

	return &RedisStrategy{
		Client: rdb,
	}, nil
}

func (rs *RedisStrategy) Increment(ctx context.Context, key string, expiration time.Duration) (int, error) {
	ok, err := rs.Client.SetNX(ctx, key, 0, expiration).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to initialize key '%s': %w", key, err)
	}

	count, err := rs.Client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key '%s': %w", key, err)
	}

	if ok {
		if _, err := rs.Client.Expire(ctx, key, expiration).Result(); err != nil {
			return 0, fmt.Errorf("failed to set expiration for key '%s': %w", key, err)
		}
	}

	return int(count), nil
}

func (rs *RedisStrategy) Block(ctx context.Context, key string, duration time.Duration) error {
	blockKey := fmt.Sprintf("blocked:%s", key)

	if err := rs.Client.Set(ctx, blockKey, BlockedValue, duration).Err(); err != nil {
		return fmt.Errorf("failed to set block for key '%s': %w", key, err)
	}

	return nil
}

func (rs *RedisStrategy) IsBlocked(ctx context.Context, key string) (bool, error) {
	blockKey := fmt.Sprintf("blocked:%s", key)

	result, err := rs.Client.Get(ctx, blockKey).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}

		return false, fmt.Errorf("failed to get block status for key '%s': %w", blockKey, err)
	}

	return result == BlockedValue, nil
}
