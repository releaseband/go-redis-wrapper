package goredis_test

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	goredis "github.com/releaseband/go-redis-wrapper/v3"
)

// invalidClientType is deliberately not one of the SimpleClientType /
// ClusterClientType / TestClientType constants.
const invalidClientType uint8 = 0

func TestNewClient(t *testing.T) {
	t.Parallel()

	client := goredis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})

	require.NotNil(t, client)
	assert.NotNil(t, client.UniversalClient)
	assert.Equal(t, goredis.SimpleClientType, client.Type)
}

func TestNewClusterClient(t *testing.T) {
	t.Parallel()

	client := goredis.NewClusterClient(&redis.ClusterOptions{Addrs: []string{"127.0.0.1:0"}})

	require.NotNil(t, client)
	assert.NotNil(t, client.UniversalClient)
	assert.Equal(t, goredis.ClusterClientType, client.Type)
}

func TestStartMiniRedis(t *testing.T) {
	t.Parallel()

	client := mustStartMiniRedis(t)

	assert.Equal(t, goredis.TestClientType, client.Type)
	assert.NoError(t, client.Ping(context.Background()))
}

func TestClientAdapter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		clientType uint8
		checkErr   func(t *testing.T, err error)
	}{
		{
			name:       "simple client type is accepted",
			clientType: goredis.SimpleClientType,
			checkErr:   noErr,
		},
		{
			name:       "cluster client type is accepted",
			clientType: goredis.ClusterClientType,
			checkErr:   noErr,
		},
		{
			name:       "test client type is accepted",
			clientType: goredis.TestClientType,
			checkErr:   noErr,
		},
		{
			name:       "invalid client type is rejected",
			clientType: invalidClientType,
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, goredis.ErrInvalidClientType)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			uc := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
			t.Cleanup(func() { _ = uc.Close() })

			client, err := goredis.ClientAdapter(uc, tc.clientType)
			tc.checkErr(t, err)

			if err != nil {
				assert.Nil(t, client)
				return
			}

			require.NotNil(t, client)
			assert.Equal(t, tc.clientType, client.Type)
		})
	}
}

func TestCastToRedisCluster(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		client   redis.UniversalClient
		checkErr func(t *testing.T, err error)
	}{
		{
			name:     "cluster client casts successfully",
			client:   redis.NewClusterClient(&redis.ClusterOptions{Addrs: []string{"127.0.0.1:0"}}),
			checkErr: noErr,
		},
		{
			name:   "simple client fails to cast",
			client: redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"}),
			checkErr: func(t *testing.T, err error) {
				require.ErrorIs(t, err, goredis.ErrCastToClusterClient)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Cleanup(func() { _ = tc.client.Close() })

			cluster, err := goredis.CastToRedisCluster(tc.client)
			tc.checkErr(t, err)

			if err != nil {
				assert.Nil(t, cluster)
				return
			}

			assert.Same(t, tc.client, cluster)
		})
	}
}

func TestSimplePing(t *testing.T) {
	t.Parallel()

	client := mustStartMiniRedis(t)

	assert.NoError(t, goredis.SimplePing(context.Background(), client.UniversalClient))
}

func TestClient_Ping(t *testing.T) {
	t.Parallel()

	t.Run("test client type pings successfully", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)

		assert.NoError(t, client.Ping(context.Background()))
	})

	t.Run("cluster type backed by a non-cluster client fails to cast", func(t *testing.T) {
		t.Parallel()

		uc := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
		t.Cleanup(func() { _ = uc.Close() })

		client, err := goredis.ClientAdapter(uc, goredis.ClusterClientType)
		require.NoError(t, err)

		require.ErrorIs(t, client.Ping(context.Background()), goredis.ErrCastToClusterClient)
	})

	t.Run("unknown type is not implemented", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)
		client.Type = invalidClientType

		require.ErrorIs(t, client.Ping(context.Background()), goredis.ErrPingNotImplemented)
	})
}

func TestClient_Status(t *testing.T) {
	t.Parallel()

	t.Run("healthy client reports ok", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)

		status, err := client.Status()
		require.NoError(t, err)
		assert.Equal(t, "ok", status)
	})

	t.Run("ping failure is propagated", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)
		client.Type = invalidClientType

		status, err := client.Status()
		require.ErrorIs(t, err, goredis.ErrPingNotImplemented)
		assert.Nil(t, status)
	})
}

func TestClient_SlotsCount(t *testing.T) {
	t.Parallel()

	t.Run("test client type has no slots", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)

		count, err := client.SlotsCount(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("cluster type backed by a non-cluster client fails to cast", func(t *testing.T) {
		t.Parallel()

		uc := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
		t.Cleanup(func() { _ = uc.Close() })

		client, err := goredis.ClientAdapter(uc, goredis.ClusterClientType)
		require.NoError(t, err)

		count, err := client.SlotsCount(context.Background())
		require.ErrorIs(t, err, goredis.ErrCastToClusterClient)
		assert.Equal(t, 0, count)
	})

	t.Run("unknown type is not implemented", func(t *testing.T) {
		t.Parallel()

		client := mustStartMiniRedis(t)
		client.Type = invalidClientType

		count, err := client.SlotsCount(context.Background())
		require.ErrorIs(t, err, goredis.ErrSlotsCountNotImplemented)
		assert.Equal(t, 0, count)
	})
}
