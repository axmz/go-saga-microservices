package publisher

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/axmz/go-saga-microservices/pkg/events"
	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	Writer *kafka.Writer
}

func New(writer *kafka.Writer) *Publisher {
	return &Publisher{Writer: writer}
}

type InventoryEvent struct {
	EventType string    `json:"event_type"`
	OrderID   string    `json:"order_id"`
	ProductID string    `json:"product_id"`
	SKU       string    `json:"sku"`
	Quantity  int       `json:"quantity"`
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func (k *Publisher) PublishInventoryEvent(eventType, orderID, productID string, quantity int, success bool, message string) {
	event := InventoryEvent{
		EventType: eventType,
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		Success:   success,
		Message:   message,
		Timestamp: time.Now(),
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return
	}

	err = k.Writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(orderID),
		Value: eventJSON,
	})
	if err != nil {
		log.Printf("Error publishing event: %v", err)
	}
}

func (k *Publisher) PublishReserveProductsEvent(orderID string, status events.EventStatus) {
	log.Printf("[InventoryService] Publishing reserveProductsEvent for order: %s, status: %s", orderID, status)
	event := map[string]interface{}{
		"orderId": orderID,
		"status":  status,
	}
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling reserveProductsEvent: %v", err)
		return
	}
	err = k.Writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(orderID),
		Value: eventJSON,
	})
	if err != nil {
		log.Printf("Error publishing reserveProductsEvent: %v", err)
	}
}
