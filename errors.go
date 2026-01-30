package go_redis_wrapper

import (
	"errors"

	"github.com/redis/go-redis/v9"
)

var (
	// ErrPingNotImplemented indicates that the Ping method is not implemented for the given client type.
	ErrPingNotImplemented = errors.New("ping not implemented for this redis client type")
	// ErrClusterPingNotImplemented indicates that the ClusterPing method is not implemented for the given client type.
	ErrSlotsCountNotImplemented = errors.New("slots count not implemented for this client type")
	// ErrCastToSimpleClient indicates a failure to cast to a simple Redis client.
	ErrCastToClusterClient = errors.New("cast to redis cluster client failed")
	// ErrInvalidClientType indicates that the provided client type is invalid.
	ErrInvalidClientType = errors.New("invalid client type")
	// ErrLockFailed indicates that acquiring the lock has failed.
	ErrUnlockStatusIsFailure = errors.New("unlock status is failure")
)

// IsNotFoundErr checks if the provided error indicates a "not found" condition in Redis.
func IsNotFoundErr(err error) bool {
	return err != nil && err == redis.Nil
}
