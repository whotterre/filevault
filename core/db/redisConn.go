package db

import "github.com/redis/go-redis/v9"

// GetRedisClient initializes and returns a Redis client.
// It connects to a Redis server running on localhost at port 6379.
func GetRedisClient() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "172.17.0.3:6379", // TODO: Load this from config file 
		Password: "", 
		DB:       0, 
	})
	
	return client, nil 
}