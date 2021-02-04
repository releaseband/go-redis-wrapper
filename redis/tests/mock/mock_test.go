package mock

import (
	"context"
	"errors"
	"testing"

	"github.com/releaseband/go-redis-wrapper/redis/tests/internal"
)

func TestRedisClientMock(t *testing.T) {
	const (
		key   = "testKey"
		val   = "value"
		start = 1
		stop  = 20
	)

	ctx := context.TODO()
	client, mock := NewRedisClientMock(true)
	checkErr := internal.ErrorChecker(t)

	internal.TestCase(t, "Set")(
		func(t *testing.T) {
			expErr := errors.New("set failed")
			mock.Set(key, val, 0)(expErr)
			err := client.Set(ctx, key, val, 0)
			checkErr(expErr, err)
		},

		func(t *testing.T) {
			mock.Set(key, val, 0)(nil)
			err := client.Set(ctx, key, val, 0)
			checkErr(nil, err)
		},
	)

	internal.TestCase(t, "Get")(
		func(t *testing.T) {
			expErr := errors.New("get failed")
			mock.Get(key)("", expErr)
			_, err := client.Get(ctx, key)
			checkErr(expErr, err)

		},

		func(t *testing.T) {
			const expRes = "value"
			mock.Get(key)(expRes, nil)
			gotRes, err := client.Get(ctx, key)
			checkErr(nil, err)
			if gotRes != expRes {
				t.Fatalf("expRes = '%s', gotRes = '%s'", expRes, gotRes)
			}
		},
	)

	internal.TestCase(t, "RPush")(
		func(t *testing.T) {
			expErr := errors.New("rPush failed")
			mock.RPush(key, val)(expErr)
			err := client.RPush(ctx, key, val)
			checkErr(expErr, err)

		},

		func(t *testing.T) {
			mock.RPush(key, val)(nil)
			err := client.RPush(ctx, key, val)
			checkErr(nil, err)
		},
	)

	internal.TestCase(t, "LRange")(
		func(t *testing.T) {
			expErr := errors.New("LRange failed")
			mock.LRange(key, start, stop)(nil, expErr)
			_, err := client.LRange(ctx, key, start, stop)
			checkErr(expErr, err)
		},

		func(t *testing.T) {
			expResult := []string{
				"1", "2",
			}

			mock.LRange(key, start, stop)(expResult, nil)
			gotRes, err := client.LRange(ctx, key, start, stop)
			checkErr(nil, err)
			if len(gotRes) != len(expResult) {
				t.Fatal("gotResult != expResult")
			}
		},
	)

	internal.TestCase(t, "LTrim")(
		func(t *testing.T) {
			expErr := errors.New("lTrim failed")
			mock.LTrim(key, start, stop)(expErr)
			err := client.LTrim(ctx, key, start, stop)
			checkErr(expErr, err)
		},

		func(t *testing.T) {
			mock.LTrim(key, start, stop)(nil)
			err := client.LTrim(ctx, key, start, stop)
			checkErr(nil, err)
		},
	)

	internal.TestCase(t, "Status")(
		func(t *testing.T) {
			expErr := errors.New("status failed")
			mock.Status()(nil, expErr)

			_, err := client.Status()
			checkErr(expErr, err)
		},

		func(t *testing.T) {
			const expRes = "OK"
			mock.Status()(expRes, nil)
			res, err := client.Status()
			checkErr(nil, err)
			if res != expRes {
				t.Fatalf("gotResult != expResul")
			}
		},
	)

	internal.TestCase(t, "LLen")(
		func(t *testing.T) {
			expErr := errors.New("LLen failed")
			mock.LLen(key)(0, expErr)

			_, err := client.LLen(ctx, key)
			checkErr(expErr, err)
		},

		func(t *testing.T) {
			const expLen = 13

			mock.LLen(key)(expLen, nil)

			res, err := client.LLen(ctx, key)
			checkErr(nil, err)
			if res != expLen {
				t.Fatalf("expResult should be equal gotResul")
			}
		},
	)
}
