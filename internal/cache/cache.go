package cache

import (
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func NewCacheFromEnvironment() (*redis.Client, error) {
	host, err := getEnv("CACHE_HOST")
	if err != nil {
		return nil, err
	}

	port, err := getEnv("CACHE_PORT")
	if err != nil {
		return nil, err
	}

	password, err := getEnv("CACHE_PASSWORD")
	if err != nil {
		return nil, err
	}

	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0,
	}), nil
}

func getEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("not found env variable by %s name", key)
	}

	return value, nil
}
