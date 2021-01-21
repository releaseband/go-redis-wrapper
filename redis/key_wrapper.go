package redis

import (
	"context"
	"errors"
	"fmt"
	"strconv"
)

const multiplier = 8

var (
	ErrKeyWrapperOnlyForCluster = errors.New("key wrapper only for cluster")
)

type KeyWrapper struct {
	i           int
	shardsCount int
	postfixes   []string
}

func getPostfix(i int) string {
	return ":" + strconv.Itoa(i*multiplier)
}

func makePostfixes(count int) []string {
	postfixes := make([]string, count)
	for i := 0; i < count; i++ {
		postfixes[i] = getPostfix(i)
	}

	return postfixes
}

func NewKeyWrapper(ctx context.Context, client BaseRedisClient) (*KeyWrapper, error) {
	if client.Entity() != entityCluster {
		return nil, ErrKeyWrapperOnlyForCluster
	}

	slots, err := client.impl.ClusterSlots(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("ClusterSlots failed: %w", err)
	}

	return &KeyWrapper{
		shardsCount: len(slots),
		postfixes:   makePostfixes(len(slots)),
	}, nil
}

func (b *KeyWrapper) WrapKey(key string) string {
	var postfix string
	if b.shardsCount > 1 {
		b.i++
		if b.i >= b.shardsCount {
			b.i = 0
		}

		postfix = b.postfixes[b.i]
	}

	return key + postfix
}
