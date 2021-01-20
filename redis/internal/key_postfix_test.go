package internal

import (
	"strconv"
	"testing"
)

func TestKeyPostfix_Next(t *testing.T) {
	t.Run("shards count == 0", func(t *testing.T) {
		kp := newKeyPostfix(0)
		exp := ""
		for i := 0; i < 100; i++ {
			got := kp.Next()

			if got != exp {
				t.Fatalf("exp=%s | got=%s", exp, got)
			}
		}
	})

	t.Run("shard count > 0", func(t *testing.T) {
		const shardsCount = 10

		kp := newKeyPostfix(shardsCount)
		if kp.shardsCount != shardsCount {
			t.Fatalf("shards count should be equal %d", shardsCount)
		}

		if len(kp.postfixes) != shardsCount {
			t.Fatalf("postfixes count should be equeal %d", shardsCount)
		}

		key := 0
		for i := 0; i < shardsCount*2-1; i++ {
			key++

			if i == shardsCount-1 {
				key = 0
			}

			exp := ":" + strconv.Itoa(key*multiplier)
			got := kp.Next()
			if got != exp {
				t.Fatalf("exp=%s | got=%s", exp, got)
			}
		}
	})
}
