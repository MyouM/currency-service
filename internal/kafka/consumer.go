package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

//go:generate mockgen -source=consumer.go -destination=mock/mock_consumer.go -package=mock_kafka

type Consumer struct {
	reader *kafka.Reader
}

type ConsumerFuncs interface {
	Listen(context.Context) (kafka.Message, error)
	Close() error
}

func NewConsumer(broker, topic, groupID string) Consumer {
	return Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{broker},
			GroupID: groupID,
			Topic:   topic,
		}),
	}
}

func (c Consumer) Listen(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

func (c Consumer) Close() error {
	return c.reader.Close()
}
