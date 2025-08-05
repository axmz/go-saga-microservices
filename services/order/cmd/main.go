package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/axmz/go-graceful"
	"github.com/axmz/go-saga-microservices/config"
	"github.com/axmz/go-saga-microservices/lib/adapter/db"
	"github.com/axmz/go-saga-microservices/lib/adapter/http"
	"github.com/axmz/go-saga-microservices/lib/adapter/kafka"
	"github.com/axmz/go-saga-microservices/lib/logger"
	"github.com/axmz/go-saga-microservices/services/order/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	log.SetOutput(os.Stdout)
	log.Println("Order service starting")

	// load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// setup logger
	logger, err := logger.Setup(cfg.Env)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// connect to database
	db, err := db.Connect(db.Config(cfg.Order.DB))
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// initialize kafka
	kafka, err := kafka.Init(kafka.Config(cfg.Order.Kafka))
	if err != nil {
		slog.Error("Failed to initialize Kafka:", "err", err)
		cancel()
	}

	// initialize http server
	srv, err := http.NewServer((http.Config(cfg.Order.HTTP)))
	if err != nil {
		slog.Error("Failed to initialize HTTP server:", "err", err)
		cancel()
	}

	// setup app
	app, err := app.SetupApp(cfg, logger, db, srv, kafka)
	if err != nil {
		slog.Error("Failed to initialize app:", "err", err)
		cancel()
	}

	// HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.HTTP.Run(); err != nil {
			slog.Error("HTTP server terminated:", "err", err)
			cancel()
		}
	}()

	// Kafka consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := app.Consumer.Start(ctx); err != nil {
			slog.Error("Kafka consumer group terminated:", "err", err)
			cancel()
		}
	}()

	// Wait for shutdown signal or context cancellation
	<-graceful.Shutdown(ctx, app.Config.GracefulTimeout, map[string]graceful.Operation{
		"kafka":       app.Kafka.Shutdown,
		"database":    app.DB.Shutdown,
		"http-server": app.HTTP.Shutdown,
	})

	wg.Wait()
	app.Log.Warn("Application stopped")
}
