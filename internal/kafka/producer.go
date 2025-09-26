package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

//go:generate mockgen -source=producer.go -destination=mock/mock_producer.go -package=mock_kafka

type Producer struct {
	writer *kafka.Writer
}

type ProducerFuncs interface {
	Send(context.Context, []byte, []byte) error
	Close() error
}

func NewProducer(broker, topic string) Producer {
	return Producer{
		writer: kafka.NewWriter(kafka.WriterConfig{
			Brokers:  []string{broker},
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}),
	}
}

func (p Producer) Send(ctx context.Context, key, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}

func (p Producer) Close() error {
	return p.writer.Close()
}
