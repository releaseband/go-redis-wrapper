package go_redis_wrapper

import (
	"context"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"go.opentelemetry.io/otel/attribute"
	"time"
)

const (
	empty uint8 = iota
	simpleClientType
	clusterClientType
	testClientType
)

var (
	attributeLock   = attribute.String(commandKey, "lock")
	attributeUnlock = attribute.String(commandKey, "unlock")
)

type Client struct {
	redis.UniversalClient
	rs   *redsync.Redsync
	Type uint8
}

func newRedSync(client redis.UniversalClient) *redsync.Redsync {
	return redsync.New(goredis.NewPool(client))
}

func newClient(uc redis.UniversalClient, _type uint8) *Client {
	return &Client{
		UniversalClient: uc,
		rs:              newRedSync(uc),
		Type:            _type,
	}
}

func addHooks(cli redis.UniversalClient) {
	cli.AddHook(newRedisHookMetrics())
}

func NewClusterClient(opt *redis.ClusterOptions) *Client {
	cli := newClient(redis.NewClusterClient(opt), clusterClientType)
	addHooks(cli.UniversalClient)

	return cli
}

func NewClient(opt *redis.Options) *Client {
	cli := newClient(redis.NewClient(opt), simpleClientType)
	addHooks(cli.UniversalClient)

	return cli
}

func StartMiniRedis() (*Client, error) {
	mr, err := miniredis.Run()
	if err != nil {
		return nil, fmt.Errorf("miniredis.Run: %w", err)
	}

	uc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cli := newClient(uc, testClientType)
	addHooks(cli.UniversalClient)

	return cli, nil
}

func ClientAdapter(uc redis.UniversalClient, _type uint8) (*Client, error) {
	switch _type {
	case simpleClientType, clusterClientType, testClientType:
	//
	default:
		return nil, ErrInvalidClientType

	}

	return newClient(uc, _type), nil
}

func CastToRedisCluster(client redis.UniversalClient) (*redis.ClusterClient, error) {
	cluster, ok := client.(*redis.ClusterClient)
	if !ok {
		return nil, ErrCastToClusterClient
	}

	return cluster, nil
}

func ClusterPing(ctx context.Context, client redis.UniversalClient) error {
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
	case clusterClientType:
		return ClusterPing(ctx, c.UniversalClient)
	case simpleClientType, testClientType:
		return SimplePing(ctx, c.UniversalClient)
	default:
		return fmt.Errorf("clientType=%d: %w", c.Type, ErrPingNotImplemented)
	}
}

func ClusterSlotsCount(ctx context.Context, client redis.UniversalClient) (int, error) {
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
	case clusterClientType:
		return ClusterSlotsCount(ctx, c.UniversalClient)
	case simpleClientType, testClientType:
		return 0, nil
	default:
		return 0, fmt.Errorf("clientType=%d: %w", c.Type, ErrSlotsCountNotImplemented)
	}
}

func (c *Client) Lock(ctx context.Context, key string, options ...redsync.Option) (*redsync.Mutex, error) {
	mutex := c.rs.NewMutex(key, options...)

	if err := mutex.LockContext(ctx); err != nil {
		return nil, err
	}

	return mutex, nil
}

func (c *Client) LockKey(ctx context.Context, key string, options ...redsync.Option) (func(context.Context) error, error) {
	start := time.Now()

	mutex, err := c.Lock(ctx, key, options...)
	measure.Record(ctx, time.Since(start).Seconds(), attributeLock)
	if err != nil {
		return nil, err
	}

	unlock := func(ctx context.Context) error {
		start := time.Now()

		ok, err := mutex.UnlockContext(ctx)
		measure.Record(ctx, time.Since(start).Seconds(), attributeUnlock)

		if err != nil {
			return err
		}

		if !ok {
			return ErrUnlockStatusIsFailure
		}

		return nil
	}

	return unlock, nil
}
