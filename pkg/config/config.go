package config

import (
	"context"
	"sync"

	"github.com/go-redis/redis/v8"
)

var (
	RedisClient *redis.Client
	once        sync.Once
)

func init() {
	RedisClient = GetRedisClient()
}

func GetRedisClient() *redis.Client {
	once.Do(func() {
		RedisClient = redis.NewClient(&redis.Options{
			Addr: "localhost:6379",
			// Other configurations like password, DB, etc.
		})
	})
	return RedisClient
}
func GetContext() context.Context {
	return context.Background()
}
