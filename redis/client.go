package redis

import (
	"context"
	"fmt"
	"time"

	"errors"

	"github.com/go-redis/redis/v8"
)

const (
	entityCluster     = "redis_cluster"
	entitySimpleRedis = "redis"
)

var (
	ErrNotFound = errors.New("not found")
	pingCtx     = context.TODO()
)

type BaseRedisClient struct {
	impl   implClient
	entity string
	ping   func() error
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

func (c BaseRedisClient) HGetAll(ctx context.Context, groupKey string) (map[string]string, error) {
	return c.impl.HGetAll(ctx, groupKey).Result()
}

func (c BaseRedisClient) Get(ctx context.Context, key string) (string, error) {
	result, err := c.impl.Get(ctx, key).Result()
	if err != nil && err == redis.Nil {
		err = ErrNotFound
	}

	return result, err
}

func (c BaseRedisClient) HSet(ctx context.Context, key, field string, value interface{}) error {
	return c.impl.HSet(ctx, key, field, value).Err()
}

func (c BaseRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.impl.Set(ctx, key, value, expiration).Err()
}

func (c BaseRedisClient) HDel(ctx context.Context, key, field string) error {
	return c.impl.HDel(ctx, key, field).Err()
}

func (c BaseRedisClient) Del(ctx context.Context, key string) error {
	return c.impl.Del(ctx, key).Err()
}

func simplePing(cli *redis.Client) func() error {
	return func() error {
		err := cli.Ping(pingCtx).Err()
		if err != nil {
			err = fmt.Errorf("redis ping failed: %w", err)
		}

		return err
	}
}

func clusterPing(cli *redis.ClusterClient) func() error {
	return func() error {
		err := cli.ForEachShard(pingCtx, func(ctx context.Context, shard *redis.Client) error {
			return shard.Ping(ctx).Err()
		})

		if err != nil {
			err = fmt.Errorf("clusterPing failed: %w", err)
		}

		return err
	}
}

func (c BaseRedisClient) Ping() error {
	return c.ping()
}

func (c BaseRedisClient) Status() (interface{}, error) {
	if err := c.Ping(); err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}

	return "ok", nil
}

func (c BaseRedisClient) Entity() string {
	return c.entity
}
