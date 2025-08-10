package publisher

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/axmz/go-saga-microservices/pkg/proto/events"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

type Publisher struct {
	Writer *kafka.Writer
}

func New(writer *kafka.Writer) *Publisher {
	return &Publisher{Writer: writer}
}

func (k *Publisher) PublishInventoryReservationSucceededEvent(orderID string) {
	slog.Info("[InventoryService] Publishing inventory reservation success event for order: %s, status: %s", orderID, "success")

	event := &events.InventoryEventEnvelope{
		Event: &events.InventoryEventEnvelope_ReservationSucceeded{
			ReservationSucceeded: &events.InventoryReservationSucceeded{
				Id: orderID,
			},
		},
	}

	value, err := proto.Marshal(event)
	if err != nil {
		slog.Warn("Failed to marshal InventoryReservationSucceeded event: ", "err", err)
	}

	err = k.Writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(orderID),
		Value: value,
	})

	if err != nil {
		slog.Error("Error publishing inventory reservation success event: ", "err", err)
	}
}

func (k *Publisher) PublishInventoryReservationFailedEvent(orderID string) {
	slog.Info("[InventoryService] Publishing inventory reservation failed event for order: %s, status: %s", orderID, "failed")

	event := &events.InventoryReservationFailed{
		Id: orderID,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		slog.Error("Error marshaling inventory reservation event: ", "err", err)
		return
	}
	err = k.Writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(orderID),
		Value: eventJSON,
	})
	if err != nil {
		slog.Error("Error publishing inventory reservation failed event: %v", err)
	}
}
