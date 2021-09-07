package redis

import (
	"testing"

	"github.com/go-redis/redis/v8"
)

func Test_implementation(t *testing.T) {
	var client RedisClient
	client = NewSimple(&redis.Options{})
	client = NewCluster(&redis.ClusterOptions{})
	client = NewRedisMetricsDecorator(client)
}
