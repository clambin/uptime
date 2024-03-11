package logger

import (
	"context"
	"log/slog"
	"net/http"
)

var LoggerCtxValue = "logger"

func WithLogger(logger *slog.Logger) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), LoggerCtxValue, logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Logger(r *http.Request) *slog.Logger {
	if cv := r.Context().Value(LoggerCtxValue); cv != nil {
		if l, ok := cv.(*slog.Logger); ok {
			return l
		}
	}
	return slog.Default()
}
