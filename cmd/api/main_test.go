package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/drwgrc/multi-tenant-event-pipeline/internal/httpapi"
	"github.com/drwgrc/multi-tenant-event-pipeline/internal/observability"
)

type healthResponse struct {
	Status string `json:"status"`
}

func TestNewHandlerServesLiveness(t *testing.T) {
	handler := newHandler(
		observability.NewLogger("test"),
		httpapi.CheckFunc(func(context.Context) error { return nil }),
		httpapi.CheckFunc(func(context.Context) error { return nil }),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body healthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("status body = %q, want %q", body.Status, "ok")
	}
}

func TestNewHandlerServesReadiness(t *testing.T) {
	handler := newHandler(
		observability.NewLogger("test"),
		httpapi.CheckFunc(func(context.Context) error { return nil }),
		httpapi.CheckFunc(func(context.Context) error { return errors.New("redis down") }),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}
