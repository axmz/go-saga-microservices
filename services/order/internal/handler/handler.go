package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	httputils "github.com/axmz/go-saga-microservices/lib/adapter/http"
	"github.com/axmz/go-saga-microservices/pkg/proto/events"
	httppb "github.com/axmz/go-saga-microservices/pkg/proto/http"
	"github.com/axmz/go-saga-microservices/services/order/internal/domain"
	"github.com/axmz/go-saga-microservices/services/order/internal/service"
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

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	domainItems, err := h.processCreateOrderRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := h.Service.CreateOrder(r.Context(), domainItems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.respondWithCreateOrderSuccess(w, order)
}

func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("orderID")
	if orderID == "" {
		httputils.ErrorBadRequest(w, errors.New("missing orderId"))
		return
	}

	ord, err := h.Service.GetOrder(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			slog.Error("failed to get order", "err", err)
		}
		return
	}

	if ord != nil {
		slog.Info("Fetched order", "id", ord.ID, "status", ord.Status)
	}

	h.respondWithGetOrderSuccess(w, ord)
}

func (h *Handler) OrderStatusWS(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("orderId")
	if orderID == "" {
		httputils.ErrorBadRequest(w, errors.New("missing orderId"))
		return
	}
	if r.Header.Get("Connection") != "Upgrade" || r.Header.Get("Upgrade") != "websocket" {
		http.Error(w, "Not a websocket upgrade request", http.StatusBadRequest)
		return
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, "Failed to hijack connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	// Write WebSocket handshake response
	bufrw.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\n\r\n")
	bufrw.Flush()
	for {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		ord, err := h.Service.GetOrder(ctx, orderID)
		cancel()
		if err != nil {
			bufrw.WriteString("error: " + err.Error() + "\n")
			bufrw.Flush()
			return
		}
		if ord == nil {
			bufrw.WriteString("error: Order not found\n")
			bufrw.Flush()
			return
		}
		bufrw.WriteString("status: " + string(ord.Status) + "\n")
		bufrw.Flush()
		if ord.Status == domain.StatusPaid || ord.Status == domain.StatusFailed {
			return
		}
		time.Sleep(2 * time.Second)
	}
}

func (h *Handler) InventoryEvents(ctx context.Context, m kafka.Message) {
	var envelope events.InventoryEventEnvelope
	if err := proto.Unmarshal(m.Value, &envelope); err != nil {
		slog.Warn("failed to unmarshal InventoryEventEnvelope: ", "err", err)
		return
	}

	slog.Info("Received inventory event", "topic", m.Topic, "partition", m.Partition, "offset", m.Offset)
	switch evt := envelope.Event.(type) {
	case *events.InventoryEventEnvelope_ReservationSucceeded:
		h.Service.UpdateOrderAwaitingPayment(ctx, evt.ReservationSucceeded.Id)
	case *events.InventoryEventEnvelope_ReservationFailed:
		h.Service.UpdateOrderFailed(ctx, evt.ReservationFailed.Id)
	default:
		slog.Warn("Unknown or missing event type in envelope")
	}
}

func (h *Handler) PaymentEvents(ctx context.Context, m kafka.Message) {
	var envelope events.PaymentEventEnvelope
	if err := proto.Unmarshal(m.Value, &envelope); err != nil {
		slog.Warn("Failed to unmarshal PaymentEventEnvelope:", "err", err)
		return
	}

	slog.Info("Received payment event", "topic", m.Topic, "partition", m.Partition, "offset", m.Offset)
	switch evt := envelope.Event.(type) {
	case *events.PaymentEventEnvelope_PaymentSucceeded:
		h.Service.UpdateOrderPaid(ctx, evt.PaymentSucceeded.Id)
	case *events.PaymentEventEnvelope_PaymentFailed:
		h.Service.UpdateOrderFailed(ctx, evt.PaymentFailed.Id)
	default:
		slog.Warn("Unknown or missing event type in envelope")
	}
}

func (h *Handler) processCreateOrderRequest(r *http.Request) ([]domain.Item, error) {
	var req httppb.CreateOrderRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if err := proto.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	if len(req.Items) == 0 {
		return nil, fmt.Errorf("no items provided")
	}

	domainItems := make([]domain.Item, len(req.Items))
	for i, item := range req.Items {
		domainItems[i] = domain.Item{
			ProductID: item.ProductId,
		}
	}

	return domainItems, nil
}

func (h *Handler) respondWithCreateOrderSuccess(w http.ResponseWriter, order *domain.Order) {
	protoOrder := &httppb.Order{
		Id:     order.ID,
		Status: string(order.Status),
	}

	for _, item := range order.Items {
		protoOrder.Items = append(protoOrder.Items, &httppb.OrderItem{
			ProductId: item.ProductID,
		})
	}

	response := &httppb.CreateOrderResponse{
		Order: protoOrder,
	}

	httputils.RespondProto(w, response, http.StatusCreated)
}

func (h *Handler) respondWithGetOrderSuccess(w http.ResponseWriter, order *domain.Order) {
	protoOrder := &httppb.Order{
		Id:        order.ID,
		Status:    string(order.Status),
		CreatedAt: order.CreatedAt.Format(time.RFC3339),
		UpdatedAt: order.UpdatedAt.Format(time.RFC3339),
	}

	for _, item := range order.Items {
		protoOrder.Items = append(protoOrder.Items, &httppb.OrderItem{
			ProductId: item.ProductID,
		})
	}

	response := &httppb.GetOrderResponse{
		Order: protoOrder,
	}

	httputils.RespondProto(w, response, http.StatusOK)
}
