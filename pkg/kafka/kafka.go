package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type IKafkaProducer interface {
	Produce(ctx context.Context, key string, value map[string]string) error
}

type Producer struct {
	writer *kafka.Writer
}

type Config struct {
	Broker string `mapstructure:"broker"`
	Topic  string `mapstructure:"topic"`
}

func NewKafkaProducer(cfg Config) IKafkaProducer {
	w := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Broker),
		Topic:    cfg.Topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: w,
	}
}

type Publisher interface {
	Produce(ctx context.Context, payload interface{}) error
}

func (kp *Producer) Produce(ctx context.Context, key string, value map[string]string) error {
	message, err := kp.encodeMessage(key, value)
	if err != nil {
		return err
	}
	return kp.writer.WriteMessages(ctx, message)
}

func (kp *Producer) encodeMessage(key string, value map[string]string) (kafka.Message, error) {
	v, err := json.Marshal(value)
	if err != nil {
		return kafka.Message{}, err
	}

	return kafka.Message{
		Key:   []byte(key),
		Value: v,
	}, nil
}
