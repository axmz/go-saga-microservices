package publisher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/axmz/go-saga-microservices/services/order/internal/domain"
	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	Writer *kafka.Writer
}

func New(writer *kafka.Writer) *Publisher {
	return &Publisher{Writer: writer}
}

type OrderCreatedEvent struct {
	OrderID string    `json:"orderId"`
	Time    time.Time `json:"timestamp"`
}

func (p *Publisher) PublishOrderCreated(ctx context.Context, order *domain.Order) error {
	event := OrderCreatedEvent{
		OrderID: order.ID,
		Time:    time.Now(),
	}
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.Writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(order.ID),
		Value: value,
	})
}
