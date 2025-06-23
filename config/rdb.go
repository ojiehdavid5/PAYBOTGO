package config

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis() (*redis.Client, error) {
	redisPassword := os.Getenv("REDIS_PASSWORD") // Get Redis password from environment variable
	redisAddress := os.Getenv("REDIS_ADDRESS")   // Get Redis address from environment variable

	ctx := context.Background()

	// Create a new Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddress,  // Redis server address
		Password: redisPassword, // Password set
		DB:       0,             // Use default DB
	})

	// Ping the Redis server to check if it's available
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	fmt.Println("Connected to Redis")
	return rdb, nil
}
