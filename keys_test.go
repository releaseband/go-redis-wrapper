package go_redis_wrapper

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/howeyc/crc16"
)

func Test_keys(t *testing.T) {
	const (
		shards = 3
		mode   = 16384
	)

	check := func(keys []string) {
		for _, key := range keys {
			fmt.Println(key, ":")
			for i := 0; i < shards; i++ {
				q := key + ":" + strconv.Itoa(i*9)
				sum := crc16.ChecksumSCSI([]byte(q)) % mode
				fmt.Println(sum)
			}
		}
	}

	keys := []string{
		"history",
		"failed_histories",
		"invalid_histories",
		"manual_fix",
	}

	check(keys)
}
