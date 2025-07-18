package app

import (
	"log/slog"

	"github.com/axmz/go-saga-microservices/lib/adapter/db"
	"github.com/axmz/go-saga-microservices/lib/adapter/http"
	"github.com/axmz/go-saga-microservices/lib/adapter/kafka"
	"github.com/axmz/go-saga-microservices/services/order/internal/config"
	"github.com/axmz/go-saga-microservices/services/order/internal/consumer"
	"github.com/axmz/go-saga-microservices/services/order/internal/publisher"
	"github.com/axmz/go-saga-microservices/services/order/internal/repository"
	"github.com/axmz/go-saga-microservices/services/order/internal/router"
	"github.com/axmz/go-saga-microservices/services/order/internal/service"
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
	svc := service.New(rep, pub)
	mux := router.New(svc)
	con := consumer.New(kfk.Reader, svc)
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

	app.Log.Info("Application initialized", slog.String("env", app.Config.Env))

	return app, nil
}
