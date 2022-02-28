package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

var (
	ErrPingNotImplemented       = errors.New("ping not implemented for this redis client type")
	ErrSlotsCountNotImplemented = errors.New("slots count not implemented for this client type")
	ErrCastToClusterClient      = errors.New("cast to redis cluster client failed")
)

const (
	Empty uint8 = iota
	SimpleClient
	ClusterClient
	TestClient
)

type Client struct {
	redis.Cmdable
	Type uint8
}

func NewClusterClient(opt *redis.ClusterOptions) *Client {
	return &Client{
		Cmdable: redis.NewClusterClient(opt),
		Type:    ClusterClient,
	}
}

func NewClient(opt *redis.Options) *Client {
	return &Client{
		Cmdable: redis.NewClient(opt),
		Type:    SimpleClient,
	}
}

func MakeTestClient() (*Client, error) {
	mr, err := miniredis.Run()
	if err != nil {
		return nil, fmt.Errorf("miniredis.Run: %w", err)
	}

	client := &Client{
		Type: TestClient,
		Cmdable: redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		}),
	}

	return client, nil
}

func CastToRedisCluster(client redis.Cmdable) (*redis.ClusterClient, error) {
	cluster, ok := client.(*redis.ClusterClient)
	if !ok {
		return nil, ErrCastToClusterClient
	}

	return cluster, nil
}

func ClusterPing(ctx context.Context, client redis.Cmdable) error {
	cluster, err := CastToRedisCluster(client)
	if err != nil {
		return err
	}

	return cluster.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
		return shard.Ping(ctx).Err()
	})
}

func SimplePing(ctx context.Context, client redis.Cmdable) error {
	return client.Ping(ctx).Err()
}

func (c Client) Ping(ctx context.Context) error {
	switch c.Type {
	case ClusterClient:
		return ClusterPing(ctx, c.Cmdable)
	case SimpleClient, TestClient:
		return SimplePing(ctx, c.Cmdable)
	default:
		return fmt.Errorf("clientType=%d: %w", c.Type, ErrPingNotImplemented)
	}
}

func ClusterSlotsCount(ctx context.Context, client redis.Cmdable) (int, error) {
	cluster, err := CastToRedisCluster(client)
	if err != nil {
		return 0, err
	}

	slots, err := cluster.ClusterSlots(ctx).Result()
	if err != nil {
		return 0, err
	}

	return len(slots), nil
}

func (c Client) SlotsCount(ctx context.Context) (int, error) {
	switch c.Type {
	case ClusterClient:
		return ClusterSlotsCount(ctx, c.Cmdable)
	case SimpleClient, TestClient:
		return 0, nil
	default:
		return 0, fmt.Errorf("clientType=%d: %w", c.Type, ErrSlotsCountNotImplemented)
	}
}

func IsNotFoundErr(err error) bool {
	return err != nil && err == redis.Nil
}
