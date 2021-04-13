package redis

import (
	"context"
	"time"
)

type ReadinessChecker struct {
	timeout time.Duration
	ping    func(ctx context.Context) error
}

func NewReadinessChecker(timeout time.Duration, ping func(ctx context.Context) error) *ReadinessChecker {
	return &ReadinessChecker{
		timeout: timeout,
		ping:    ping,
	}
}

func (c *ReadinessChecker) Status() (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	if err := c.ping(ctx); err != nil {
		return nil, err
	}

	return "ok", nil
}
