package internal

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
)

func NewRedSync(client redis.UniversalClient) *redsync.Redsync {
	return redsync.New(goredis.NewPool(client))
}

func Lock(ctx context.Context, rs *redsync.Redsync, key string, options ...redsync.Option) (*redsync.Mutex, error) {
	mutex := rs.NewMutex(key, options...)

	if err := mutex.LockContext(ctx); err != nil {
		return nil, err
	}

	return mutex, nil
}
