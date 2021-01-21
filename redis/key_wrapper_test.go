package redis

import (
	"strconv"
	"testing"
)

func TestKeyPostfix_Next(t *testing.T) {
	const key = "key"

	makeKeyWrapper := func(count int) *KeyWrapper {
		return &KeyWrapper{
			shardsCount: count,
			postfixes:   makePostfixes(count),
		}
	}

	t.Run("shards count <= 1", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			kp := makeKeyWrapper(i)
			exp := key + ""
			for i := 0; i < 100; i++ {
				got := kp.WrapKey(key)

				if got != exp {
					t.Fatalf("exp=%s | got=%s", exp, got)
				}
			}
		}
	})

	t.Run("EmptyKeyWrapper", func(t *testing.T) {
		const key = "key"

		kw := EmptyKeyWrapper()
		for i := 0; i < 100; i++ {
			got := kw.WrapKey(key)
			if got != key {
				t.Fatalf("exp=%s | got=%s", key, got)
			}
		}
	})

	t.Run("shard count > 0", func(t *testing.T) {
		const shardsCount = 10
		kp := makeKeyWrapper(shardsCount)

		if kp.shardsCount != shardsCount {
			t.Fatalf("shards count should be equal %d", shardsCount)
		}

		if len(kp.postfixes) != shardsCount {
			t.Fatalf("postfixes count should be equeal %d", shardsCount)
		}

		index := 0
		for i := 0; i < shardsCount*2-1; i++ {
			index++

			if i == shardsCount-1 {
				index = 0
			}

			exp := key + ":" + strconv.Itoa(index*multiplier)
			got := kp.WrapKey(key)
			if got != exp {
				t.Fatalf("exp=%s | got=%s", exp, got)
			}
		}
	})
}
