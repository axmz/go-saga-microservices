package graceful

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Operation = func(ctx context.Context) error

func Shutdown(ctx context.Context, timeout time.Duration, ops map[string]Operation) <-chan struct{} {
	wait := make(chan struct{})
	go func() {
		s := make(chan os.Signal, 1)
		signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

		select {
		case sig := <-s:
			slog.Info("shutting down due to signal", "signal", sig)
		case <-ctx.Done():
			slog.Info("shutting down due to context cancellation", "reason", ctx.Err())
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		var wg sync.WaitGroup
		for key, op := range ops {
			wg.Add(1)
			go func() {
				defer wg.Done()
				slog.Info(fmt.Sprintf("cleaning up: %s", key))
				if err := op(shutdownCtx); err != nil {
					slog.Info(fmt.Sprintf("%s: clean up failed: %s", key, err.Error()))
					return
				}
				slog.Info(fmt.Sprintf("%s was shutdown gracefully", key))
			}()
		}
		wg.Wait()
		close(wait)
	}()
	return wait
}
