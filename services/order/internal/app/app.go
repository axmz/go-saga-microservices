package app

import (
	"log/slog"

	"github.com/axmz/go-saga-microservices/config"
	"github.com/axmz/go-saga-microservices/lib/adapter/db"
	"github.com/axmz/go-saga-microservices/lib/adapter/http"
	"github.com/axmz/go-saga-microservices/lib/adapter/kafka"
	"github.com/axmz/go-saga-microservices/services/order/internal/consumer"
	"github.com/axmz/go-saga-microservices/services/order/internal/handler"
	"github.com/axmz/go-saga-microservices/services/order/internal/publisher"
	"github.com/axmz/go-saga-microservices/services/order/internal/repository"
	"github.com/axmz/go-saga-microservices/services/order/internal/router"
	"github.com/axmz/go-saga-microservices/services/order/internal/service"
	"github.com/axmz/go-saga-microservices/services/order/internal/sync"
)

type App struct {
	Config    *config.Config
	Consumer  *consumer.Consumer
	DB        *db.DB
	HTTP      *http.Server
	Kafka     *kafka.Broker
	Log       *slog.Logger
	Publisher *publisher.Publisher
	Repo      *repository.Repository
	Services  *service.Service
}

func SetupApp(
	cfg *config.Config,
	log *slog.Logger,
	db *db.DB,
	srv *http.Server,
	kfk *kafka.Broker,
) (*App, error) {
	rep := repository.New(db)
	pub := publisher.New(kfk.Writer)
	syn := sync.New()
	svc := service.New(rep, pub, syn)
	han := handler.New(svc)
	con := consumer.New(kfk.Reader, han)
	mux := router.New(svc, han)
	srv.Router.Handler = mux

	app := &App{
		Config:   cfg,
		Consumer: con,
		DB:       db,
		HTTP:     srv,
		Kafka:    kfk,
		Log:      log,
		Services: svc,
	}

	slog.Info("Application initialized", slog.String("env", app.Config.Env))
	return app, nil
}
