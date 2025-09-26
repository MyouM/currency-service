package test

import (
	"context"
	"currency-service/internal/auth"
	"currency-service/internal/config"
	kafkaCur "currency-service/internal/kafka"
	mock_postgres "currency-service/internal/repository/postgres/mock"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestAuthService(t *testing.T) {
	ctx := context.Background()
	zkReq := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-zookeeper:7.5.0",
		ExposedPorts: []string{"2181/tcp"},
		Env: map[string]string{
			"ZOOKEEPER_CLIENT_PORT": "2181",
			"ZOOKEEPER_TICK_TIME":   "2000",
		},
		WaitingFor: wait.ForListeningPort("2181/tcp"),
	}
	zkC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: zkReq,
		Started:          true,
	})
	assert.NoError(t, err)
	zkHost, _ := zkC.Host(ctx)
	zkPort, _ := zkC.MappedPort(ctx, "2181")
	zkAddr := fmt.Sprintf("%s:%s", zkHost, zkPort.Port())

	kafkaReq := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-kafka:7.5.0",
		ExposedPorts: []string{"9092/tcp", "9093/tcp", "9999/tcp"},
		Env: map[string]string{
			"KAFKA_BROKER_ID":                        "1",
			"KAFKA_ZOOKEEPER_CONNECT":                zkAddr,
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":   "PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT",
			"KAFKA_LISTENERS":                        "PLAINTEXT://:9092,PLAINTEXT_HOST://:9093",
			"KAFKA_ADVERTISED_LISTENERS":             "PLAINTEXT://kafka:9092,PLAINTEXT_HOST://localhost:9093",
			"KAFKA_INTER_BROKER_LISTENER_NAME":       "PLAINTEXT",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
			"KAFKA_AUTO_CREATE_TOPICS_ENABLE":        "true",
			"KAFKA_JMX_PORT":                         "9999",
			"KAFKA_JMX_HOSTNAME":                     "kafka",
		},
		WaitingFor: wait.ForListeningPort("9092/tcp").WithStartupTimeout(2 * time.Minute),
	}

	kafkaC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: kafkaReq,
		Started:          true,
	})
	assert.NoError(t, err)

	kafkaHost, _ := kafkaC.Host(ctx)
	kafkaPort, _ := kafkaC.MappedPort(ctx, "9092")
	brokerAddr := fmt.Sprintf("%s:%s", kafkaHost, kafkaPort.Port())

	kafkaConfig := &config.KafkaConfig{BrokerHost: brokerAddr}
	err = kafkaCur.InitKafkaTopics(kafkaConfig)
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPsql := mock_postgres.NewMockAuthPsqlFuncs(ctrl)
	mockPsql.EXPECT().LogIn("login", "1234").Return(true, nil)
	consumer := kafkaCur.NewConsumer(
		kafkaConfig.BrokerHost,
		kafkaCur.AuthGatewayTopic,
		kafkaCur.GroupID)
	producer := kafkaCur.NewProducer(
		kafkaConfig.BrokerHost,
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
