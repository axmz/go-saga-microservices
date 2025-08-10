package http

import (
	"log/slog"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.status = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) Write(p []byte) (int, error) {
	if rr.status == 0 {
		// default status if Write called before WriteHeader
		rr.status = http.StatusOK
	}
	n, err := rr.ResponseWriter.Write(p)
	rr.bytes += n
	return n, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rr := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(rr, r)
		if rr.status == 0 {
			rr.status = http.StatusOK
		}
		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"status", rr.status,
			"bytes", rr.bytes,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
