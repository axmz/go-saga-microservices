package main

import (
	"context"
	"log"
	"os"

	"github.com/axmz/go-saga-microservices/config"
	"github.com/axmz/go-saga-microservices/lib/adapter/http"
	"github.com/axmz/go-saga-microservices/lib/logger"
	"github.com/axmz/go-saga-microservices/pkg/graceful"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/app"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/renderer"
)

func main() {
	ctx, _ := context.WithCancel(context.Background())

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

	// setup app
	app, err := app.SetupApp(cfg, logger, srv, renderer)
	if err != nil {
		log.Fatalf("Failed to initialize app %v", err)
	}

	err = app.HTTP.Run()
	if err != nil {
		log.Fatalf("Failed to start the server %v", err)
	}

	<-graceful.Shutdown(ctx, app.Config.GracefulTimeout, map[string]graceful.Operation{
		"http-server": app.HTTP.Shutdown,
	})

	app.Log.Info("Application stopped")
}
