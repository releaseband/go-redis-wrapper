package internal

import "strconv"

const multiplier = 8

type keyPostfix struct {
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

func newKeyPostfix(shardsCount int) *keyPostfix {
	return &keyPostfix{
		shardsCount: shardsCount,
		postfixes:   makePostfixes(shardsCount),
	}
}

func (b *keyPostfix) Next() string {
	if b.shardsCount == 0 {
		return ""
	}

	b.i++
	if b.i >= b.shardsCount {
		b.i = 0
	}

	return b.postfixes[b.i]
}
