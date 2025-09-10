package kafka

import (
	"currency-service/internal/config"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	RegisterReqTopic  = "registration-request"
	RegisterRespTopic = "registration-response"
	LoginReqTopic     = "login-request"
	LoginRespTopic    = "login-response"
)

var topics = []string{
	RegisterReqTopic,
	RegisterRespTopic,
	LoginReqTopic,
	LoginRespTopic}

func InitKafkaTopics(cfg *config.KafkaConfig) error {
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": cfg.BrokerHost,
	})
	if err != nil {
		return err
	}
	defer admin.Close()

	tpcConfigs := make([]kafka.TopicSpecification, 0, 1)
	for _, topic := range topics {
		tpc := kafka.TopicSpecification{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
		tpcConfigs = append(tpcConfigs, tpc)
	}
	results, err := admin.CreateTopics(
		nil,
		tpcConfigs,
		kafka.SetAdminOperationTimeout(10*time.Second))
	if err != nil {
		return err
	}

	for _, res := range results {
		if res.Error.Code() != kafka.ErrNoError &&
			res.Error.Code() != kafka.ErrTopicAlreadyExists {
			return fmt.Errorf("failed to create topic %s: %v", res.Topic, res.Error)
		}
	}
	return nil
}
