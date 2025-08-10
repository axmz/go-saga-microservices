package consumer

import (
	"context"
	"io"
	"log/slog"

	"errors"

	"github.com/axmz/go-saga-microservices/services/order/internal/handler"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	Reader  *kafka.Reader
	Handler *handler.Handler
}

func New(r *kafka.Reader, h *handler.Handler) *Consumer {
	return &Consumer{Reader: r, Handler: h}
}

func (c *Consumer) Start(ctx context.Context) error {
	slog.Info("Consumer started")
	for {
		m, err := c.Reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, kafka.ErrGroupClosed) ||
				errors.Is(err, io.EOF) ||
				ctx.Err() != nil {
				return err
			}
			slog.Warn("Kafka read error:", "err", err)
			continue
		}
		slog.Info("Kafka message", "topic", m.Topic, "partition", m.Partition, "offset", m.Offset)
		switch m.Topic {
		case "inventory.events":
			c.Handler.InventoryEvents(ctx, m)
		case "payment.events":
			c.Handler.PaymentEvents(ctx, m)
		default:
			slog.Warn("Unhandled event", "topic", m.Topic)
		}
	}
}
