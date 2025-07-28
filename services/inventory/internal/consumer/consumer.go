package consumer

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"

	"errors"

	"github.com/axmz/go-saga-microservices/inventory-service/internal/handler"
	"github.com/axmz/go-saga-microservices/pkg/events"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	Reader  *kafka.Reader
	Handler *handler.InventoryHandler
}

func New(reader *kafka.Reader, handler *handler.InventoryHandler) *Consumer {
	return &Consumer{Reader: reader, Handler: handler}
}

func (c *Consumer) Start(ctx context.Context) error {
	slog.Info("Consumer started")
	for {
		message, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, kafka.ErrGroupClosed) ||
				errors.Is(err, io.EOF) ||
				ctx.Err() != nil {
				return err
			}
			slog.Warn("Kafka read error:", "err", err)
			continue
		}
		var event *events.OrderCreatedEvent
		if err := json.Unmarshal(message.Value, event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
		}
		switch event.Type {
		case "orderCreatedEvent":
			c.Handler.HandleOrderCreatedEvent(ctx, event)
		case "paymentFailedEvent":
			c.Handler.HandlePaymentFailedEvent(ctx, event)
		}
	}
}
