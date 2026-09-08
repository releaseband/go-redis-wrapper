package goredis_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	goredis "github.com/releaseband/go-redis-wrapper/v3"
)

// fastFailOptions makes a single, near-instant lock attempt so contention
// tests don't wait through redsync's default retry backoff.
func fastFailOptions() []redsync.Option {
	return []redsync.Option{
		redsync.WithTries(1),
		redsync.WithRetryDelay(time.Millisecond),
	}
}

func TestClient_Lock(t *testing.T) {
	t.Parallel()

	t.Run("acquires a free lock", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)

		mutex, err := client.Lock(context.Background(), "lock-happy-path")
		require.NoError(t, err)
		require.NotNil(t, mutex)

		ok, err := mutex.UnlockContext(context.Background())
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("acquires a free lock with custom options", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)

		mutex, err := client.Lock(
			context.Background(),
			"lock-with-options",
			redsync.WithExpiry(5*time.Second),
			redsync.WithTries(3),
		)
		require.NoError(t, err)
		require.NotNil(t, mutex)

		ok, err := mutex.UnlockContext(context.Background())
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("fails when the key is already locked", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)
		key := "lock-contended"

		mutex, err := client.Lock(context.Background(), key)
		require.NoError(t, err)
		t.Cleanup(func() { _, _ = mutex.UnlockContext(context.Background()) })

		_, err = client.Lock(context.Background(), key, fastFailOptions()...)
		require.Error(t, err)
	})
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

	t.Run("acquires a free lock with custom options", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)

		unlock, err := client.LockKey(
			context.Background(),
			"lock-key-with-options",
			redsync.WithExpiry(5*time.Second),
			redsync.WithTries(3),
		)
		require.NoError(t, err)
		require.NotNil(t, unlock)

		assert.NoError(t, unlock(context.Background()))
	})

	t.Run("fails when the key is already locked", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)
		key := "lock-key-contended"

		unlock, err := client.LockKey(context.Background(), key)
		require.NoError(t, err)
		t.Cleanup(func() { _ = unlock(context.Background()) })

		_, err = client.LockKey(context.Background(), key, fastFailOptions()...)
		require.Error(t, err)
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

func TestClient_TryLock(t *testing.T) {
	t.Parallel()

	t.Run("acquires a free lock", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)

		mutex, err := client.TryLock(context.Background(), "try-lock-happy-path")
		require.NoError(t, err)
		require.NotNil(t, mutex)

		ok, err := mutex.UnlockContext(context.Background())
		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("fails immediately when the key is already locked", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)
		key := "try-lock-contended"

		mutex, err := client.TryLock(context.Background(), key)
		require.NoError(t, err)
		t.Cleanup(func() { _, _ = mutex.UnlockContext(context.Background()) })

		_, err = client.TryLock(context.Background(), key)
		require.ErrorIs(t, err, goredis.ErrResourceBusy)
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

func BenchmarkClient_Ping(b *testing.B) {
	client := mustStartMiniRedis(b)

	ctx := context.Background()

	b.ResetTimer()
	for range b.N {
		_ = client.Ping(ctx)
	}
}

func BenchmarkClient_Lock(b *testing.B) {
	client := mustStartMiniRedis(b)

	ctx := context.Background()

	b.ResetTimer()
	for i := range b.N {
		key := "bench-lock-" + strconv.Itoa(i)

		mutex, err := client.Lock(ctx, key)
		if err != nil {
			b.Errorf("Lock failed: %v", err)
			continue
		}
		_, _ = mutex.UnlockContext(ctx)
	}
}
