package app

import (
	"log/slog"

	"github.com/axmz/go-saga-microservices/config"
	"github.com/axmz/go-saga-microservices/lib/adapter/http"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/renderer"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/router"
	"github.com/axmz/go-saga-microservices/services/storefront/internal/service"
)

type App struct {
	Config   *config.Config
	HTTP     *http.Server
	Log      *slog.Logger
	Services *service.Service
}

func SetupApp(
	cfg *config.Config,
	log *slog.Logger,
	srv *http.Server,
	renderer *renderer.TemplateRenderer,
) (*App, error) {
	svc := service.New(cfg)
	mux := router.New(svc, renderer)
	srv.Router.Handler = mux

	app := &App{
		Config:   cfg,
		HTTP:     srv,
		Log:      log,
		Services: svc,
	}

	slog.Info("Application initialized", slog.String("env", app.Config.Env))
	return app, nil
}
