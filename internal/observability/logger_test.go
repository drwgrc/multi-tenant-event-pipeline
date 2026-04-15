package observability

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestContextWithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(testDiscardWriter{}, nil))
	ctx := ContextWithLogger(context.Background(), logger)

	if got := LoggerFromContext(ctx); got != logger {
		t.Fatalf("LoggerFromContext() = %p, want %p", got, logger)
	}
}

func TestRequestIDContextRoundTrip(t *testing.T) {
	ctx := ContextWithRequestID(context.Background(), "req-123")

	requestID, ok := RequestIDFromContext(ctx)
	if !ok {
		t.Fatal("RequestIDFromContext() ok = false, want true")
	}

	if requestID != "req-123" {
		t.Fatalf("RequestIDFromContext() = %q, want %q", requestID, "req-123")
	}
}

func TestLoggingAttrs(t *testing.T) {
	assertAttr(t, RequestIDAttr("req-123"), "request_id", "req-123")
	assertAttr(t, TenantIDAttr("tenant-123"), "tenant_id", "tenant-123")
	assertAttr(t, JobIDAttr("job-123"), "job_id", "job-123")
	assertAttr(t, WorkerIDAttr("worker-123"), "worker_id", "worker-123")

	durationAttr := DurationAttr(1500 * time.Millisecond)
	if durationAttr.Key != "duration_ms" {
		t.Fatalf("DurationAttr key = %q, want %q", durationAttr.Key, "duration_ms")
	}
	if got := durationAttr.Value.Int64(); got != 1500 {
		t.Fatalf("DurationAttr value = %d, want %d", got, 1500)
	}
}

func assertAttr(t *testing.T, attr slog.Attr, wantKey, wantValue string) {
	t.Helper()

	if attr.Key != wantKey {
		t.Fatalf("attr key = %q, want %q", attr.Key, wantKey)
	}
	if got := attr.Value.String(); got != wantValue {
		t.Fatalf("attr value = %q, want %q", got, wantValue)
	}
}

type testDiscardWriter struct{}

func (testDiscardWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
