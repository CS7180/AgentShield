package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

func NewClient(redisURL string) (*Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}

	rdb := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

func (c *Client) Close() error {
	return c.rdb.Close()
}

// TokenBucketAllow implements a Redis Lua-based token bucket rate limiter.
// key: unique identifier (e.g., "ratelimit:user:<id>:scans")
// maxTokens: bucket capacity
// refillRate: tokens added per second
// Returns true if the request is allowed.
var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]
local max_tokens = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local requested = tonumber(ARGV[4])

local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens = tonumber(bucket[1])
local last_refill = tonumber(bucket[2])

if tokens == nil then
  tokens = max_tokens
  last_refill = now
end

local elapsed = now - last_refill
local new_tokens = math.min(max_tokens, tokens + elapsed * refill_rate)

if new_tokens >= requested then
  redis.call('HMSET', key, 'tokens', new_tokens - requested, 'last_refill', now)
  redis.call('EXPIRE', key, math.ceil(max_tokens / refill_rate) + 10)
  return 1
else
  redis.call('HMSET', key, 'tokens', new_tokens, 'last_refill', now)
  redis.call('EXPIRE', key, math.ceil(max_tokens / refill_rate) + 10)
  return 0
end
`)

// AllowRequest checks the token bucket for the given key.
// maxTokens is the bucket size, refillPerSec is the refill rate.
func (c *Client) AllowRequest(ctx context.Context, key string, maxTokens int, refillPerSec float64) (bool, error) {
	now := float64(time.Now().UnixNano()) / 1e9
	result, err := tokenBucketScript.Run(ctx, c.rdb,
		[]string{key},
		maxTokens,
		refillPerSec,
		now,
		1,
	).Int()
	if err != nil {
		return false, fmt.Errorf("rate limit script: %w", err)
	}
	return result == 1, nil
}
