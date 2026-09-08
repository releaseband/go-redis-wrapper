package go_redis_wrapper_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	goredis "github.com/releaseband/go-redis-wrapper/v2"
)

// fastFailOptions makes a single, near-instant lock attempt so contention
// tests don't wait through redsync's default retry backoff.
func fastFailOptions() []redsync.Option {
	return []redsync.Option{
		redsync.WithTries(1),
		redsync.WithRetryDelay(time.Millisecond),
	}
}

func TestClient_LockKey(t *testing.T) {
	t.Parallel()

	t.Run("acquires a free lock and releases it", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)

		unlock, err := client.LockKey(context.Background(), "lock-key-happy-path")
		require.NoError(t, err)
		require.NotNil(t, unlock)

		assert.NoError(t, unlock(context.Background()))
	})

	t.Run("fails when the key is already locked", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)
		key := "lock-key-contended"

		unlock, err := client.TryLockKey(context.Background(), key)
		require.NoError(t, err)
		t.Cleanup(func() { _ = unlock(context.Background()) })

		_, err = client.LockKey(context.Background(), key, fastFailOptions()...)
		require.Error(t, err)
		assert.NotErrorIs(t, err, goredis.ErrResourceBusy)
	})

	t.Run("unlocking twice reports the second unlock as a failure", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)

		unlock, err := client.LockKey(context.Background(), "lock-key-double-unlock")
		require.NoError(t, err)

		require.NoError(t, unlock(context.Background()))
		require.Error(t, unlock(context.Background()))
	})
}

func TestClient_TryLockKey(t *testing.T) {
	t.Parallel()

	t.Run("acquires a free lock and releases it", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)

		unlock, err := client.TryLockKey(context.Background(), "try-lock-key-happy-path")
		require.NoError(t, err)
		require.NotNil(t, unlock)

		assert.NoError(t, unlock(context.Background()))
	})

	t.Run("fails immediately when the key is already locked", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)
		key := "try-lock-key-contended"

		unlock, err := client.TryLockKey(context.Background(), key)
		require.NoError(t, err)
		t.Cleanup(func() { _ = unlock(context.Background()) })

		_, err = client.TryLockKey(context.Background(), key)
		require.ErrorIs(t, err, goredis.ErrResourceBusy)
	})

	t.Run("unlocking twice reports the second unlock as a failure", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)

		unlock, err := client.TryLockKey(context.Background(), "try-lock-key-double-unlock")
		require.NoError(t, err)

		require.NoError(t, unlock(context.Background()))
		require.Error(t, unlock(context.Background()))
	})
}
