package router

import (
	"net/http"

	"github.com/axmz/go-saga-microservices/payment-service/internal/handler"
)

func New(handlers *handler.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /payment-success", handlers.PaymentSuccess)
	mux.HandleFunc("POST /payment-fail", handlers.PaymentFail)
	return mux
}
