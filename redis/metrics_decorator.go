package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/releaseband/metrics/opencensus/views"

	"go.opencensus.io/stats/view"

	"github.com/releaseband/metrics/measure"

	"go.opencensus.io/tag"
)

var (
	methodKey = tag.MustNewKey("method")
	latency   *measure.LatencyMeasure
)

func RedisMetricsView() *view.View {
	latency = measure.NewLatencyMeasure("redis", "redis")

	return views.MakeLatencyView("redis_latency", "redis delay measures", latency.Measure(),
		[]tag.Key{methodKey})
}

//deprecated
type RedisMetricsDecorator struct {
	client RedisClient
}

func NewRedisMetricsDecorator(client RedisClient) *RedisMetricsDecorator {
	return &RedisMetricsDecorator{
		client: client,
	}
}

func wrapToLatencyContext(ctx context.Context, method string) context.Context {
	ctx, _ = tag.New(ctx, tag.Insert(methodKey, method))
	return ctx
}

func record(ctx context.Context, start time.Time) {
	if latency != nil {
		latency.Record(ctx, measure.End(start))
	}
}

func (r RedisMetricsDecorator) ClientType() uint8 {
	return r.client.ClientType()
}

func (r *RedisMetricsDecorator) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	ctx = wrapToLatencyContext(ctx, "Set")
	start := measure.Start()
	err := r.client.Set(ctx, key, value, expiration)
	record(ctx, start)

	return err
}

func (r *RedisMetricsDecorator) Get(ctx context.Context, key string) (string, error) {
	ctx = wrapToLatencyContext(ctx, "Get")
	start := measure.Start()
	resp, err := r.client.Get(ctx, key)
	record(ctx, start)

	return resp, err
}

func (r *RedisMetricsDecorator) RPush(ctx context.Context, listKey string, val ...interface{}) error {
	ctx = wrapToLatencyContext(ctx, "RPush")
	start := measure.Start()
	err := r.client.RPush(ctx, listKey, val)
	record(ctx, start)

	return err
}

func (r *RedisMetricsDecorator) LTrim(ctx context.Context, listKey string, start, stop int64) error {
	ctx = wrapToLatencyContext(ctx, "LTrim")
	startTime := measure.Start()
	err := r.client.LTrim(ctx, listKey, start, stop)
	record(ctx, startTime)

	return err
}

func (r *RedisMetricsDecorator) LRange(ctx context.Context, listKey string, start, stop int64) ([]string, error) {
	ctx = wrapToLatencyContext(ctx, "LRange")
	startTime := measure.Start()
	resp, err := r.client.LRange(ctx, listKey, start, stop)
	record(ctx, startTime)

	return resp, err
}

func (r *RedisMetricsDecorator) Ping(ctx context.Context) error {
	ctx = wrapToLatencyContext(ctx, "Ping")
	start := measure.Start()
	err := r.client.Ping(ctx)
	record(ctx, start)

	return err
}

func (r *RedisMetricsDecorator) SlotsCount(ctx context.Context) (int, error) {
	return r.client.SlotsCount(ctx)
}

func (r *RedisMetricsDecorator) LLen(ctx context.Context, listKey string) (int64, error) {
	ctx = wrapToLatencyContext(ctx, "LLen")
	start := measure.Start()
	resp, err := r.client.LLen(ctx, listKey)
	record(ctx, start)

	return resp, err
}

func (r *RedisMetricsDecorator) ReadinessChecker(timeout time.Duration) *ReadinessChecker {
	return NewReadinessChecker(timeout, r.Ping)
}

func (r *RedisMetricsDecorator) Watch(ctx context.Context, txf func(tx *redis.Tx) error, key ...string) error {
	return r.client.Watch(ctx, txf, key...)
}

func (r *RedisMetricsDecorator) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	ctx = wrapToLatencyContext(ctx, "SetEX")
	start := measure.Start()
	err := r.client.SetEX(ctx, key, value, expiration)
	record(ctx, start)

	return err
}

func (r *RedisMetricsDecorator) Del(ctx context.Context, key string) error {
	ctx = wrapToLatencyContext(ctx, "Del")
	start := measure.Start()
	err := r.client.Del(ctx, key)
	record(ctx, start)

	return err
}

func (r *RedisMetricsDecorator) Impl() redis.Cmdable {
	return r.client.Impl()
}

func (r *RedisMetricsDecorator) Uc() redis.UniversalClient {
	return r.client.Uc()
}

func (r *RedisMetricsDecorator) Incr(ctx context.Context, key string) (int64, error) {
	ctx = wrapToLatencyContext(ctx, "Incr")
	start := measure.Start()
	resp, err := r.client.Incr(ctx, key)
	record(ctx, start)

	return resp, err
}

func (r *RedisMetricsDecorator) HSet(ctx context.Context, key string, val ...interface{}) error {
	ctx = wrapToLatencyContext(ctx, "HSet")
	start := measure.Start()
	err := r.client.HSet(ctx, key, val...)
	record(ctx, start)

	return err
}

func (r *RedisMetricsDecorator) HGet(ctx context.Context, key, field string) (string, error) {
	ctx = wrapToLatencyContext(ctx, "HGet")
	start := measure.Start()
	resp, err := r.client.HGet(ctx, key, field)
	record(ctx, start)

	return resp, err
}

func (r *RedisMetricsDecorator) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	ctx = wrapToLatencyContext(ctx, "HGetAll")
	start := measure.Start()
	resp, err := r.client.HGetAll(ctx, key)
	record(ctx, start)

	return resp, err
}

func (r *RedisMetricsDecorator) HDel(ctx context.Context, key string, field ...string) error {
	ctx = wrapToLatencyContext(ctx, "HDel")
	start := measure.Start()
	err := r.client.HDel(ctx, key, field...)
	record(ctx, start)

	return err
}
