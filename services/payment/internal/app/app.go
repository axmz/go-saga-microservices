package app

import (
	"log/slog"

	"github.com/axmz/go-saga-microservices/config"
	"github.com/axmz/go-saga-microservices/lib/adapter/http"
	"github.com/axmz/go-saga-microservices/lib/adapter/kafka"
	"github.com/axmz/go-saga-microservices/payment-service/internal/handler"
	"github.com/axmz/go-saga-microservices/payment-service/internal/publisher"
	"github.com/axmz/go-saga-microservices/payment-service/internal/router"
	"github.com/axmz/go-saga-microservices/payment-service/internal/service"
)

type App struct {
	Config    *config.Config
	HTTP      *http.Server
	Kafka     *kafka.Broker
	Log       *slog.Logger
	Publisher *publisher.Publisher
	Services  *service.Service
}

func SetupApp(
	cfg *config.Config,
	log *slog.Logger,
	srv *http.Server,
	kfk *kafka.Broker,
) (*App, error) {
	pub := publisher.New(kfk.Writer)
	svc := service.New(pub)
	han := handler.New(svc)
	mux := router.New(han)
	srv.Router.Handler = mux

	app := &App{
		Config:   cfg,
		HTTP:     srv,
		Kafka:    kfk,
		Log:      log,
		Services: svc,
	}

	slog.Info("Application initialized", slog.String("env", app.Config.Env))
	return app, nil
}
