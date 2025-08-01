package router

import (
	"net/http"

	"github.com/axmz/go-saga-microservices/services/order/internal/handler"
	"github.com/axmz/go-saga-microservices/services/order/internal/service"
)

func New(svc *service.Service, h *handler.OrderHandler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", h.CreateOrder)
	mux.HandleFunc("GET /orders/{orderID}", h.GetOrder)
	mux.HandleFunc("GET /orders/ws", h.OrderStatusWS)
	return mux
}
