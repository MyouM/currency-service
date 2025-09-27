package test

import (
	"context"
	"currency-service/internal/auth"
	"currency-service/internal/config"
	kafkaCur "currency-service/internal/kafka"
	mock_postgres "currency-service/internal/repository/postgres/mock"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestAuthService(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	networkName := "kafka-test-net"

	nw, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name: networkName,
		},
	})
	require.NoError(t, err)
	defer nw.Remove(ctx)

	zkReq := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-zookeeper:7.5.0",
		ExposedPorts: []string{"2181/tcp"},
		Env: map[string]string{
			"ZOOKEEPER_CLIENT_PORT": "2181",
			"ZOOKEEPER_TICK_TIME":   "2000",
		},
		WaitingFor: wait.ForListeningPort("2181/tcp"),
		Networks:   []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"zookeeper"},
		},
	}

	zkC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: zkReq,
		Started:          true,
	})
	require.NoError(t, err)
	defer zkC.Terminate(ctx)

	nets, err := zkC.Networks(ctx)
	require.NoError(t, err)
	require.Greater(t, len(nets), 0)

	kafkaReq := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-kafka:7.5.0",
		ExposedPorts: []string{"9093:9093/tcp"},
		Env: map[string]string{
			"KAFKA_BROKER_ID":                        "1",
			"KAFKA_ZOOKEEPER_CONNECT":                "zookeeper:2181",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":   "PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
			"KAFKA_LISTENERS":                        "PLAINTEXT://:9092,PLAINTEXT_HOST://:9093",
			"KAFKA_ADVERTISED_LISTENERS":             "PLAINTEXT://kafka:9092,PLAINTEXT_HOST://localhost:9093",
			"KAFKA_INTER_BROKER_LISTENER_NAME":       "PLAINTEXT",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE":        "true",
		},
		WaitingFor: wait.ForLog("started (kafka.server.KafkaServer)").WithStartupTimeout(2 * time.Minute),
		Networks:   []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"kafka"},
		},
	}

	kafkaC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: kafkaReq,
		Started:          true,
	})
	require.NoError(t, err)
	defer kafkaC.Terminate(ctx)

	controllerAddr := "localhost:9093"

	kafkaConfig := &config.KafkaConfig{ControllerHost: controllerAddr}
	err = kafkaCur.InitKafkaTopics(kafkaConfig)
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPsql := mock_postgres.NewMockAuthPsqlFuncs(ctrl)
	mockPsql.EXPECT().LogIn("login", "1234").Return(true, nil)
	consumer := kafkaCur.NewConsumer(
		kafkaConfig.ControllerHost,
		kafkaCur.AuthGatewayTopic,
		kafkaCur.GroupID)
	producer := kafkaCur.NewProducer(
		kafkaConfig.ControllerHost,
		kafkaCur.GatewayAuthTopic)
	defer consumer.Close()
	defer producer.Close()

	go auth.StartAuthService(ctx, kafkaConfig, mockPsql)

	valueProd := `{"type":"login","login":"login","password":"1234"}`
	err = producer.Send(ctx, []byte("login"), []byte(valueProd))
	assert.NoError(t, err)

	msg, err := consumer.Listen(ctx)
	assert.NoError(t, err)

	resp := kafkaCur.AuthResponse{}

	err = json.Unmarshal(msg.Value, &resp)
	assert.NoError(t, err)
	assert.Equal(t, "", resp.Error)
}
