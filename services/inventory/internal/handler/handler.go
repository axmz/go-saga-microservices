package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/axmz/go-saga-microservices/inventory-service/internal/service"
	"github.com/axmz/go-saga-microservices/pkg/events"
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

func (h *InventoryHandler) HandleOrderCreatedEvent(ctx context.Context, event *events.OrderCreatedEvent) {
	h.Service.ReserveItems(ctx, event)
}

func (h *InventoryHandler) HandlePaymentFailedEvent(ctx context.Context, event *events.OrderCreatedEvent) {
}
