package consumer

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"time"

	"errors"

	"github.com/axmz/go-saga-microservices/services/order/internal/domain"
	"github.com/axmz/go-saga-microservices/services/order/internal/repository"
	"github.com/axmz/go-saga-microservices/services/order/internal/service"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	Reader *kafka.Reader
	Repo   *repository.Repository
}

func New(reader *kafka.Reader, svc *service.Service) *Consumer {
	return &Consumer{Reader: reader, Repo: svc.Repo}
}

func (c *Consumer) Start(ctx context.Context) error {
	slog.Info("Consumer started")
	for {
		time.Sleep(time.Millisecond * 1500)
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
		var evt map[string]interface{}
		if err := json.Unmarshal(m.Value, &evt); err != nil {
			log.Printf("Failed to unmarshal event: %v", err)
			continue
		}
		eventType, _ := evt["event_type"].(string)
		orderID, _ := evt["order_id"].(string)
		switch eventType {
		case "inventory.reserved":
			status, _ := evt["status"].(string)
			c.updateOrderStatus(ctx, orderID, status)
		case "payments.success":
			c.updateOrderStatus(ctx, orderID, domain.StatusPaid)
		case "payments.failed":
			c.updateOrderStatus(ctx, orderID, domain.StatusFailed)
		}
	}
}

func (c *Consumer) updateOrderStatus(ctx context.Context, orderID, status string) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	ord, err := c.Repo.GetOrder(ctx, orderID)
	if err != nil || ord == nil {
		log.Printf("Order not found or error: %v", err)
		return
	}
	ord.Status = status
	ord.UpdatedAt = time.Now()
	_ = c.Repo.UpdateOrder(ctx, ord)
}
