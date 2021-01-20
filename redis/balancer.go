package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/releaseband/go-redis-wrapper/redis/internal"
)

type BalancerDecorator struct {
	client        RedisClient
	commPostfixes *internal.CommandPostfixes
}

func NewBalancerDecorator(ctx context.Context, client BaseRedisClient) (*BalancerDecorator, error) {
	slots, err := client.impl.ClusterSlots(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster slots: %w", err)
	}

	return &BalancerDecorator{
		client:        client,
		commPostfixes: internal.NewCommandPostfixes(len(slots)),
	}, nil
}

func (b BalancerDecorator) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return b.client.HGetAll(ctx, key)
}

func (b BalancerDecorator) HSet(ctx context.Context, key string, val ...interface{}) error {
	return b.client.HSet(ctx, key, val)
}

func (b BalancerDecorator) HDel(ctx context.Context, key string, fields ...string) error {
	return b.client.HDel(ctx, key, fields...)
}

func (b *BalancerDecorator) RPush(ctx context.Context, listKey string, val ...interface{}) error {
	newKey := listKey + b.commPostfixes.RPushKey()
	return b.client.RPush(ctx, newKey, val)
}

func (b BalancerDecorator) LTrim(ctx context.Context, listKey string, start, stop int64) error {
	newKey := listKey + b.commPostfixes.LTrimKey()
	return b.client.LTrim(ctx, newKey, start, stop)
}

func (b BalancerDecorator) LRange(ctx context.Context, listKey string, start, stop int64) ([]string, error) {
	newKey := listKey + b.commPostfixes.LRangeKey()
	return b.client.LRange(ctx, newKey, start, stop)
}

func (b BalancerDecorator) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return b.client.Set(ctx, key, value, expiration)
}

func (b BalancerDecorator) Get(ctx context.Context, key string) (string, error) {
	return b.client.Get(ctx, key)
}

func (b BalancerDecorator) Status() (interface{}, error) {
	return b.client.Status()
}

func (b BalancerDecorator) Entity() string {
	return b.client.Entity()
}
