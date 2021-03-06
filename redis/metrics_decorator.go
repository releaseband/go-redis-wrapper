package redis

import (
	"context"
	"time"

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
