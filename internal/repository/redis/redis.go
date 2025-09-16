package redis

import (
	"context"
	"currency-service/internal/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var rds_clnt *redis.Client

type RedisClient struct {
	Client *redis.Client
}

type RedisFuncs interface {
	FindToken(string) error
	SetToken(string) error
	Ping() error
}

func InitRedis(cfg *config.RedisConfig) (RedisClient, error) {
	var ctx = context.Background()

	rds_clnt = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return RedisClient{Client: rds_clnt}, rds_clnt.Ping(ctx).Err()
}

func (rds RedisClient) SetToken(token string) error {
	return rds.Client.Set(context.Background(), token, "txt", 2*time.Minute).Err()
}

func (rds RedisClient) FindToken(token string) error {
	_, err := rds.Client.Get(context.Background(), token).Result()
	return err
}

func (rds RedisClient) Ping() error {
	return rds.Client.Ping(context.Background()).Err()
}

func GetRedisClient() RedisClient {
	return RedisClient{Client: rds_clnt}
}
