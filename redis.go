package go_redis_wrapper

import (
	"context"
	"errors"
	"fmt"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

const (
	empty uint8 = iota
	simpleClientType
	clusterClientType
	testClientType
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

func NewClusterClient(opt *redis.ClusterOptions) *Client {
	return newClient(redis.NewClusterClient(opt), clusterClientType)
}

func NewClient(opt *redis.Options) *Client {
	return newClient(redis.NewClient(opt), simpleClientType)
}

func StartMiniRedis() (*Client, error) {
	mr, err := miniredis.Run()
	if err != nil {
		return nil, fmt.Errorf("miniredis.Run: %w", err)
	}

	uc := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return newClient(uc, testClientType), nil
}

func ClientAdapter(uc redis.UniversalClient, _type uint8) (*Client, error) {
	switch _type {
	case simpleClientType, clusterClientType, testClientType:
	// ok
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

func (c *Client) Ping(ctx context.Context) error {
	switch c.Type {
	case clusterClientType:
		return ClusterPing(ctx, c.UniversalClient)
	case simpleClientType, testClientType:
		return SimplePing(ctx, c.UniversalClient)
	default:
		return fmt.Errorf("clientType=%d: %w", c.Type, ErrPingNotImplemented)
	}
}

func (c *Client) Status() (interface{}, error) {
	if err := c.Ping(context.Background()); err != nil {
		return nil, err
	}

	return "ok", nil
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

func (c *Client) Lock(
	ctx context.Context,
	key string,
	options ...redsync.Option,
) (*redsync.Mutex, error) {
	mutex := c.rs.NewMutex(key, options...)

	if err := mutex.LockContext(ctx); err != nil {
		return nil, fmt.Errorf("lock context: %w", err)
	}

	return mutex, nil
}

// LockKey acquires a distributed lock on key and returns an unlock function.
// Retries up to 32 times (redsync default) with delays between attempts — blocks until the lock
// is acquired or all retries are exhausted. Use redsync.WithTries to override the retry count.
func (c *Client) LockKey(
	ctx context.Context,
	key string,
	options ...redsync.Option,
) (func(context.Context) error, error) {
	mutex, err := c.Lock(ctx, key, options...)
	if err != nil {
		return nil, err
	}

	callback := func(ctx context.Context) error {
		return unlock(ctx, mutex)
	}

	return callback, nil
}

// TryLockKey makes a single attempt to acquire a distributed lock on key without retries.
// Returns ErrResourceBusy immediately if the key is already locked — use this when waiting
// for a lock is not acceptable. Unlike LockKey, it never blocks on contention.
func (c *Client) TryLockKey(
	ctx context.Context,
	key string,
	options ...redsync.Option,
) (func(context.Context) error, error) {
	mutex, err := c.TryLock(ctx, key, options...)
	if err != nil {
		return nil, err
	}

	callback := func(ctx context.Context) error {
		return unlock(ctx, mutex)
	}

	return callback, nil
}

func (c *Client) TryLock(
	ctx context.Context,
	key string,
	options ...redsync.Option,
) (*redsync.Mutex, error) {
	mutex := c.rs.NewMutex(key, options...)

	if err := mutex.TryLockContext(ctx); err != nil {
		var taken *redsync.ErrTaken

		if errors.As(err, &taken) {
			return nil, ErrResourceBusy
		}

		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	return mutex, nil
}

func unlock(ctx context.Context, mutex *redsync.Mutex) error {
	ok, err := mutex.UnlockContext(ctx)
	if err != nil {
		return err
	}

	if !ok {
		return ErrUnlockStatusIsFailure
	}

	return nil
}
