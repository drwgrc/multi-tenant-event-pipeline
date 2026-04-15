package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/drwgrc/multi-tenant-event-pipeline/internal/observability"
)

func RequestID(baseLogger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := strings.TrimSpace(r.Header.Get(observability.RequestIDHeader))
			if requestID == "" {
				requestID = observability.NewCorrelationID()
			}

			ctx := observability.ContextWithRequestID(r.Context(), requestID)
			ctx = observability.ContextWithLogger(ctx, baseLogger.With(observability.RequestIDAttr(requestID)))

			w.Header().Set(observability.RequestIDHeader, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
