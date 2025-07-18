package router

import (
	"net/http"

	"github.com/axmz/go-saga-microservices/services/order/internal/handler"
	"github.com/axmz/go-saga-microservices/services/order/internal/service"
)

func New(svc *service.Service) *http.ServeMux {
	handlers := handler.New(svc)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /orders", handlers.CreateOrder)
	mux.HandleFunc("GET /orders", handlers.GetOrder)
	mux.HandleFunc("GET /orders/ws", handlers.OrderStatusWS)
	return mux
}
