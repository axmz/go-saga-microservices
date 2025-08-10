package router

import (
	"net/http"

	"github.com/axmz/go-saga-microservices/inventory-service/internal/handler"
)

func New(handlers *handler.Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /products", handlers.GetProducts)
	return mux
}
