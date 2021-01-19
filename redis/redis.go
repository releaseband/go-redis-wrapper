package redis

import (
	"context"
	"time"
)

// deprecated
type deprecated interface {
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HSet(ctx context.Context, key string, val ...interface{}) error
	HDel(ctx context.Context, key string, fields ...string) error
}

type RedisClient interface {
	deprecated
	RPush(ctx context.Context, listKey string, val ...interface{}) error
	LTrim(ctx context.Context, listKey string, start, stop int64) error
	LRange(ctx context.Context, listKey string, start, stop int64) ([]string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Status() (interface{}, error)
	Entity() string
}
