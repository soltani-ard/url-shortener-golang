package database

import (
	"context"
	"github.com/go-redis/redis/v8"
	"os"
)

var Ctx = context.Background()

// CreateClient : create redis client
func CreateClient(dbNO int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("DB_ADDR"),
		Password: os.Getenv("DB_PASS"),
		DB:       dbNO,
	})

	return rdb
}
