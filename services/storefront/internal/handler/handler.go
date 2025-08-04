package handler

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"

	"github.com/axmz/go-saga-microservices/pkg/events"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/domain"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/renderer"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/service"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/ws"
	"github.com/segmentio/kafka-go"
	"google.golang.org/protobuf/proto"
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

func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	products, err := h.Service.GetAllProducts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Products": products,
		"Title":    "Saga Microservices Storefront",
	}

	err = h.Renderer.Render(w, "home.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) PaymentHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("orderId")
	if orderID == "" {
		http.Error(w, "Missing orderId", http.StatusBadRequest)
		return
	}
	order, err := h.Service.GetOrder(r.Context(), orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]*domain.Order{
		"Order": order,
	}
	err = h.Renderer.Render(w, "payment.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) OrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("orderId")
	if orderID == "" {
		http.Error(w, "Missing orderId", http.StatusBadRequest)
		return
	}
	order, err := h.Service.GetOrder(r.Context(), orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]*domain.Order{
		"Order": order,
	}
	err = h.Renderer.Render(w, "order.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) ConfirmationHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("orderId")
	if orderID == "" {
		http.Error(w, "Missing orderId", http.StatusBadRequest)
		return
	}
	order, err := h.Service.GetOrder(r.Context(), orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]*domain.Order{
		"Order": order,
	}
	err = h.Renderer.Render(w, "confirmation.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) APIProductsHandler(w http.ResponseWriter, r *http.Request) {
	products, err := h.Service.GetAllProducts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func (h *Handler) APICreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := h.Service.CreateOrder(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&order)
}

func (h *Handler) APIPaymentSuccessHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	orderID := r.FormValue("order_id")
	if orderID == "" {
		http.Error(w, "Missing order_id", http.StatusBadRequest)
		return
	}

	if err := h.Service.PaymentSuccess(r.Context(), orderID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/confirmation/"+orderID, http.StatusSeeOther)
}

func (h *Handler) APIPaymentFailHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	orderID := r.FormValue("order_id")
	if orderID == "" {
		http.Error(w, "Missing order_id", http.StatusBadRequest)
		return
	}

	if err := h.Service.PaymentFail(r.Context(), orderID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/confirmation/"+orderID, http.StatusSeeOther)
}

func (h *Handler) OrderStatusWSHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.PathValue("orderId")
	if orderID == "" {
		http.Error(w, "Missing orderId", http.StatusBadRequest)
		return
	}

	conn, err := ws.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	h.WSManager.Register(orderID, conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("WS read error:", err)
			break
		}
	}

	h.WSManager.Unregister(orderID, conn)
}

func (h *Handler) PaymentEvents(ctx context.Context, m kafka.Message) {
	var envelope events.PaymentEventEnvelope
	if err := proto.Unmarshal(m.Value, &envelope); err != nil {
		slog.Warn("Failed to unmarshal PaymentEventEnvelope:", "err", err)
		return
	}

	switch evt := envelope.Event.(type) {
	case *events.PaymentEventEnvelope_PaymentSucceeded:
		h.WSManager.Broadcast(evt.PaymentSucceeded.Id, "status: Paid")
	case *events.PaymentEventEnvelope_PaymentFailed:
		h.WSManager.Broadcast(evt.PaymentFailed.Id, "status: Failed")
	default:
		slog.Warn("Unknown or missing event type in envelope")
	}
}
