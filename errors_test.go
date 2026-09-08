package go_redis_wrapper_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	goredis "github.com/releaseband/go-redis-wrapper/v2"
)

func TestIsNotFoundErr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error is not a not-found error",
			err:  nil,
			want: false,
		},
		{
			name: "redis.Nil is a not-found error",
			err:  redis.Nil,
			want: true,
		},
		{
			name: "wrapped redis.Nil is a not-found error",
			err:  fmt.Errorf("get key: %w", redis.Nil),
			want: true,
		},
		{
			name: "unrelated error is not a not-found error",
			err:  errors.New("boom"),
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, goredis.IsNotFoundErr(tc.err))
		})
	}
}
