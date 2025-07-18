package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/axmz/go-saga-microservices/lib/adapter/db"
	"github.com/axmz/go-saga-microservices/lib/adapter/http"
	"github.com/axmz/go-saga-microservices/lib/adapter/kafka"
	"github.com/axmz/go-saga-microservices/lib/logger"
	"github.com/axmz/go-saga-microservices/pkg/graceful"
	"github.com/axmz/go-saga-microservices/services/order/internal/app"
	"github.com/axmz/go-saga-microservices/services/order/internal/config"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// load config
	cfg, err := config.MustLoad()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// setup logger
	logger, err := logger.Setup(cfg.Env)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// connect to database
	db, err := db.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// initialize kafka
	kafka, err := kafka.Init(cfg.Kafka)
	if err != nil {
		logger.Error("Failed to initialize Kafka", "err", err)
		cancel()
	}

	// initialize http server
	srv, err := http.NewServer(&cfg.HTTP)
	if err != nil {
		logger.Error("Failed to initialize HTTP server", "err", err)
		cancel()
	}

	// setup app
	app, err := app.SetupApp(cfg, logger, db, srv, kafka)
	if err != nil {
		app.Log.Error("Failed to initialize app", "err", err)
		cancel()
	}

	// start blocking services
	var g errgroup.Group

	g.Go(func() error {
		return app.HTTP.Run()
	})

	g.Go(func() error {
		app.Consumer.Start(ctx)
		return nil
	})

	if err = g.Wait(); err != nil {
		app.Log.Error("Service error, shutting down", slog.String("err", err.Error()))
		cancel()
	}

	// shutdown services if ctx cancelled or signal received
	<-graceful.Shutdown(ctx, app.Config.GracefulTimeout, map[string]graceful.Operation{
		"kafka":       app.Kafka.Shutdown,
		"database":    app.DB.Shutdown,
		"http-server": app.HTTP.Shutdown,
	})

	app.Log.Info("Application stopped")
}
