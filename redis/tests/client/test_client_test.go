package client

import (
	"context"
	"strconv"
	"testing"

	"github.com/go-redis/redis/v8"

	"github.com/releaseband/go-redis-wrapper/redis/tests/internal"
)

func TestTestClient(t *testing.T) {
	const (
		key     = "key"
		val     = "value"
		listKey = "list_key"
	)

	checkErr := internal.ErrorChecker(t)
	client, err := MakeTestClient()
	checkErr(nil, err)

	ctx := context.TODO()

	internal.TestCase(t, "redis test client")(
		func(t *testing.T) {
			_, err = client.Get(ctx, key)
			checkErr(redis.Nil, err)

			list, err := client.LRange(ctx, key, 0, 1)
			checkErr(nil, err)
			if len(list) != 0 {
				t.Fatalf("list len should be 0")
			}
		},

		func(t *testing.T) {
			t.Run("Set, Get", func(t *testing.T) {
				err := client.Set(ctx, key, val, 0)
				checkErr(nil, err)

				res, err := client.Get(ctx, key)
				checkErr(nil, err)
				if res != val {
					t.Fatalf("expRes should be gotRes")
				}
			})

			t.Run("List", func(t *testing.T) {
				const count = 10

				makeVal := func(i int) string {
					return val + strconv.Itoa(i)
				}

				t.Run("RPush", func(t *testing.T) {
					for i := 0; i < count; i++ {
						err := client.RPush(ctx, listKey, makeVal(i))
						checkErr(nil, err)
					}
				})

				t.Run("LLen", func(t *testing.T) {
					res, err := client.LLen(ctx, listKey)
					checkErr(nil, err)
					if res != count {
						t.Fatal("expRes should be equal gotRes")
					}
				})

				t.Run("LRange", func(t *testing.T) {
					results, err := client.LRange(ctx, listKey, 0, -1)
					checkErr(nil, err)
					if len(results) != count {
						t.Fatal("len(gotResult) != len(expResult)")
					}

					for i, got := range results {
						exp := makeVal(i)
						if got != exp {
							t.Fatalf("expResul := '%s', gotResult = '%s'", exp, got)
						}
					}
				})

				t.Run("LTrim", func(t *testing.T) {
					err = client.LTrim(ctx, listKey, 0, -1)
					checkErr(nil, err)
				})
			})
		})
}
