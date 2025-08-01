package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/axmz/go-saga-microservices/services/storefront/internal/domain"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/renderer"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/service"
)

// Service URLs
type Handler struct {
	Service  *service.Service
	Renderer *renderer.TemplateRenderer
}

func New(service *service.Service, renderer *renderer.TemplateRenderer) *Handler {
	return &Handler{
		Service:  service,
		Renderer: renderer,
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
	if r.Method == "POST" {
		orderID := r.FormValue("orderId")
		fail := r.FormValue("fail")
		if orderID == "" {
			http.Error(w, "Missing orderId", http.StatusBadRequest)
			return
		}
		// Call payment-service
		paymentURL := fmt.Sprintf("http://payment-service:8080/payment?orderId=%s&fail=%s", orderID, fail)
		resp, err := http.Post(paymentURL, "application/json", nil)
		if err != nil {
			http.Error(w, "Payment service error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		// Always redirect to order page after payment attempt
		http.Redirect(w, r, "/order?orderId="+orderID, http.StatusSeeOther)
		return
	}
	// Show payment page
	err := h.Renderer.Render(w, "payment.html", map[string]interface{}{
		"Title": "Payment",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) ConfirmationHandler(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("order_id")

	data := map[string]interface{}{
		"Title":   "Order Confirmation",
		"OrderID": orderID,
	}

	err := h.Renderer.Render(w, "confirmation.html", data)
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

// WS /orders/ws?orderId=...
func (h *Handler) OrderStatusWS(w http.ResponseWriter, r *http.Request) {
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
		// if ord.Status == domain.StatusPaid || ord.Status == domain.StatusFailed {
		// 	return
		// }
		time.Sleep(2 * time.Second)
	}
}
