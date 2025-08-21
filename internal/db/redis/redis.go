package redis

import (
	"context"
	"currency-service/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var rds_clnt *redis.Client

func InitRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	var ctx = context.Background()

	rds_clnt = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return rds_clnt, rds_clnt.Ping(ctx).Err()
}

func SetToken(token string) error {
	return rds_clnt.Set(context.Background(), token, "txt", 2*time.Minute).Err()
}

func FindToken(token string) error {
	_, err := rds_clnt.Get(context.Background(), token).Result()
	return err
}
