package go_redis_wrapper

import (
	"errors"
	"fmt"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestIsNotFoundErr(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "redis.Nil error",
			err:      redis.Nil,
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "wrapped redis.Nil error",
			err:      fmt.Errorf("failed: %w", redis.Nil),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFoundErr(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFoundErr(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}
