package infrastructure

import (
	"context"
	"discord-playlist-notifier/src/interfaces"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisHandler struct {
	Redis *redis.Client
	ctx   context.Context
}

func NewRedisHandler(ctx context.Context) (interfaces.RedisHandler, error) {
	Redis := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	err := Redis.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	return RedisHandler{Redis, ctx}, nil
}

func (h RedisHandler) Get(key string) interfaces.StringCmd {
	return h.Redis.Get(h.ctx, key)
}

func (h RedisHandler) Set(key string, value interface{}, expiration time.Duration) interfaces.StatusCmd {
	return h.Redis.Set(h.ctx, key, value, expiration)
}
