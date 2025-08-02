package handler

import (
	"encoding/json"
	"net/http"

	"github.com/axmz/go-saga-microservices/payment-service/internal/service"
)

type PaymentHandler struct {
	Service *service.Service
}

func New(service *service.Service) *PaymentHandler {
	return &PaymentHandler{
		Service: service,
	}
}

func (h *PaymentHandler) PaymentSuccess(w http.ResponseWriter, r *http.Request) {
	var data struct {
		OrderID string `json:"order_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.Service.PaymentSuccess(r.Context(), data.OrderID); err != nil {
		http.Error(w, "Failed to process payment success", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *PaymentHandler) PaymentFail(w http.ResponseWriter, r *http.Request) {
	var data struct {
		OrderID string `json:"order_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.Service.PaymentFail(r.Context(), data.OrderID); err != nil {
		http.Error(w, "Failed to process payment success", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
