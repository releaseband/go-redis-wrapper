package client

import (
	"context"
	"fmt"
	"time"

	redisWrapper "github.com/releaseband/go-redis-wrapper/redis"

	"github.com/go-redis/redis/v8"

	"github.com/alicebob/miniredis/v2"
)

type TestClient struct {
	impl *redis.Client
}

func MakeTestClient() (*TestClient, error) {
	mr, err := miniredis.Run()
	if err != nil {
		return nil, fmt.Errorf("miniredis.Run: %w", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return &TestClient{impl: client}, nil
}

func (t *TestClient) RPush(ctx context.Context, listKey string, val ...interface{}) error {
	return t.impl.RPush(ctx, listKey, val...).Err()
}

func (t *TestClient) LTrim(ctx context.Context, listKey string, start, stop int64) error {
	return t.impl.LTrim(ctx, listKey, start, stop).Err()
}

func (t *TestClient) LRange(ctx context.Context, listKey string, start, stop int64) ([]string, error) {
	return t.impl.LRange(ctx, listKey, start, stop).Result()
}

func (t *TestClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return t.impl.Set(ctx, key, value, expiration).Err()
}

func (t *TestClient) Get(ctx context.Context, key string) (string, error) {
	res, err :=  t.impl.Get(ctx, key).Result()
	if err != nil && redisWrapper.IsNotFoundErr(err) {
		err = redisWrapper.ErrNotFound
	}

	return res, err
}

func (t *TestClient) Ping(ctx context.Context) error {
	return t.impl.Ping(ctx).Err()
}

func (t *TestClient) ReadinessChecker(timeout time.Duration) *redisWrapper.ReadinessChecker {
	return redisWrapper.NewReadinessChecker(timeout, t.Ping)
}

func (t *TestClient) SlotsCount(ctx context.Context) (int, error) {
	slots, err := t.impl.ClusterSlots(ctx).Result()
	if err != nil {
		return 0, err
	}

	return len(slots), nil
}

func (t *TestClient) LLen(ctx context.Context, listKey string) (int64, error) {
	return t.impl.LLen(ctx, listKey).Result()
}

func (t *TestClient) Watch(ctx context.Context, txf func(tx *redis.Tx) error, key ...string) error {
	return t.impl.Watch(ctx, txf, key...)
}

func (t *TestClient) SetEx(ctx context.Context, key string, val interface{}, expiration time.Duration) error {
	return t.impl.SetEX(ctx, key, val, expiration).Err()
}

func (t *TestClient) Impl() *redis.Client {
	return t.impl
}