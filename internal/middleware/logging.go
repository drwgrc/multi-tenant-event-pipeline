package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/drwgrc/multi-tenant-event-pipeline/internal/observability"
)

func RequestLogging(baseLogger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(recorder, r)

			logger := baseLogger
			requestLogger, hasRequestLogger := observability.LookupLogger(r.Context())
			if hasRequestLogger {
				logger = requestLogger
			}

			requestID, ok := observability.RequestIDFromContext(r.Context())
			if !ok {
				requestID = recorder.Header().Get(observability.RequestIDHeader)
			}

			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", recorder.status),
				slog.String("remote_addr", r.RemoteAddr),
				observability.DurationAttr(time.Since(start)),
			}
			if requestID != "" && !hasRequestLogger {
				attrs = append(attrs, observability.RequestIDAttr(requestID))
			}

			logger.LogAttrs(r.Context(), slog.LevelInfo, "http request complete", attrs...)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
