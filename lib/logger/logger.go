package logger

import (
	"log/slog"
	"os"
)

const (
	Local       = "local"
	Development = "dev"
	Production  = "prod"
)

func Setup(env string) (*slog.Logger, error) {
	var log *slog.Logger

	switch env {
	case Local:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))
	case Development:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))
	case Production:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: true}))
	}

	slog.SetDefault(log)
	log.Info("Logger initialized", "env", env)
	return log, nil
}
