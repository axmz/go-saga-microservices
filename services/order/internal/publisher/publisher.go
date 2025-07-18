package publisher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	Writer *kafka.Writer
}

func New(writer *kafka.Writer) *Publisher {
	return &Publisher{Writer: writer}
}

type OrderCreatedEvent struct {
	OrderID string      `json:"orderId"`
	Items   interface{} `json:"items"`
	Status  string      `json:"status"`
	Time    time.Time   `json:"timestamp"`
}

func (p *Publisher) PublishOrderCreated(ctx context.Context, orderID string, items interface{}, status string) error {
	event := OrderCreatedEvent{
		OrderID: orderID,
		Items:   items,
		Status:  status,
		Time:    time.Now(),
	}
	value, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.Writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(orderID),
		Value: value,
	})
}
