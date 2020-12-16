package services

import (
	"context"
	"github.com/go-redis/redis/v8"
	"os"
)

type RedisService interface {
	Set(string, string) error
	Get(string) (string, error)
	Delete(string) error
}

type redisService struct {
	client *redis.Client
}

func NewRedisService() RedisService {
	redisURL := os.Getenv("REDIS_URL")
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return &redisService{
		client: rdb,
	}
}

func (redisClient *redisService) Set(key, message string) error {
	ctx := context.Background()
	err := redisClient.client.Set(ctx, key, message, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (redisClient *redisService) Get(key string) (string, error) {

	ctx := context.Background()
	val, err := redisClient.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (redisClient *redisService) Delete(key string) error {

	ctx := context.Background()
	err := redisClient.client.Del(ctx, key).Err()
	return err

}
