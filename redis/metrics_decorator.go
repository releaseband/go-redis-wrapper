package redis

import (
	"context"
	"time"
)

const (
	methodHGetAll = "HGetAll"
	methodGet     = "Get"
	methodHSet    = "HSet"
	methodSet     = "Set"
	methodHDel    = "HDel"
	methodDel     = "Del"
	methodPing    = "Ping"
)

type Metrics interface {
	MeasureLatency(ctx context.Context, entity, method string, callback func())
}

type RedisMetricsDecorator struct {
	client  RedisClient
	measure func(ctx context.Context, method string, callback func())
}

func measure(metrics Metrics, entity string) func(ctx context.Context, method string, callback func()) {
	return func(ctx context.Context, method string, callback func()) {
		metrics.MeasureLatency(ctx, entity, method, callback)
	}
}

func NewRedisMetricsDecorator(client RedisClient, metrics Metrics) *RedisMetricsDecorator {
	return &RedisMetricsDecorator{
		client:  client,
		measure: measure(metrics, client.Entity()),
	}
}

func (r RedisMetricsDecorator) HGetAll(ctx context.Context, groupKey string) (map[string]string, error) {
	var (
		resp map[string]string
		err  error
	)

	callback := func() {
		resp, err = r.client.HGetAll(ctx, groupKey)
	}

	r.measure(ctx, methodHGetAll, callback)

	return resp, err
}

func (r RedisMetricsDecorator) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var err error

	callback := func() {
		err = r.client.Set(ctx, key, value, expiration)
	}

	r.measure(ctx, methodSet, callback)

	return err
}

func (r RedisMetricsDecorator) HSet(ctx context.Context, key, field string, value interface{}) error {
	var err error

	callback := func() {
		err = r.client.HSet(ctx, key, field, value)
	}

	r.measure(ctx, methodHSet, callback)

	return err
}

func (r RedisMetricsDecorator) Get(ctx context.Context, key string) (string, error) {
	var (
		resp string
		err  error
	)

	callback := func() {
		resp, err = r.client.Get(ctx, key)
	}

	r.measure(ctx, methodGet, callback)

	return resp, err
}

func (r RedisMetricsDecorator) Del(ctx context.Context, key string) error {
	var err error

	callback := func() {
		err = r.client.Del(ctx, key)
	}

	r.measure(ctx, methodDel, callback)

	return err
}

func (r RedisMetricsDecorator) HDel(ctx context.Context, key, field string) error {
	var err error

	callback := func() {
		err = r.client.HDel(ctx, key, field)
	}

	r.measure(ctx, methodHDel, callback)

	return err
}

func (r RedisMetricsDecorator) Ping() error {
	var err error
	callback := func() {
		err = r.client.Ping()
	}

	r.measure(pingCtx, methodPing, callback)

	return err
}

func (r RedisMetricsDecorator) Status() (interface{}, error) {
	var (
		resp interface{}
		err  error
	)

	callback := func() {
		resp, err = r.client.Status()
	}

	r.measure(pingCtx, methodPing, callback)

	return resp, err
}

func (r RedisMetricsDecorator) Entity() string {
	return r.client.Entity()
}