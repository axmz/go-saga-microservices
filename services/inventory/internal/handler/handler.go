package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/axmz/go-saga-microservices/inventory-service/internal/service"
	"github.com/axmz/go-saga-microservices/pkg/events"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

type InventoryHandler struct {
	Service *service.Service
}

func New(service *service.Service) *InventoryHandler {
	return &InventoryHandler{
		Service: service,
	}
}

func (h *InventoryHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.Service.GetProducts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *InventoryHandler) OrderEvents(ctx context.Context, event kafka.Message) {
	var envelope events.OrderEventEnvelope

	err := proto.Unmarshal(event.Value, &envelope)
	if err != nil {
		log.Printf("failed to unmarshal OrderEventEnvelope: %v", err)
		return
	}

	switch evt := envelope.Event.(type) {
	case *events.OrderEventEnvelope_OrderCreated:
		h.Service.ReserveItems(ctx, evt.OrderCreated)
	default:
		log.Printf("Unknown or missing event type in envelope")
	}
}

func (h *InventoryHandler) PaymentEvents(ctx context.Context, message kafka.Message) {
	var envelope events.PaymentEventEnvelope
	if err := proto.Unmarshal(message.Value, &envelope); err != nil {
		log.Printf("failed to unmarshal PaymentEventEnvelope: %v", err)
		return
	}

	switch evt := envelope.Event.(type) {
	case *events.PaymentEventEnvelope_PaymentSucceeded:
		h.Service.MarkItemsSold(ctx, evt.PaymentSucceeded.Id)
	case *events.PaymentEventEnvelope_PaymentFailed:
		h.Service.ReleaseReservedItems(ctx, evt.PaymentFailed.Id)
	default:
		log.Printf("Unknown or missing event type in envelope")
	}
}
