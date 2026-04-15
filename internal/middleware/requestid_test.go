package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/drwgrc/multi-tenant-event-pipeline/internal/observability"
)

func TestRequestIDPreservesIncomingHeader(t *testing.T) {
	logger, _ := newTestLogger(t)

	var gotRequestID string
	handler := RequestID(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		gotRequestID, ok = observability.RequestIDFromContext(r.Context())
		if !ok {
			t.Fatal("RequestIDFromContext() ok = false, want true")
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	req.Header.Set(observability.RequestIDHeader, "req-123")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if gotRequestID != "req-123" {
		t.Fatalf("request id = %q, want %q", gotRequestID, "req-123")
	}

	if got := rec.Header().Get(observability.RequestIDHeader); got != "req-123" {
		t.Fatalf("response %s = %q, want %q", observability.RequestIDHeader, got, "req-123")
	}
}

func TestRequestIDGeneratesHeaderWhenMissing(t *testing.T) {
	logger, _ := newTestLogger(t)

	handler := RequestID(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID, ok := observability.RequestIDFromContext(r.Context())
		if !ok {
			t.Fatal("RequestIDFromContext() ok = false, want true")
		}
		if requestID == "" {
			t.Fatal("request id = empty, want generated value")
		}

		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/livez", nil))

	if got := rec.Header().Get(observability.RequestIDHeader); got == "" {
		t.Fatalf("response %s = empty, want generated value", observability.RequestIDHeader)
	}
}
