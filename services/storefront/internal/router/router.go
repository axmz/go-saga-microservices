package router

import (
	"net/http"

	"github.com/axmz/go-saga-microservices/services/storefront/internal/handler"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/renderer"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/service"
)

func New(handlers *handler.Handler, svc *service.Service, renderer *renderer.TemplateRenderer) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("GET /payment/{orderId}", handlers.PaymentHandler)
	mux.HandleFunc("GET /order/{orderId}", handlers.OrderHandler)
	mux.HandleFunc("GET /orders/ws/{orderId}", handlers.OrderStatusWSHandler)
	mux.HandleFunc("GET /confirmation/{orderId}", handlers.ConfirmationHandler)
	mux.HandleFunc("GET /api/products", handlers.APIProductsHandler)
	mux.HandleFunc("POST /api/orders", handlers.APICreateOrderHandler)
	mux.HandleFunc("POST /api/payment-success", handlers.APIPaymentSuccessHandler)
	mux.HandleFunc("POST /api/payment-fail", handlers.APIPaymentFailHandler)
	return mux
}
