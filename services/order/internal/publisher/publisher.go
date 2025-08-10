package publisher

import (
	"context"

	"github.com/axmz/go-saga-microservices/pkg/proto/events"
	"github.com/axmz/go-saga-microservices/services/order/internal/domain"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

type Publisher struct {
	Writer *kafka.Writer
}

func New(writer *kafka.Writer) *Publisher {
	return &Publisher{Writer: writer}
}

func (p *Publisher) PublishOrderCreated(ctx context.Context, order *domain.Order) error {
	items := make([]*events.Item, len(order.Items))
	for i, item := range order.Items {
		items[i] = &events.Item{
			Id: item.ProductID,
		}
	}
	event := events.OrderEventEnvelope{
		Event: &events.OrderEventEnvelope_OrderCreated{
			OrderCreated: &events.OrderCreatedEvent{
				Id:    order.ID,
				Items: items,
			},
		},
	}
	value, err := proto.Marshal(&event)
	if err != nil {
		return err
	}
	return p.Writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(order.ID),
		Value: value,
	})
}
