package mock

import (
	"time"

	"github.com/go-redis/redismock/v8"
)

type RedisClientMock struct {
	mock redismock.ClientMock
}

func newRedisClientMock(mock redismock.ClientMock) *RedisClientMock {
	return &RedisClientMock{
		mock: mock,
	}
}

func NewRedisClientMock(order bool) (*RedisClientDb, *RedisClientMock) {
	db, mock := redismock.NewClientMock()
	mock.MatchExpectationsInOrder(order)

	return newRedisClientDb(db), newRedisClientMock(mock)
}

type errorSetter interface {
	SetErr(err error)
	SetVal(val string)
}

func setErr(errSetter errorSetter) func(err error) {
	return func(err error) {
		if err != nil {
			errSetter.SetErr(err)
		} else {
			errSetter.SetVal("ok")
		}
	}
}

func (m *RedisClientMock) RPush(expListKey string, expVal ...interface{}) func(error) {
	res := m.mock.ExpectRPush(expListKey, expVal)
	return func(err error) {
		if err != nil {
			res.SetErr(err)
		} else {
			res.SetVal(1)
		}
	}
}

func (m *RedisClientMock) LTrim(listKey string, start, stop int64) func(error) {
	return setErr(m.mock.ExpectLTrim(listKey, start, stop))
}

func (m *RedisClientMock) LRange(listKey string, start, stop int64) func([]string, error) {
	res := m.mock.ExpectLRange(listKey, start, stop)

	return func(expVal []string, expErr error) {
		if expErr != nil {
			res.SetErr(expErr)
		} else if expVal == nil {
			res.RedisNil()
		} else {
			res.SetVal(expVal)
		}
	}
}

func (m *RedisClientMock) Set(key string, value interface{}, expiration time.Duration) func(error) {
	return setErr(m.mock.ExpectSet(key, value, expiration))
}

func (m *RedisClientMock) Get(key string) func(string, error) {
	res := m.mock.ExpectGet(key)

	return func(expRes string, expErr error) {
		if expErr != nil {
			res.SetErr(expErr)
		} else if expRes == "" {
			res.RedisNil()
		} else {
			res.SetVal(expRes)
		}
	}
}

func (m *RedisClientMock) Status() func(interface{}, error) {
	res := m.mock.ExpectPing()

	return func(expRes interface{}, expErr error) {
		if expErr != nil {
			res.SetErr(expErr)
		} else {
			res.SetVal("Ok")
		}
	}
}

func (m *RedisClientMock) LLen(listKey string) func(int64, error) {
	res := m.mock.ExpectLLen(listKey)

	return func(i int64, err error) {
		if err != nil {
			res.SetErr(err)
		} else {
			res.SetVal(i)
		}
	}
}

func (m *RedisClientMock) Done() error {
	return m.mock.ExpectationsWereMet()
}
