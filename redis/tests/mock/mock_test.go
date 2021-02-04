package mock

import (
	"context"
	"errors"
	"testing"
)

func testCase(t *testing.T, name string) func(failed, success func(t *testing.T)) {
	return func(failed, success func(t *testing.T)) {
		t.Run(name, func(t *testing.T) {
			t.Run("failed", failed)
			t.Run("success", success)
		})
	}
}

func TestRedisClientMock(t *testing.T) {
	const (
		key   = "testKey"
		val   = "value"
		start = 1
		stop  = 20
	)

	checkErr := func(expErr, gotErr error) {
		if gotErr == nil && expErr == nil {
			return
		}

		if gotErr == nil && expErr != nil {
			t.Fatal("the error received must not be nil")
		}

		if gotErr != nil && expErr == nil {
			t.Fatalf("gotErr := '%s'; the error received must be nil", errors.Unwrap(gotErr))
		}

		if !errors.Is(gotErr, expErr) {
			t.Fatalf("expErr := '%s' ; gotErr := '%s'; gotErr should must be equal expErr",
				expErr.Error(), errors.Unwrap(gotErr))
		}
	}

	ctx := context.TODO()
	client, mock := NewRedisClientMock(true)

	testCase(t, "Set")(
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

	testCase(t, "Get")(
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

	testCase(t, "RPush")(
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

	testCase(t, "LRange")(
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

	testCase(t, "LTrim")(
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

	testCase(t, "Status")(
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
		})
}
