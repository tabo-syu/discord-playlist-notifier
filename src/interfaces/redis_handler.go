package interfaces

import (
	"time"
)

type RedisHandler interface {
	Get(key string) StringCmd
	Set(key string, value interface{}, expiration time.Duration) StatusCmd
}

type StringCmd interface {
	Result() (string, error)
}

type StatusCmd interface {
	Err() error
}
