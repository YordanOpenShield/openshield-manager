package redis

import (
	"github.com/go-redis/redis/v8"
)

// Initialize Redis client
func Init() *redis.Client {
	options := &redis.Options{
		Addr:     "localhost:6379", // Change this if using a cloud Redis instance
		Password: "",               // No password set
		DB:       0,                // Default DB
	}

	client := redis.NewClient(options)
	return client
}
