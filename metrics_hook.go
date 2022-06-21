package go_redis_wrapper

import (
	"context"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"os"
	"time"
)

const (
	prefixEnvKey       = "RB_SERVICE"
	redisHistogramName = "redis_duration_seconds"
	commandKey         = "command"
)

type timeCtx struct{}

var (
	meter      = global.MeterProvider().Meter(getPrefix() + ".redis")
	measure, _ = meter.SyncFloat64().Histogram(
		getPrefix()+"."+redisHistogramName,
		instrument.WithDescription("redis duration in seconds"),
		instrument.WithUnit("sec"),
	)
)

func getPrefix() string {
	return os.Getenv(prefixEnvKey)
}

type redisHookMetrics struct {
}

func newRedisHookMetrics() redisHookMetrics {
	return redisHookMetrics{}
}

func (r redisHookMetrics) BeforeProcess(ctx context.Context, _ redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, timeCtx{}, time.Now()), nil
}

func (r redisHookMetrics) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	start, ok := ctx.Value(timeCtx{}).(time.Time)
	if !ok {
		return nil
	}

	attr := attribute.String(commandKey, cmd.Name())

	measure.Record(ctx, time.Since(start).Seconds(), attr)
	return nil
}

func (r redisHookMetrics) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (r redisHookMetrics) AfterProcessPipeline(_ context.Context, _ []redis.Cmder) error {
	return nil
}