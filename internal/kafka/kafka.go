package kafka

import (
	"currency-service/internal/config"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	GroupID          = "auth-gateway-work"
	AuthGatewayTopic = "auth-gateway"
	GatewayAuthTopic = "gateway-Aauth"
)

var topics = []string{AuthGatewayTopic, GatewayAuthTopic}

type AuthRequest struct {
	Type     string `json:"type"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Type  string `json:"type"`
	Token string `json:"token"`
	Error string `json:"error"`
}

func InitKafkaTopics(cfg *config.KafkaConfig) error {
	ctrlConn, err := kafka.Dial("tcp", cfg.ControllerHost)
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
