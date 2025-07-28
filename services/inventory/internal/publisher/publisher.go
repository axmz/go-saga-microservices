package publisher

import (
	"context"
	"encoding/json"
	"log"

	"github.com/axmz/go-saga-microservices/pkg/events"
	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	Writer *kafka.Writer
}

func New(writer *kafka.Writer) *Publisher {
	return &Publisher{Writer: writer}
}

func (k *Publisher) PublishInventoryReservationSucceededEvent(orderID string) {
	log.Printf("[InventoryService] Publishing inventory reservation event for order: %s, status: %s", orderID, "success")

	event := &events.InventoryReservationSucceeded{
		Id: orderID,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling inventory reservation event: %v", err)
		return
	}
	err = k.Writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(orderID),
		Value: eventJSON,
	})
	if err != nil {
		log.Printf("Error publishing inventory reservation event: %v", err)
	}
}

func (k *Publisher) PublishInventoryReservationFailedEvent(orderID string) {
	log.Printf("[InventoryService] Publishing inventory reservation event for order: %s, status: %s", orderID, "failed")

	event := &events.InventoryReservationFailed{
		Id: orderID,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling inventory reservation event: %v", err)
		return
	}
	err = k.Writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(orderID),
		Value: eventJSON,
	})
	if err != nil {
		log.Printf("Error publishing inventory reservation event: %v", err)
	}
}
