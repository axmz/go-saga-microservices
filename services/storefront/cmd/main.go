package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/axmz/go-saga-microservices/config"
	"github.com/axmz/go-saga-microservices/lib/adapter/http"
	"github.com/axmz/go-saga-microservices/lib/adapter/kafka"
	"github.com/axmz/go-saga-microservices/lib/logger"
	"github.com/axmz/go-saga-microservices/pkg/graceful"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/app"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/renderer"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/ws"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	log.SetOutput(os.Stdout)
	log.Println("Storefront service starting")

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

	// initialize kafka
	kafka, err := kafka.Init(kafka.Config(cfg.Storefront.Kafka))
	if err != nil {
		slog.Error("Failed to initialize Kafka:", "err", err)
		cancel()
	}

	// initialize renderer
	renderer, err := renderer.NewTemplateRenderer()
	if err != nil {
		log.Fatalf("Failed to create TemplateRenderer: %v", err)
	}

	// initialize http server
	srv, err := http.NewServer((http.Config(cfg.Storefront.HTTP)))
	if err != nil {
		log.Fatalf("Failed to initialize HTTP server %v", err)
	}

	// initialize WebSocket manager
	wsManager := ws.NewWSManager()
	if wsManager == nil {
		log.Fatalf("Failed to create WebSocket manager")
	}

	// setup app
	app, err := app.SetupApp(cfg, logger, srv, renderer, wsManager, kafka)
	if err != nil {
		log.Fatalf("Failed to initialize app %v", err)
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
		"http-server": app.HTTP.Shutdown,
	})

	app.Log.Info("Application stopped")
}
