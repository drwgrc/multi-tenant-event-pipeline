package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLivenessHandlerReportsOK(t *testing.T) {
	handler := NewLivenessHandler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body livenessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("status body = %q, want %q", body.Status, "ok")
	}
}

func TestReadinessHandlerReportsHealthyDependencies(t *testing.T) {
	handler := NewReadinessHandler(
		time.Second,
		CheckFunc(func(context.Context) error { return nil }),
		CheckFunc(func(context.Context) error { return nil }),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body readinessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("body status = %q, want %q", body.Status, "ok")
	}

	if body.Database.Status != "ok" {
		t.Fatalf("database status = %q, want %q", body.Database.Status, "ok")
	}

	if body.Redis.Status != "ok" {
		t.Fatalf("redis status = %q, want %q", body.Redis.Status, "ok")
	}
}

func TestReadinessHandlerReportsDatabaseFailure(t *testing.T) {
	handler := NewReadinessHandler(
		time.Second,
		CheckFunc(func(context.Context) error { return errors.New("db down") }),
		CheckFunc(func(context.Context) error { return nil }),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body readinessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if body.Database.Status != "error" {
		t.Fatalf("database status = %q, want %q", body.Database.Status, "error")
	}

	if body.Database.Error != "db down" {
		t.Fatalf("database error = %q, want %q", body.Database.Error, "db down")
	}

	if body.Redis.Status != "ok" {
		t.Fatalf("redis status = %q, want %q", body.Redis.Status, "ok")
	}
}

func TestReadinessHandlerReportsRedisFailure(t *testing.T) {
	handler := NewReadinessHandler(
		time.Second,
		CheckFunc(func(context.Context) error { return nil }),
		CheckFunc(func(context.Context) error { return errors.New("redis down") }),
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}

	var body readinessResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	if body.Redis.Status != "error" {
		t.Fatalf("redis status = %q, want %q", body.Redis.Status, "error")
	}

	if body.Redis.Error != "redis down" {
		t.Fatalf("redis error = %q, want %q", body.Redis.Error, "redis down")
	}

	if body.Database.Status != "ok" {
		t.Fatalf("database status = %q, want %q", body.Database.Status, "ok")
	}
}
