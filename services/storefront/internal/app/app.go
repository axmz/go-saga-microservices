package app

import (
	"log/slog"

	"github.com/axmz/go-saga-microservices/config"
	"github.com/axmz/go-saga-microservices/lib/adapter/http"
	"github.com/axmz/go-saga-microservices/lib/adapter/kafka"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/client"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/consumer"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/handler"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/renderer"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/router"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/service"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/ws"
)

type App struct {
	Config    *config.Config
	HTTP      *http.Server
	Log       *slog.Logger
	Services  *service.Service
	WSManager *ws.WSManager
	Kafka     *kafka.Broker
	Consumer  *consumer.Consumer
}

func SetupApp(
	cfg *config.Config,
	log *slog.Logger,
	srv *http.Server,
	renderer *renderer.TemplateRenderer,
	wsManager *ws.WSManager,
	kfk *kafka.Broker,
) (*App, error) {
	ocl := client.NewHTTPOrderClient(cfg.Order.HTTP.URL())
	pcl := client.NewHTTPPaymentClient(cfg.Payment.HTTP.URL())
	icl := client.NewHTTPInventoryClient(cfg.Inventory.HTTP.URL())
	svc := service.New(cfg, ocl, pcl, icl)
	han := handler.New(svc, renderer, wsManager)
	mux := router.New(han, svc, renderer)
	con := consumer.New(kfk.Reader, han)
	srv.Router.Handler = http.LoggingMiddleware(mux)

	app := &App{
		Config:   cfg,
		HTTP:     srv,
		Log:      log,
		Services: svc,
		Consumer: con,
	}

	slog.Info("Application initialized", slog.String("env", app.Config.Env))
	return app, nil
}
