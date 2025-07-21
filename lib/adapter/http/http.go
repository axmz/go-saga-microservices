package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Config struct {
	Protocol     string
	Host         string
	Port         string
	IdleTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type Server struct {
	Router *http.Server
}

func NewServer(cfg Config) (*Server, error) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	r := &http.Server{
		Addr:         addr,
		IdleTimeout:  cfg.IdleTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	slog.Info("HTTP server initialized", "addr", addr)
	return &Server{Router: r}, nil
}

func (s *Server) Run() error {
	slog.Info(fmt.Sprintf("Starting server on %s", s.Router.Addr))
	if err := s.Router.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server ListenAndServe: %v", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.Router.Shutdown(ctx)
}
