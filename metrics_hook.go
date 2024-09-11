package go_redis_wrapper

import (
	"context"

	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const (
	prefixEnvKey       = "RB_SERVICE"
	redisHistogramName = "redis_duration_seconds"
	commandKey         = "command"
)

type timeCtx struct{}

var (
	meter      = sdkmetric.NewMeterProvider().Meter(getPrefix() + ".redis")
	measure, _ = meter.Float64Histogram(
		getPrefix()+"."+redisHistogramName,
		metric.WithDescription("redis duration in seconds"),
		metric.WithUnit("sec"),
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

	attr := metric.WithAttributes(attribute.String(commandKey, cmd.Name()))

	measure.Record(ctx, time.Since(start).Seconds(), attr)
	return nil
}

func (r redisHookMetrics) BeforeProcessPipeline(ctx context.Context, _ []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (r redisHookMetrics) AfterProcessPipeline(_ context.Context, _ []redis.Cmder) error {
	return nil
}
