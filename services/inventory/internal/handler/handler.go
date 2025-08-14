package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/axmz/go-saga-microservices/inventory-service/internal/domain"
	"github.com/axmz/go-saga-microservices/inventory-service/internal/service"
	httputils "github.com/axmz/go-saga-microservices/lib/adapter/http"
	"github.com/axmz/go-saga-microservices/pkg/proto/events"
	httppb "github.com/axmz/go-saga-microservices/pkg/proto/http"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
)

type Handler struct {
	Service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{
		Service: service,
	}
}

func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	slog.Info("Inventory.GetProducts request", "method", r.Method, "path", r.URL.Path, "remote", r.RemoteAddr)
	products, err := h.Service.GetProducts(r.Context())
	if err != nil {
		slog.Error("Inventory.GetProducts service error", "err", err)
		httputils.ErrorInternal(w, err)
		return
	}

	protoProducts := h.toProtoProducts(products)

	slog.Info("Inventory.GetProducts success", "count", len(protoProducts))
	h.respondWithGetProductsResponse(w, protoProducts)
}

func (h *Handler) ResetAllProducts(w http.ResponseWriter, r *http.Request) {
	if err := h.Service.ResetAllProducts(r.Context()); err != nil {
		slog.Error("Inventory.ResetAllProducts service error", "err", err)
		httputils.ErrorInternal(w, err)
		return
	}

	slog.Info("Inventory.ResetAllProducts success")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) OrderEvents(ctx context.Context, event kafka.Message) {
	data, ok := toProtoPayload(event.Value)
	if !ok {
		slog.Warn("OrderEvents: JSON outbox parse failed; dropping message")
		return
	}

	var envelope events.OrderEventEnvelope
	if err := proto.Unmarshal(data, &envelope); err != nil {
		slog.Warn("OrderEvents: proto unmarshal failed", "err", err)
		return
	}
	switch evt := envelope.Event.(type) {
	case *events.OrderEventEnvelope_OrderCreated:
		slog.Info("Order event: created", "topic", event.Topic, "partition", event.Partition, "offset", event.Offset, "orderId", evt.OrderCreated.Id)
		h.Service.ReserveItems(ctx, evt.OrderCreated)
	default:
		slog.Warn("OrderEvents: unknown or missing event type")
	}
}

func (h *Handler) PaymentEvents(ctx context.Context, message kafka.Message) {
	var envelope events.PaymentEventEnvelope
	if err := proto.Unmarshal(message.Value, &envelope); err != nil {
		slog.Warn("Failed to unmarshal PaymentEventEnvelope", "err", err)
		return
	}

	switch evt := envelope.Event.(type) {
	case *events.PaymentEventEnvelope_PaymentSucceeded:
		slog.Info("Payment event: succeeded", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "orderId", evt.PaymentSucceeded.Id)
		h.Service.MarkItemsSold(ctx, evt.PaymentSucceeded.Id)
	case *events.PaymentEventEnvelope_PaymentFailed:
		slog.Info("Payment event: failed", "topic", message.Topic, "partition", message.Partition, "offset", message.Offset, "orderId", evt.PaymentFailed.Id)
		h.Service.ReleaseReservedItems(ctx, evt.PaymentFailed.Id)
	default:
		slog.Warn("Unknown or missing event type in envelope")
	}
}

// RESPONSES
func (h *Handler) respondWithGetProductsResponse(w http.ResponseWriter, products []*httppb.Product) {
	resp := &httppb.GetProductsResponse{Products: products}
	httputils.RespondProto(w, resp, http.StatusOK)
}

// MAPPERS
func (h *Handler) toProtoProducts(products []domain.Product) []*httppb.Product {
	out := make([]*httppb.Product, len(products))
	for i, p := range products {
		out[i] = &httppb.Product{
			Id:     int64(p.ID),
			Name:   p.Name,
			Price:  p.Price,
			Sku:    p.SKU,
			Status: p.Status,
		}
	}
	return out
}

func toProtoPayload(value []byte) ([]byte, bool) {
	var outbox struct {
		ID            string          `json:"id"`
		AggregateType string          `json:"aggregate_type"`
		AggregateID   string          `json:"aggregate_id"`
		EventType     string          `json:"event_type"`
		Payload       string          `json:"payload"`
		Headers       json.RawMessage `json:"headers"`
		CreatedAt     string          `json:"created_at"`
		PublishedAt   *string         `json:"published_at"`
	}
	if err := json.Unmarshal(value, &outbox); err != nil {
		return nil, false
	}
	b, err := base64.StdEncoding.DecodeString(outbox.Payload)
	if err != nil {
		return nil, false
	}
	return b, true
}
