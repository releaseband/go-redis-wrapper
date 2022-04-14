package go_redis_wrapper

import (
	"errors"
	"github.com/go-redis/redis/v8"
)

var (
	ErrPingNotImplemented       = errors.New("ping not implemented for this redis client type")
	ErrSlotsCountNotImplemented = errors.New("slots count not implemented for this client type")
	ErrCastToClusterClient      = errors.New("cast to redis cluster client failed")
)

func IsNotFoundErr(err error) bool {
	return err != nil && err == redis.Nil
}
