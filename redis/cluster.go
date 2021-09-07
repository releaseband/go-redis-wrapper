package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Cluster struct {
	impl *redis.ClusterClient
}

func NewRedisCluster(options *redis.ClusterOptions) *Cluster {
	return &Cluster{
		impl: redis.NewClusterClient(options),
	}
}

func (c *Cluster) Get(ctx context.Context, key string) (string, error) {
	result, err := c.impl.Get(ctx, key).Result()
	if isNotFoundErr(err) {
		err = ErrNotFound
	}

	return result, err
}

func (c *Cluster) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.impl.Set(ctx, key, value, expiration).Err()
}

func (c *Cluster) RPush(ctx context.Context, listKey string, val ...interface{}) error {
	return c.impl.RPush(ctx, listKey, val...).Err()
}

func (c *Cluster) LTrim(ctx context.Context, listKey string, start, stop int64) error {
	return c.impl.LTrim(ctx, listKey, start, stop).Err()
}

func (c *Cluster) LRange(ctx context.Context, listKey string, start, stop int64) ([]string, error) {
	return c.impl.LRange(ctx, listKey, start, stop).Result()
}

func (c *Cluster) LLen(ctx context.Context, listKey string) (int64, error) {
	return c.impl.LLen(ctx, listKey).Result()
}

func (c *Cluster) Ping(ctx context.Context) error {
	return c.impl.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
		return shard.Ping(ctx).Err()
	})
}

func (c *Cluster) SlotsCount(ctx context.Context) (int, error) {
	slots, err := c.impl.ClusterSlots(ctx).Result()
	if err != nil {
		return 0, err
	}

	return len(slots), nil
}

func (c *Cluster) ReadinessChecker(timeout time.Duration) *ReadinessChecker {
	return NewReadinessChecker(timeout, c.Ping)
}

func (c *Cluster) Watch(ctx context.Context, txf func(tx *redis.Tx) error, key ...string) error {
	return c.impl.Watch(ctx, txf, key...)
}
