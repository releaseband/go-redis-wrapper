package redis

import (
	"context"
	"time"
)

const (
	methodLRange = "LRange"
	methodGet    = "Get"
	methodRPush  = "RPush"
	methodSet    = "Set"
	methodLTrim  = "LTrim"
	methodPing   = "Ping"
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

func NewRedisMetricsDecorator(client RedisClient, metrics Metrics) RedisMetricsDecorator {
	return RedisMetricsDecorator{
		client:  client,
		measure: measure(metrics, client.Entity()),
	}
}

func (r RedisMetricsDecorator) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var err error

	callback := func() {
		err = r.client.Set(ctx, key, value, expiration)
	}

	r.measure(ctx, methodSet, callback)

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

func (r RedisMetricsDecorator) RPush(ctx context.Context, listKey string, val ...interface{}) error {
	var err error

	callback := func() {
		err = r.client.RPush(ctx, listKey, val)
	}

	r.measure(ctx, methodRPush, callback)

	return err
}

func (r RedisMetricsDecorator) LTrim(ctx context.Context, listKey string, start, stop int64) error {
	var err error

	callback := func() {
		err = r.client.LTrim(ctx, listKey, start, stop)
	}

	r.measure(ctx, methodLTrim, callback)

	return err
}

func (r RedisMetricsDecorator) LRange(ctx context.Context, listKey string, start, stop int64) ([]string, error) {
	var (
		resp []string
		err  error
	)

	callback := func() {
		resp, err = r.client.LRange(ctx, listKey, start, stop)
	}

	r.measure(ctx, methodLRange, callback)

	return resp, err
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

func (r RedisMetricsDecorator) SlotsCount(ctx context.Context) (int, error) {
	return r.client.SlotsCount(ctx)
}
