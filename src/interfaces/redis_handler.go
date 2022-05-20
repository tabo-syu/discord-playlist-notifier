package interfaces

import "time"

type RedisHandler interface {
	Get(key string) StringCmd
	Set(key string, value interface{}, expiration time.Duration) StatusCmd
	Del(key string) IntCmd
	Exists(key string) IntCmd
}

type StringCmd interface {
	Result() (string, error)
	Bytes() ([]byte, error)
}

type StatusCmd interface {
	Err() error
}

type IntCmd interface {
	Err() error
	Result() (int64, error)
}
