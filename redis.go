package go_redis_wrapper

import (
	"context"
	"fmt"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

const (
	empty uint8 = iota
	simpleClientType
	clusterClientType
	testClientType
)

// Client is a wrapper around redis.UniversalClient with additional features.
// It includes a redsync instance for distributed locking and a Type field
// to identify the kind of Redis client (simple, cluster, or test).
type Client struct {
	redis.UniversalClient
	rs   *redsync.Redsync
	Type uint8
}

func newClient(uc redis.UniversalClient, _type uint8) *Client {
	return &Client{
		UniversalClient: uc,
		rs:              newRedSync(uc),
		Type:            _type,
	}
}

// newRedSync creates a new redsync instance using the provided
// redis.UniversalClient.
func newRedSync(client redis.UniversalClient) *redsync.Redsync {
	return redsync.New(goredis.NewPool(client))
}

// NewClusterClient creates a new Client instance configured as a Redis
// cluster client using the provided ClusterOptions.
// It initializes the underlying redis.ClusterClient and sets the client type.
// Returns the newly created Client.
func NewClusterClient(opt *redis.ClusterOptions) *Client {
	return newClient(redis.NewClusterClient(opt), clusterClientType)
}

// NewClient creates a new Client instance configured as a simple Redis
// client using the provided Options.
// It initializes the underlying redis.Client and sets the client type.
// Returns the newly created Client.
func NewClient(opt *redis.Options) *Client {
	return newClient(redis.NewClient(opt), simpleClientType)
}

// StartMiniRedis starts a new instance of miniredis for testing purposes.
// It creates a simple Redis client connected to the miniredis instance
// and returns a Client wrapper around it.
// Returns the Client and any error encountered during setup.
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
	//
	default:
		return nil, ErrInvalidClientType

	}

	return newClient(uc, _type), nil
}

// CastToRedisCluster attempts to cast the provided redis.UniversalClient
// to a redis.ClusterClient. If the cast is successful, it returns the
// ClusterClient; otherwise, it returns an error indicating the failure.
// Returns the casted *redis.ClusterClient and an error if the cast fails.
func CastToRedisCluster(
	client redis.UniversalClient,
) (*redis.ClusterClient, error) {
	cluster, ok := client.(*redis.ClusterClient)
	if !ok {
		return nil, ErrCastToClusterClient
	}

	return cluster, nil
}

// ClusterPing sends a PING command to all shards in the Redis cluster
// using the provided redis.UniversalClient.
// It returns an error if any shard fails to respond to the PING command.
func ClusterPing(ctx context.Context, client redis.UniversalClient) error {
	cluster, err := CastToRedisCluster(client)
	if err != nil {
		return err
	}

	return cluster.ForEachShard(ctx,
		func(ctx context.Context, shard *redis.Client) error {
			return shard.Ping(ctx).Err()
		})
}

// SimplePing sends a PING command to the Redis server
// using the provided redis.Cmdable client.
// It returns an error if the PING command fails.
func SimplePing(ctx context.Context, client redis.Cmdable) error {
	return client.Ping(ctx).Err()
}

// Ping sends a PING command to the Redis server or cluster
// based on the client type.
// It returns an error if the PING command fails or
// if the client type is unsupported.
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

// Status checks the health of the Redis client by sending a PING command.
// If the PING is successful, it returns "ok"; otherwise, it returns an error.
// used for health checks
func (c *Client) Status() (interface{}, error) {
	if err := c.Ping(context.Background()); err != nil {
		return nil, err
	}

	return "ok", nil
}

// ClusterSlotsCount retrieves the number of slots in the Redis cluster
// using the provided redis.UniversalClient.
// It returns the count of slots and any error encountered during the operation.
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

// SlotsCount retrieves the number of slots in the Redis cluster
// based on the client type.
// For cluster clients, it returns the actual slot count;
// for simple and test clients, it returns 0.
// It returns an error if the client type is unsupported.
func (c Client) SlotsCount(ctx context.Context) (int, error) {
	switch c.Type {
	case clusterClientType:
		return ClusterSlotsCount(ctx, c.UniversalClient)
	case simpleClientType, testClientType:
		return 0, nil
	default:
		return 0, fmt.Errorf("clientType=%d: %w",
			c.Type, ErrSlotsCountNotImplemented)
	}
}

// Lock creates and acquires a distributed lock for the given key
// using redsync. It accepts context and optional redsync options.
// Returns the acquired mutex and any error encountered during locking.
func (c *Client) Lock(
	ctx context.Context, key string,
	options ...redsync.Option,
) (*redsync.Mutex, error) {
	mutex := c.rs.NewMutex(key, options...)

	if err := mutex.LockContext(ctx); err != nil {
		return nil, err
	}

	return mutex, nil
}

// LockKey creates and acquires a distributed lock for the given key
// using redsync. It accepts context and optional redsync options.
// It returns an unlock function to release the lock and any error encountered
// during locking.
func (c *Client) LockKey(
	ctx context.Context, key string,
	options ...redsync.Option,
) (func(context.Context) error, error) {
	mutex, err := c.Lock(ctx, key, options...)
	if err != nil {
		return nil, err
	}

	unlock := func(ctx context.Context) error {
		ok, err := mutex.UnlockContext(ctx)
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
