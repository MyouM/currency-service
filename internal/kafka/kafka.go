package kafka

import (
	"currency-service/internal/config"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
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
	conn, err := kafka.Dial("tcp", cfg.BrokerHost)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	ctrlAddr := fmt.Sprintf("%s:%d", controller.Host, controller.Port)

	ctrlConn, err := kafka.Dial("tcp", ctrlAddr)
	if err != nil {
		return err
	}
	defer ctrlConn.Close()

	_ = ctrlConn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	tpcConfigs := make([]kafka.TopicConfig, 0, 1)
	for _, topic := range topics {
		tpc := kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
		tpcConfigs = append(tpcConfigs, tpc)
	}
	if err := ctrlConn.CreateTopics(tpcConfigs...); err != nil {
		return err
	}
	return nil
}
