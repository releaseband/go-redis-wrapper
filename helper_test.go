package goredis_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	goredis "github.com/releaseband/go-redis-wrapper/v3"
)

func noErr(t *testing.T, err error) {
	t.Helper()
	require.NoError(t, err)
}

func mustStartMiniRedis(tb testing.TB) *goredis.Client {
	tb.Helper()

	client, err := goredis.StartMiniRedis()
	require.NoError(tb, err)

	tb.Cleanup(func() { _ = client.Close() })

	return client
}
