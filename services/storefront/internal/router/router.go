package router

import (
	"net/http"

	"github.com/axmz/go-saga-microservices/services/storefront/internal/handler"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/renderer"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/service"
)

func New(svc *service.Service, renderer *renderer.TemplateRenderer) *http.ServeMux {
	handlers := handler.New(svc, renderer)
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", handlers.HomeHandler)
	mux.HandleFunc("/payment", handlers.PaymentHandler)
	mux.HandleFunc("/confirmation", handlers.ConfirmationHandler)
	mux.HandleFunc("/order", handlers.OrderHandler)
	mux.HandleFunc("GET /api/products", handlers.APIProductsHandler)
	mux.HandleFunc("POST /api/orders", handlers.APICreateOrderHandler)
	return mux
}
