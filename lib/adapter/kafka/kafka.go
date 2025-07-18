package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	Addr          string
	ProducerTopic string
	ConsumerTopic string
	GroupID       string
}

type Broker struct {
	Writer *kafka.Writer
	Reader *kafka.Reader
}

func Init(cfg Config) (*Broker, error) {
	Addr := cfg.Addr
	producerTopic := cfg.ProducerTopic
	consumerTopic := cfg.ConsumerTopic
	groupID := cfg.GroupID

	writer := &kafka.Writer{
		Addr:     kafka.TCP(Addr),
		Topic:    producerTopic,
		Balancer: &kafka.LeastBytes{},
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{Addr},
		Topic:   consumerTopic,
		GroupID: groupID,
	})

	return &Broker{
		Writer: writer,
		Reader: reader,
	}, nil
}

func (kc *Broker) ListenAndHandle(ctx context.Context, handler func(kafka.Message)) error {
	for {
		m, err := kc.Reader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		handler(m)
	}
}

func (kc *Broker) Shutdown(ctx context.Context) error {
	var errWriter, errReader error
	if kc.Writer != nil {
		errWriter = kc.Writer.Close()
	}
	if kc.Reader != nil {
		errReader = kc.Reader.Close()
	}
	if errWriter != nil {
		return errWriter
	}
	return errReader
}
