package rate_limiter

import (
	"context"
	"time"

	persistenceStrategy "github.com/jordanoluz/goexpert-rate-limiter/internal/infra/persistence_strategy"
)

type RateLimiter interface {
	Allow(ctx context.Context, token, ip string) bool
}

type rateLimiter struct {
	RateLimitToken      int
	RateLimitIP         int
	BlockDuration       time.Duration
	PersistenceStrategy persistenceStrategy.PersistenceStrategy
}

func NewRateLimiter(rateLimitToken, rateLimitIP int, blockDuration time.Duration, persistenceStrategy persistenceStrategy.PersistenceStrategy) RateLimiter {
	return &rateLimiter{
		RateLimitToken:      rateLimitToken,
		RateLimitIP:         rateLimitIP,
		BlockDuration:       blockDuration,
		PersistenceStrategy: persistenceStrategy,
	}
}

func (rl *rateLimiter) Allow(ctx context.Context, token, ip string) bool {
	if token != "" {
		return rl.checkRateLimit(ctx, token, rl.RateLimitToken)
	}

	if ip != "" {
		return rl.checkRateLimit(ctx, ip, rl.RateLimitIP)
	}

	return false
}

func (rl *rateLimiter) checkRateLimit(ctx context.Context, key string, limit int) bool {
	isBlocked, err := rl.PersistenceStrategy.IsBlocked(ctx, key)
	if err != nil {
		return false
	}

	if isBlocked {
		return false
	}

	count, err := rl.PersistenceStrategy.Increment(ctx, key, rl.BlockDuration)
	if err != nil {
		return false
	}

	if count > limit {
		rl.PersistenceStrategy.Block(ctx, key, rl.BlockDuration)

		return false
	}

	return true
}
