package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClient interface {
	Status() (interface{}, error)
	HGetAll(ctx context.Context, groupKey string) (map[string]string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	HSet(ctx context.Context, key, field string, value interface{}) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	HDel(ctx context.Context, key, field string) error
	Ping() error
}

type redisClient interface {
	RedisClient
	Entity() string
}

type implClient interface {
	Ping(ctx context.Context) *redis.StatusCmd
	HGetAll(ctx context.Context, groupKey string) *redis.StringStringMapCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	HSet(ctx context.Context, key string, value ...interface{}) *redis.IntCmd
	Set(ctx context.Context, key string, value interface{}, exp time.Duration) *redis.StatusCmd
	HDel(ctx context.Context, key string, field ...string) *redis.IntCmd
	Del(ctx context.Context, key ...string) *redis.IntCmd
}
