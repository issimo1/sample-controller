package utils

import "context"
import "github.com/go-redis/redis/v8"

type RedisConfig struct {
	Addr     string
	Password string
}

func NewRedisClient(ctx context.Context, cfg RedisConfig) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Addr,
		Password: cfg.Password,
		DB:       0})
	c := context.Background()
	_, err := redisClient.Ping(c).Result()
	if err != nil {
		return nil, err
	}
	return redisClient, nil
}
