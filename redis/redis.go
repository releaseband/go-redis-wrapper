package redis

import (
	"context"
	"time"
)

type RedisClient interface {
	RPush(ctx context.Context, listKey string, val ...interface{}) error
	LTrim(ctx context.Context, listKey string, start, stop int64) error
	LRange(ctx context.Context, listKey string, start, stop int64) ([]string, error)
	LLen(ctx context.Context, listKey string) (int64, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Ping(ctx context.Context) error
	SlotsCount(ctx context.Context) (int, error)
	ReadinessChecker(timeout time.Duration) *ReadinessChecker
}
