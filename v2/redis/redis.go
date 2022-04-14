package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/releaseband/go-redis-wrapper/v2/internal"
	"time"
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
	rs   *redsync.Redsync
	Type uint8
}

func newRedSync(client redis.UniversalClient) *redsync.Redsync {
	return redsync.New(goredis.NewPool(client))
}

func NewClusterClient(opt *redis.ClusterOptions) *Client {
	client := redis.NewClusterClient(opt)

	return &Client{
		Cmdable: client,
		rs:      newRedSync(client),
		Type:    ClusterClient,
	}
}

func NewClient(opt *redis.Options) *Client {
	client := redis.NewClient(opt)

	return &Client{
		Cmdable: client,
		rs:      newRedSync(client),
		Type:    SimpleClient,
	}
}

func MakeTestClient() (*Client, error) {
	mr, err := miniredis.Run()
	if err != nil {
		return nil, fmt.Errorf("miniredis.Run: %w", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return &Client{
		Type:    TestClient,
		rs:      newRedSync(client),
		Cmdable: client,
	}, nil
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

func (c *Client) Lock(ctx context.Context, key string, options ...redsync.Option) (*redsync.Mutex, error) {
	return internal.Lock(ctx, c.rs, key, options...)
}

type Options struct {
	Expire time.Duration
	Tries  int
}

func (o Options) makeOptionsList() []redsync.Option {
	options := make([]redsync.Option, 0, 2)

	if o.Expire != 0 {
		options = append(options, redsync.WithExpiry(o.Expire))
	}

	if o.Tries > 0 {
		options = append(options, redsync.WithTries(o.Tries))
	}

	return options
}

func (c *Client) LockWithOptions(ctx context.Context, key string, opt Options) (
	func(ctx context.Context) (bool, error), error) {
	m, err := c.Lock(ctx, key, opt.makeOptionsList()...)
	if err != nil {
		return nil, err
	}

	return m.UnlockContext, nil
}
