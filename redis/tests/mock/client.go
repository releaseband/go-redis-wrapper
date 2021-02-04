package mock

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

const entity = "redis_mock"

type RedisClientDb struct {
	db *redis.Client
}

func newRedisClientDb(db *redis.Client) *RedisClientDb {
	return &RedisClientDb{db: db}
}

func (r RedisClientDb) RPush(ctx context.Context, listKey string, val ...interface{}) error {
	return r.db.RPush(ctx, listKey, val).Err()
}

func (r RedisClientDb) LTrim(ctx context.Context, listKey string, start, stop int64) error {
	return r.db.LTrim(ctx, listKey, start, stop).Err()
}

func (r RedisClientDb) LRange(ctx context.Context, listKey string, start, stop int64) ([]string, error) {
	return r.db.LRange(ctx, listKey, start, stop).Result()
}

func (r RedisClientDb) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.db.Set(ctx, key, value, expiration).Err()
}

func (r RedisClientDb) Get(ctx context.Context, key string) (string, error) {
	return r.db.Get(ctx, key).Result()
}

func (r RedisClientDb) Status() (interface{}, error) {
	err := r.db.Ping(context.TODO()).Err()
	if err != nil {
		return nil, err
	}

	return "OK", nil
}

func (r RedisClientDb) Entity() string {
	return entity
}

func (r RedisClientDb) SlotsCount(ctx context.Context) (int, error) {
	slots, err := r.db.ClusterSlots(ctx).Result()
	if err != nil {
		return 0, err
	}

	return len(slots), nil
}
