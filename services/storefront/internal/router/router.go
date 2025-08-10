package router

import (
	"fmt"
	"net/http"

	"github.com/axmz/go-saga-microservices/services/storefront/internal/handler"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/renderer"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/service"
)

// TODO: DRY
const (
	OrderIDPathParam = "orderId"
)

var (
	root                  = "/"
	static                = "/static/"
	routeOrderPage        = fmt.Sprintf("GET /order/{%s}", OrderIDPathParam)
	routePaymentPage      = fmt.Sprintf("GET /payment/{%s}", OrderIDPathParam)
	routeConfirmationPage = fmt.Sprintf("GET /confirmation/{%s}", OrderIDPathParam)
	routeWSOrder          = fmt.Sprintf("GET /orders/ws/{%s}", OrderIDPathParam)
)

func New(handlers *handler.Handler, svc *service.Service, renderer *renderer.TemplateRenderer) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle(static, http.StripPrefix(static, http.FileServer(http.Dir("static"))))
	mux.HandleFunc(root, handlers.HomePage)

	mux.HandleFunc(routeOrderPage, handlers.OrderPage)
	mux.HandleFunc(routePaymentPage, handlers.PaymentPage)
	mux.HandleFunc(routeConfirmationPage, handlers.ConfirmationPage)

	mux.HandleFunc(routeWSOrder, handlers.WSOrderStatus)

	mux.HandleFunc("GET /api/products", handlers.APIGetProducts)
	mux.HandleFunc("POST /api/orders", handlers.APICreateOrder)
	mux.HandleFunc("POST /api/payment-success", handlers.APIPaymentSuccess)
	mux.HandleFunc("POST /api/payment-fail", handlers.APIPaymentFail)

	return mux
}
