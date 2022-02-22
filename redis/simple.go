package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Simple struct {
	impl *redis.Client
}

func NewRedisSimple(opt *redis.Options) *Simple {
	return &Simple{impl: redis.NewClient(opt)}
}

func (s *Simple) RPush(ctx context.Context, listKey string, val ...interface{}) error {
	return s.impl.RPush(ctx, listKey, val...).Err()
}

func (s *Simple) LTrim(ctx context.Context, listKey string, start, stop int64) error {
	return s.impl.LTrim(ctx, listKey, start, stop).Err()
}

func (s *Simple) LRange(ctx context.Context, listKey string, start, stop int64) ([]string, error) {
	return s.impl.LRange(ctx, listKey, start, stop).Result()
}

func (s *Simple) LLen(ctx context.Context, listKey string) (int64, error) {
	return s.impl.LLen(ctx, listKey).Result()
}

func (s *Simple) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return s.impl.Set(ctx, key, value, expiration).Err()
}

func (s *Simple) Get(ctx context.Context, key string) (string, error) {
	result, err := s.impl.Get(ctx, key).Result()
	if IsNotFoundErr(err) {
		err = ErrNotFound
	}

	return result, err
}

func (s *Simple) Ping(ctx context.Context) error {
	return s.impl.Ping(ctx).Err()
}

func (s *Simple) SlotsCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (s *Simple) Watch(ctx context.Context, txf func(tx *redis.Tx) error, key ...string) error {
	return s.impl.Watch(ctx, txf, key...)
}

func (s *Simple) ReadinessChecker(timeout time.Duration) *ReadinessChecker {
	return NewReadinessChecker(timeout, s.Ping)
}

func (s *Simple) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return s.impl.SetEX(ctx, key, value, expiration).Err()
}

func (s *Simple) Del(ctx context.Context, key string) error {
	return s.impl.Del(ctx, key).Err()
}

func (s *Simple) Impl() redis.Cmdable {
	return s.impl
}

func (s *Simple) Uc() redis.UniversalClient {
	return s.impl
}

func (s *Simple) Incr(ctx context.Context, key string) (int64, error) {
	return s.impl.Incr(ctx, key).Result()
}

func (s *Simple) HSet(ctx context.Context, key string, val ...interface{}) error {
	return s.impl.HSet(ctx, key, val...).Err()
}

func (s *Simple) HGet(ctx context.Context, key, field string) (string, error) {
	return s.impl.HGet(ctx, key, field).Result()
}

func (s *Simple) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return s.impl.HGetAll(ctx, key).Result()
}

func (s *Simple) HDel(ctx context.Context, key string) error {
	return s.impl.HDel(ctx, key).Err()
}
