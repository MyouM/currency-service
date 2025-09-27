package test

import (
	"context"
	redisCur "currency-service/internal/repository/redis"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRedisFuncs(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(5 * time.Second),
	}

	redisC, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
	assert.NoError(t, err)
	defer redisC.Terminate(ctx)

	host, _ := redisC.Host(ctx)
	port, _ := redisC.MappedPort(ctx, "6379")

	rdb := redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
	})
	client := redisCur.RedisClient{Client: rdb}

	err = client.SetToken("1234")
	assert.NoError(t, err)

	err = client.FindToken("1234")
	assert.NoError(t, err)
}
