package publisher

import (
	"context"
	"log"

	"github.com/axmz/go-saga-microservices/pkg/events"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

type Publisher struct {
	Writer *kafka.Writer
}

func New(writer *kafka.Writer) *Publisher {
	return &Publisher{Writer: writer}
}

func (k *Publisher) PublishPaymentSucceededEvent(orderID string) error {
	log.Printf("[Payment Service] Publishing payment success event for order: %s, status: %s", orderID, "success")

	event := &events.PaymentEventEnvelope{
		Event: &events.PaymentEventEnvelope_PaymentSucceeded{
			PaymentSucceeded: &events.PaymentSucceeded{
				Id: orderID,
			},
		},
	}

	value, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	return k.Writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(orderID),
		Value: value,
	})
}

func (k *Publisher) PublishPaymentFailedEvent(orderID string) error {
	log.Printf("[Payment Service] Publishing payment failed event for order: %s, status: %s", orderID, "failed")

	event := &events.PaymentEventEnvelope{
		Event: &events.PaymentEventEnvelope_PaymentFailed{
			PaymentFailed: &events.PaymentFailed{
				Id: orderID,
			},
		},
	}

	value, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	return k.Writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(orderID),
		Value: value,
	})
}
