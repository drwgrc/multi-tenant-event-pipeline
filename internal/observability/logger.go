package observability

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"
)

type loggerContextKey struct{}

func NewLogger(service string) *slog.Logger {
	return newLogger(os.Stdout, service)
}

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	logger, ok := LookupLogger(ctx)
	if !ok {
		return slog.Default()
	}

	return logger
}

func LookupLogger(ctx context.Context) (*slog.Logger, bool) {
	logger, ok := ctx.Value(loggerContextKey{}).(*slog.Logger)
	if !ok || logger == nil {
		return nil, false
	}

	return logger, true
}

func RequestIDAttr(requestID string) slog.Attr {
	return slog.String("request_id", requestID)
}

func TenantIDAttr(tenantID string) slog.Attr {
	return slog.String("tenant_id", tenantID)
}

func JobIDAttr(jobID string) slog.Attr {
	return slog.String("job_id", jobID)
}

func WorkerIDAttr(workerID string) slog.Attr {
	return slog.String("worker_id", workerID)
}

func DurationAttr(duration time.Duration) slog.Attr {
	return slog.Int64("duration_ms", duration.Milliseconds())
}

func newLogger(writer io.Writer, service string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(writer, nil)).With(slog.String("service", service))
}
