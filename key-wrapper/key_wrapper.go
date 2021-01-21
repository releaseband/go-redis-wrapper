package key_wrapper

import (
	"strconv"
)

const multiplier = 8

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

func NewKeyWrapper(slotsCount int) *KeyWrapper {
	return &KeyWrapper{
		shardsCount: slotsCount,
		postfixes:   makePostfixes(slotsCount),
	}
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
