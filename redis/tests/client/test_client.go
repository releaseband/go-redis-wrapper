package client

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/alicebob/miniredis/v2"
)

const entity = "redis_test_client"

type TestClient struct {
	client *redis.Client
}

func MakeTestClient() (*TestClient, error) {
	mr, err := miniredis.Run()
	if err != nil {
		return nil, fmt.Errorf("miniredis.Run: %w", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return &TestClient{client: client}, nil
}

func (t *TestClient) RPush(ctx context.Context, listKey string, val ...interface{}) error {
	return t.client.RPush(ctx, listKey, val).Err()
}

func (t *TestClient) LTrim(ctx context.Context, listKey string, start, stop int64) error {
	return t.client.LTrim(ctx, listKey, start, stop).Err()
}

func (t *TestClient) LRange(ctx context.Context, listKey string, start, stop int64) ([]string, error) {
	return t.client.LRange(ctx, listKey, start, stop).Result()
}

func (t *TestClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return t.client.Set(ctx, key, value, expiration).Err()
}

func (t *TestClient) Get(ctx context.Context, key string) (string, error) {
	return t.client.Get(ctx, key).Result()
}

func (t *TestClient) Status() (interface{}, error) {
	if err := t.client.Ping(context.TODO()).Err(); err != nil {
		return nil, err
	}

	return "OK", nil
}

func (t *TestClient) Entity() string {
	return entity
}

func (t *TestClient) SlotsCount(ctx context.Context) (int, error) {
	slots, err := t.client.ClusterSlots(ctx).Result()
	if err != nil {
		return 0, err
	}

	return len(slots), nil
}
