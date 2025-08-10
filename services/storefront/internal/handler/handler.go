package handler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	httputils "github.com/axmz/go-saga-microservices/lib/adapter/http"
	"github.com/axmz/go-saga-microservices/pkg/proto/events"
	httppb "github.com/axmz/go-saga-microservices/pkg/proto/http"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/renderer"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/service"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/ws"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// TODO: DRY
const (
	OrderIDPathParam = "orderId"
)

type Handler struct {
	Service   *service.Service
	Renderer  *renderer.TemplateRenderer
	WSManager *ws.WSManager
}

func New(service *service.Service, renderer *renderer.TemplateRenderer, wsManager *ws.WSManager) *Handler {
	return &Handler{
		Service:   service,
		Renderer:  renderer,
		WSManager: wsManager,
	}
}

// PAGES
func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		err := errors.New("not found")
		slog.Warn("HomePage not found", "path", r.URL.Path, "err", err)
		httputils.ErrorNotFound(w, err)
		return
	}

	products, err := h.Service.GetProducts(r.Context())
	if err != nil {
		slog.Error("GetProducts failed", "err", err)
		httputils.ErrorInternal(w, err)
		return
	}

	if err = h.Renderer.Render(w, "home.html", map[string]any{
		"Products": products,
		"Title":    "Saga Microservices Storefront",
	}); err != nil {
		slog.Error("Render home.html failed", "err", err)
		httputils.ErrorInternal(w, err)
		return
	}
	slog.Info("HomePage served", "status", http.StatusOK)
}

func (h *Handler) PaymentPage(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue(OrderIDPathParam)
	if orderID == "" {
		slog.Warn("PaymentPage missing orderId")
		httputils.ErrorBadRequest(w, errors.New("missing orderId"))
		return
	}

	order, err := h.Service.GetOrder(r.Context(), orderID)
	if err != nil {
		slog.Error("GetOrder failed", "orderId", orderID, "err", err)
		httputils.ErrorInternal(w, err)
		return
	}

	if err = h.Renderer.Render(w, "payment.html", map[string]any{
		"Order": order,
	}); err != nil {
		slog.Error("Render payment.html failed", "orderId", orderID, "err", err)
		httputils.ErrorInternal(w, err)
		return
	}
	slog.Info("PaymentPage served", "orderId", orderID, "status", http.StatusOK)
}

func (h *Handler) OrderPage(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue(OrderIDPathParam)
	if orderID == "" {
		slog.Warn("OrderPage missing orderId")
		httputils.ErrorBadRequest(w, errors.New("missing orderId"))
		return
	}

	order, err := h.Service.GetOrder(r.Context(), orderID)
	if err != nil {
		slog.Error("GetOrder failed", "orderId", orderID, "err", err)
		httputils.ErrorInternal(w, err)
		return
	}

	if err = h.Renderer.Render(w, "order.html", map[string]*httppb.Order{
		"Order": order,
	}); err != nil {
		slog.Error("Render order.html failed", "orderId", orderID, "err", err)
		httputils.ErrorInternal(w, err)
		return
	}
	slog.Info("OrderPage served", "orderId", orderID, "status", http.StatusOK)
}

func (h *Handler) ConfirmationPage(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue(OrderIDPathParam)
	if orderID == "" {
		slog.Warn("ConfirmationPage missing orderId")
		httputils.ErrorBadRequest(w, errors.New("missing orderId"))
		return
	}

	order, err := h.Service.GetOrder(r.Context(), orderID)
	if err != nil {
		slog.Error("GetOrder failed", "orderId", orderID, "err", err)
		httputils.ErrorInternal(w, err)
		return
	}

	if err = h.Renderer.Render(w, "confirmation.html", map[string]*httppb.Order{
		"Order": order,
	}); err != nil {
		slog.Error("Render confirmation.html failed", "orderId", orderID, "err", err)
		httputils.ErrorInternal(w, err)
		return
	}
	slog.Info("ConfirmationPage served", "orderId", orderID, "status", http.StatusOK)
}

// API
func (h *Handler) APIGetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.Service.GetProducts(r.Context())
	if err != nil {
		slog.Error("APIGetProducts GetProducts failed", "err", err)
		httputils.ErrorInternal(w, err)
		return
	}

	slog.Info("APIGetProducts success", "count", len(products))
	h.respondWithGetProductsResponse(w, products)
}

func (h *Handler) APICreateOrder(w http.ResponseWriter, r *http.Request) {
	req, err := h.processCreateOrderRequest(r)
	if err != nil {
		slog.Warn("APICreateOrder bad request", "err", err)
		httputils.ErrorBadRequest(w, err)
		return
	}

	order, err := h.Service.CreateOrder(r.Context(), req)
	if err != nil {
		slog.Error("CreateOrder failed", "err", err)
		httputils.ErrorInternal(w, err)
		return
	}

	slog.Info("APICreateOrder success", "orderId", order.Id)
	h.respondWithCreateOrderSuccess(w, order)
}

func (h *Handler) APIPaymentSuccess(w http.ResponseWriter, r *http.Request) {
	req, err := h.processPaymentSuccessRequest(r)
	if err != nil {
		slog.Warn("APIPaymentSuccess bad request", "err", err)
		httputils.ErrorBadRequest(w, err)
		return
	}

	if err := h.Service.PaymentSuccess(r.Context(), req.OrderId); err != nil {
		slog.Error("PaymentSuccess failed", "orderId", req.OrderId, "err", err)
		httputils.ErrorInternal(w, err)
		return
	}

	slog.Info("APIPaymentSuccess success", "orderId", req.OrderId)
	h.respondWithPaymentSuccess(w)
}

func (h *Handler) APIPaymentFail(w http.ResponseWriter, r *http.Request) {
	req, err := h.processPaymentFailRequest(r)
	if err != nil {
		slog.Warn("APIPaymentFail bad request", "err", err)
		httputils.ErrorBadRequest(w, err)
		return
	}

	if err := h.Service.PaymentFail(r.Context(), req.OrderId); err != nil {
		slog.Error("PaymentFail failed", "orderId", req.OrderId, "err", err)
		httputils.ErrorInternal(w, err)
		return
	}

	slog.Info("APIPaymentFail success", "orderId", req.OrderId)
	h.respondWithPaymentFail(w)
}

func (h *Handler) APIResetProducts(w http.ResponseWriter, r *http.Request) {
	if err := h.Service.ResetInventory(r.Context()); err != nil {
		slog.Error("APIResetProducts failed", "err", err)
		httputils.ErrorInternal(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// WS
func (h *Handler) WSOrderStatus(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue(OrderIDPathParam)
	if orderID == "" {
		httputils.ErrorBadRequest(w, errors.New("missing orderId"))
		return
	}

	conn, err := ws.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("WebSocket upgrade error:", "err", err)
		return
	}
	defer conn.Close()

	slog.Info("WS connected", "orderId", orderID, "remote", r.RemoteAddr)
	h.WSManager.Register(orderID, conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			slog.Warn("WS read error:", "err", err)
			break
		}
	}

	h.WSManager.Unregister(orderID, conn)
	slog.Info("WS disconnected", "orderId", orderID)
}

// EVENTS
func (h *Handler) PaymentEvents(ctx context.Context, m kafka.Message) {
	slog.Info("PaymentEvents received", "topic", m.Topic, "partition", m.Partition, "offset", m.Offset)
	var envelope events.PaymentEventEnvelope
	if err := proto.Unmarshal(m.Value, &envelope); err != nil {
		slog.Warn("Failed to unmarshal PaymentEventEnvelope:", "err", err)
		return
	}

	switch evt := envelope.Event.(type) {
	case *events.PaymentEventEnvelope_PaymentSucceeded:
		orderID := evt.PaymentSucceeded.Id
		slog.Info("Payment event: succeeded", "orderId", orderID)
		h.WSManager.Broadcast(orderID, "Paid")
	case *events.PaymentEventEnvelope_PaymentFailed:
		orderID := evt.PaymentFailed.Id
		slog.Info("Payment event: failed", "orderId", orderID)
		h.WSManager.Broadcast(orderID, "Failed")

	default:
		slog.Warn("Unknown or missing event type in envelope")
	}
}

// REQ PROCESSING
func (h *Handler) parseProtoJSONBody(r *http.Request, msg proto.Message) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("Read body failed", "path", r.URL.Path, "err", err)
		return err
	}

	if err := protojson.Unmarshal(body, msg); err != nil {
		typeName := string(msg.ProtoReflect().Descriptor().FullName())
		slog.Warn("Unmarshal "+typeName+" failed", "err", err, "bytes", len(body))
		return err
	}
	return nil
}

func (h *Handler) processCreateOrderRequest(r *http.Request) (*httppb.CreateOrderRequest, error) {
	req := new(httppb.CreateOrderRequest)
	if err := h.parseProtoJSONBody(r, req); err != nil {
		return nil, err
	}
	slog.Debug("Parsed CreateOrderRequest", "items", len(req.GetItems()))
	return req, nil
}

func (h *Handler) processPaymentSuccessRequest(r *http.Request) (*httppb.PaymentSuccessRequest, error) {
	req := new(httppb.PaymentSuccessRequest)
	if err := h.parseProtoJSONBody(r, req); err != nil {
		return nil, err
	}
	slog.Debug("Parsed PaymentSuccessRequest", "orderId", req.OrderId)
	return req, nil
}

func (h *Handler) processPaymentFailRequest(r *http.Request) (*httppb.PaymentFailRequest, error) {
	req := new(httppb.PaymentFailRequest)
	if err := h.parseProtoJSONBody(r, req); err != nil {
		return nil, err
	}
	slog.Debug("Parsed PaymentFailRequest", "orderId", req.OrderId)
	return req, nil
}

// RESPONSES
func (h *Handler) respondWithPaymentFail(w http.ResponseWriter) {
	response := &httppb.PaymentFailResponse{Success: true}
	httputils.RespondJSON(w, response, http.StatusOK)
}

func (h *Handler) respondWithPaymentSuccess(w http.ResponseWriter) {
	response := &httppb.PaymentSuccessResponse{Success: true}
	httputils.RespondJSON(w, response, http.StatusOK)
}

func (h *Handler) respondWithGetProductsResponse(w http.ResponseWriter, products []*httppb.Product) {
	response := &httppb.GetProductsResponse{Products: products}
	httputils.RespondJSON(w, response, http.StatusOK)
}

func (h *Handler) respondWithCreateOrderSuccess(w http.ResponseWriter, order *httppb.Order) {
	response := &httppb.CreateOrderResponse{Order: order}
	httputils.RespondJSON(w, response, http.StatusCreated)
}
