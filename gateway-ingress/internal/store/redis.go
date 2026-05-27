package store

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr string) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	return &RedisStore{client: client}, nil
}

// CheckLimits returns true if the request is allowed, false if rate limit or quota is exceeded
func (r *RedisStore) CheckLimits(ctx context.Context, bapID string) (bool, error) {
	if bapID == "" {
		bapID = "anonymous"
	}

	now := time.Now()
	
	// Rate limit: 100 req/sec per BAP
	rateKey := fmt.Sprintf("rate:%s:%d", bapID, now.Unix())
	// Quota: 100k req/month per BAP
	quotaKey := fmt.Sprintf("quota:%s:%s", bapID, now.Format("2006-01"))

	pipe := r.client.Pipeline()
	rateCmd := pipe.Incr(ctx, rateKey)
	pipe.Expire(ctx, rateKey, 5*time.Second) // short expiry for rate key
	
	quotaCmd := pipe.Incr(ctx, quotaKey)
	pipe.Expire(ctx, quotaKey, 32*24*time.Hour) // ~1 month expiry

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("redis pipeline error: %w", err)
	}

	// Hardcoded limits for V1 Enterprise Prototype
	if rateCmd.Val() > 100 {
		return false, nil // rate limit exceeded
	}

	if quotaCmd.Val() > 100000 {
		return false, nil // quota exceeded
	}

	return true, nil
}
