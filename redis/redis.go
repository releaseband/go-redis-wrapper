package redis

import (
	"context"
	"time"
)

type RedisClient interface {
	RPush(ctx context.Context, listKey string, val ...interface{}) error
	LTrim(ctx context.Context, listKey string, start, stop int64) error
	LRange(ctx context.Context, listKey string, start, stop int64) ([]string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Status() (interface{}, error)
	Entity() string
}
