package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/drwgrc/multi-tenant-event-pipeline/internal/observability"
)

func TestRequestLoggingIncludesRequestIDStatusAndDuration(t *testing.T) {
	logger, logs := newTestLogger(t)
	handler := RequestID(logger)(RequestLogging(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})))

	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	req.RemoteAddr = "203.0.113.1:1234"
	req.Header.Set(observability.RequestIDHeader, "req-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	entry := decodeSingleLogEntry(t, logs)
	assertLogString(t, entry, "service", "api")
	assertLogString(t, entry, "request_id", "req-123")
	assertLogString(t, entry, "method", http.MethodGet)
	assertLogString(t, entry, "path", "/livez")
	assertLogString(t, entry, "remote_addr", "203.0.113.1:1234")

	if got := int(entry["status"].(float64)); got != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", got, http.StatusAccepted)
	}

	if _, ok := entry["duration_ms"]; !ok {
		t.Fatal("duration_ms missing from log entry")
	}
}

func newTestLogger(t *testing.T) (*slog.Logger, *bytes.Buffer) {
	t.Helper()

	buffer := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buffer, nil)).With(slog.String("service", "api"))
	return logger, buffer
}

func decodeSingleLogEntry(t *testing.T, buffer *bytes.Buffer) map[string]any {
	t.Helper()

	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buffer.Bytes()), &entry); err != nil {
		t.Fatalf("json.Unmarshal(log) error = %v", err)
	}

	return entry
}

func assertLogString(t *testing.T, entry map[string]any, key, want string) {
	t.Helper()

	got, ok := entry[key].(string)
	if !ok {
		t.Fatalf("%s missing or not a string", key)
	}
	if got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}
