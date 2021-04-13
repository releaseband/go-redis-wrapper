package redis

import (
	"context"
	"fmt"
	"time"

	"errors"

	"github.com/go-redis/redis/v8"
)

const (
	entityCluster     = "cluster"
	entitySimpleRedis = "simple"
)

var (
	ErrNotFound = errors.New("not found")
)

type BaseRedisClient struct {
	impl   redis.Cmdable
	entity string
	ping   func(ctx context.Context) error
}

func NewRedisCluster(options *redis.ClusterOptions) *BaseRedisClient {
	cli := redis.NewClusterClient(options)

	return &BaseRedisClient{
		impl:   cli,
		entity: entityCluster,
		ping:   clusterPing(cli),
	}
}

func NewRedisSimple(options *redis.Options) *BaseRedisClient {
	cli := redis.NewClient(options)

	return &BaseRedisClient{
		impl:   cli,
		entity: entitySimpleRedis,
		ping:   simplePing(cli),
	}
}

func (c *BaseRedisClient) Get(ctx context.Context, key string) (string, error) {
	result, err := c.impl.Get(ctx, key).Result()
	if err != nil && err == redis.Nil {
		err = ErrNotFound
	}

	return result, err
}

func (c *BaseRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.impl.Set(ctx, key, value, expiration).Err()
}

func simplePing(cli *redis.Client) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		err := cli.Ping(ctx).Err()
		if err != nil {
			err = fmt.Errorf("redis ping failed: %w", err)
		}

		return err
	}
}

func clusterPing(cli *redis.ClusterClient) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		err := cli.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
			return shard.Ping(ctx).Err()
		})

		if err != nil {
			err = fmt.Errorf("clusterPing failed: %w", err)
		}

		return err
	}
}

func (c *BaseRedisClient) Ping(ctx context.Context) error {
	return c.ping(ctx)
}

func (c *BaseRedisClient) RPush(ctx context.Context, listKey string, val ...interface{}) error {
	return c.impl.RPush(ctx, listKey, val...).Err()
}

func (c *BaseRedisClient) LTrim(ctx context.Context, listKey string, start, stop int64) error {
	return c.impl.LTrim(ctx, listKey, start, stop).Err()
}

func (c *BaseRedisClient) LRange(ctx context.Context, listKey string, start, stop int64) ([]string, error) {
	return c.impl.LRange(ctx, listKey, start, stop).Result()
}

func (c *BaseRedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.impl.HGetAll(ctx, key).Result()
}

func (c *BaseRedisClient) HSet(ctx context.Context, key string, val ...interface{}) error {
	return c.impl.HSet(ctx, key, val...).Err()
}

func (c *BaseRedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	return c.impl.HDel(ctx, key, fields...).Err()
}

func (c *BaseRedisClient) SlotsCount(ctx context.Context) (int, error) {
	if c.entity == entityCluster {
		slots, err := c.impl.ClusterSlots(ctx).Result()
		if err != nil {
			return 0, err
		}

		return len(slots), nil
	}

	return 0, nil
}

func (c *BaseRedisClient) LLen(ctx context.Context, listKey string) (int64, error) {
	return c.impl.LLen(ctx, listKey).Result()
}

func (c *BaseRedisClient) ReadinessCheck() func(ctx context.Context) (interface{}, error) {
	return makeReadinessCheckerFunc(c.ping)
}
