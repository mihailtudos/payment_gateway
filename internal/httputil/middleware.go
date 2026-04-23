package httputil

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type loggerKey struct{}

// RequestID reads or generates an X-Request-ID, stores it for chi middleware
// compatibility, and stashes an enriched slog.Logger in the context so
// downstream handlers always log with request_id attached.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		logger := slog.Default().With(slog.String("request_id", requestID))

		ctx := context.WithValue(r.Context(), middleware.RequestIDKey, requestID)
		ctx = context.WithValue(ctx, loggerKey{}, logger)

		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestLogger is a middleware that logs each request through
// slog (and through the OTel log pipeline if configured).
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		defer func() {
			LoggerFromContext(r.Context()).InfoContext(r.Context(), "request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Duration("duration", time.Since(start)),
			)
		}()

		next.ServeHTTP(ww, r)
	})
}

// LoggerFromContext returns the request-scoped logger stored by the RequestID
// middleware. Falls back to slog.Default() if called outside a request context.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
