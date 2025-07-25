package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"log"

	"github.com/axmz/go-saga-microservices/services/order/internal/domain"
	"github.com/axmz/go-saga-microservices/services/order/internal/service"
)

type OrderHandler struct {
	Service *service.Service
}

func New(service *service.Service) *OrderHandler {
	return &OrderHandler{
		Service: service,
	}
}

// POST /orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Items []domain.OrderItem `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.Items) == 0 {
		http.Error(w, "No items provided", http.StatusBadRequest)
		return
	}
	ord := domain.NewOrder(req.Items)
	if err := h.Service.CreateOrder(r.Context(), ord); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[OrderService] Created order: %s, status: %s", ord.ID, ord.Status)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ord)
}

// GET /orders?orderId=...
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	orderID := r.URL.Query().Get("orderId")
	if orderID == "" {
		http.Error(w, "Missing orderId", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	ord, err := h.Service.GetOrder(ctx, orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if ord == nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	if ord != nil {
		log.Printf("[OrderService] Fetched order: %s, status: %s", ord.ID, ord.Status)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ord)
}

// WS /orders/ws?orderId=...
func (h *OrderHandler) OrderStatusWS(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("orderId")
	if orderID == "" {
		http.Error(w, "Missing orderId", http.StatusBadRequest)
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
		bufrw.WriteString("status: " + ord.Status + "\n")
		bufrw.Flush()
		if ord.Status == domain.StatusPaid || ord.Status == domain.StatusFailed {
			return
		}
		time.Sleep(2 * time.Second)
	}
}
