package redis

import (
	"context"
	"errors"
	"github.com/go-redsync/redsync/v4"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	ErrNotFound = errors.New("not found")
)

func IsNotFoundErr(err error) bool {
	return err != nil && err == redis.Nil
}

type RedisClient interface {
	RPush(ctx context.Context, listKey string, val ...interface{}) error
	LTrim(ctx context.Context, listKey string, start, stop int64) error
	LRange(ctx context.Context, listKey string, start, stop int64) ([]string, error)
	LLen(ctx context.Context, listKey string) (int64, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Ping(ctx context.Context) error
	SlotsCount(ctx context.Context) (int, error)
	Watch(ctx context.Context, txf func(tx *redis.Tx) error, key ...string) error
	ReadinessChecker(timeout time.Duration) *ReadinessChecker
	Del(ctx context.Context, key string) error
	Incr(ctx context.Context, key string) (int64, error)
	HSet(ctx context.Context, key string, val ...interface{}) error
	HGet(ctx context.Context, key, field string) (string, error)
	HGetAll(ctx context.Context, key string) (map[string]string, error)
	HDel(ctx context.Context, key string, field ...string) error
	Impl() redis.Cmdable
	Uc() redis.UniversalClient
	Lock(ctx context.Context, key string, opt ...redsync.Option) (*redsync.Mutex, error)
}
