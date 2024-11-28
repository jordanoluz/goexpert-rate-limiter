package rate_limiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/jordanoluz/goexpert-rate-limiter/pkg/rate_limiter"
	"github.com/stretchr/testify/assert"
)

type mockPersistenceStrategy struct {
	store       map[string]int
	blockedKeys map[string]bool
}

var ctx = context.Background()

func newMockPersistenceStrategy() *mockPersistenceStrategy {
	return &mockPersistenceStrategy{
		store:       make(map[string]int),
		blockedKeys: make(map[string]bool),
	}
}

func (mps *mockPersistenceStrategy) Increment(ctx context.Context, key string, expiration time.Duration) (int, error) {
	if _, blocked := mps.blockedKeys[key]; blocked {
		return 0, nil
	}

	mps.store[key]++

	return mps.store[key], nil
}

func (mps *mockPersistenceStrategy) Block(ctx context.Context, key string, duration time.Duration) error {
	mps.blockedKeys[key] = true

	return nil
}

func (mps *mockPersistenceStrategy) IsBlocked(ctx context.Context, key string) (bool, error) {
	_, blocked := mps.blockedKeys[key]

	return blocked, nil
}

func TestRateLimiterAllow(t *testing.T) {
	mockStrategy := newMockPersistenceStrategy()

	rl := rate_limiter.NewRateLimiter(10, 5, time.Minute, mockStrategy)

	token := "test-token"
	ip := "192.168.1.100"

	for i := 0; i < 10; i++ {
		assert.True(t, rl.Allow(ctx, ip, ""), "request %d should be allowed", i+1)
	}

	assert.False(t, rl.Allow(ctx, ip, ""), "6th request with the ip should be blocked")

	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow(ctx, "", token), "request %d with token should be allowed", i+1)
	}

	assert.False(t, rl.Allow(ctx, "", token), "11th request with token should be blocked")
}

func TestRateLimiterBlocking(t *testing.T) {
	mockStrategy := newMockPersistenceStrategy()

	rl := rate_limiter.NewRateLimiter(3, 3, time.Minute, mockStrategy)

	token := "block-test-token"

	for i := 0; i < 4; i++ {
		rl.Allow(ctx, token, "")
	}

	assert.False(t, rl.Allow(ctx, token, ""), "token should be blocked after exceeding rate limit")

	isBlocked, _ := mockStrategy.IsBlocked(ctx, token)

	assert.True(t, isBlocked, "token should be recorded as blocked in strategy")
}

func TestRateLimiterIPRateLimit(t *testing.T) {
	mockStrategy := newMockPersistenceStrategy()

	rl := rate_limiter.NewRateLimiter(0, 5, time.Minute, mockStrategy)

	ip := "192.168.1.100"

	for i := 0; i < 5; i++ {
		assert.True(t, rl.Allow(ctx, "", ip), "request %d with ip should be allowed", i+1)
	}

	assert.False(t, rl.Allow(ctx, "", ip), "6th request with ip should be blocked")
}

func TestRateLimiterNoTokenOrIP(t *testing.T) {
	mockStrategy := newMockPersistenceStrategy()

	rl := rate_limiter.NewRateLimiter(5, 5, time.Minute, mockStrategy)

	assert.False(t, rl.Allow(ctx, "", ""), "request without token or ip should not be allowed")
}

func TestRateLimiterConcurrentRequests(t *testing.T) {
	mockStrategy := newMockPersistenceStrategy()

	rl := rate_limiter.NewRateLimiter(10, 10, time.Minute, mockStrategy)

	token := "concurrent-token"

	results := make(chan bool, 20)

	for i := 0; i < 20; i++ {
		go func() {
			results <- rl.Allow(ctx, token, "")
		}()
	}

	allowed := 0
	blocked := 0

	for i := 0; i < 20; i++ {
		if <-results {
			allowed++
		} else {
			blocked++
		}
	}

	assert.Equal(t, 10, allowed, "exactly 10 requests should be allowed")
	assert.Equal(t, 10, blocked, "remaining 10 requests should be blocked")
}
