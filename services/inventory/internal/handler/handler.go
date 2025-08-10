package handler

import (
	"context"
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
		h.httpInternalServerError(w, err)
		return
	}

	protoProducts := h.toProtoProducts(products)

	slog.Info("Inventory.GetProducts success", "count", len(protoProducts))
	h.respondWithGetProductsResponse(w, protoProducts)
}

func (h *Handler) OrderEvents(ctx context.Context, event kafka.Message) {
	var envelope events.OrderEventEnvelope

	err := proto.Unmarshal(event.Value, &envelope)
	if err != nil {
		slog.Warn("Failed to unmarshal OrderEventEnvelope", "err", err)
		return
	}

	switch evt := envelope.Event.(type) {
	case *events.OrderEventEnvelope_OrderCreated:
		slog.Info("Order event: created", "topic", event.Topic, "partition", event.Partition, "offset", event.Offset, "orderId", evt.OrderCreated.Id)
		h.Service.ReserveItems(ctx, evt.OrderCreated)
	default:
		slog.Warn("Unknown or missing event type in envelope")
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

// ERROR HELPERS
func (h *Handler) httpInternalServerError(w http.ResponseWriter, err error) {
	slog.Error("HTTP 500 Internal server error", "err", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
