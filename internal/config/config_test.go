package config

import (
	"strings"
	"testing"
)

func TestLoadAPIWithDevelopmentDefaults(t *testing.T) {
	setValidCommonEnv(t)
	t.Setenv("APP_ENV", "development")
	t.Setenv("HTTP_ADDR", "")

	cfg, err := LoadAPI()
	if err != nil {
		t.Fatalf("LoadAPI() error = %v", err)
	}

	if cfg.AppEnv != "development" {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, "development")
	}

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want %q", cfg.HTTPAddr, ":8080")
	}
}

func TestLoadWorkerDoesNotRequireHTTPAddr(t *testing.T) {
	setValidCommonEnv(t)
	t.Setenv("HTTP_ADDR", "")

	cfg, err := LoadWorker()
	if err != nil {
		t.Fatalf("LoadWorker() error = %v", err)
	}

	if cfg.HTTPAddr != "" {
		t.Fatalf("HTTPAddr = %q, want empty", cfg.HTTPAddr)
	}
}

func TestLoadAPIReportsMissingRequiredValues(t *testing.T) {
	cfg, err := LoadAPI()
	if err == nil {
		t.Fatalf("LoadAPI() error = nil, want error")
	}

	if cfg != (Config{}) {
		t.Fatalf("LoadAPI() config = %#v, want zero value", cfg)
	}

	assertErrorContains(t, err, "DATABASE_URL is required")
	assertErrorContains(t, err, "REDIS_URL is required")
}

func TestLoadAPIRejectsMalformedURLs(t *testing.T) {
	setValidCommonEnv(t)
	t.Setenv("DATABASE_URL", "not-a-url")
	t.Setenv("REDIS_URL", "://bad")

	_, err := LoadAPI()
	if err == nil {
		t.Fatal("LoadAPI() error = nil, want error")
	}

	assertErrorContains(t, err, "DATABASE_URL must be a valid URL")
	assertErrorContains(t, err, "REDIS_URL must be a valid URL")
}

func TestLoadAPIRejectsInvalidHTTPAddr(t *testing.T) {
	setValidCommonEnv(t)
	t.Setenv("HTTP_ADDR", "localhost")

	_, err := LoadAPI()
	if err == nil {
		t.Fatal("LoadAPI() error = nil, want error")
	}

	assertErrorContains(t, err, "HTTP_ADDR must be a valid TCP listen address")
}

func TestLoadAPIUsesRedisURLOverRedisAddr(t *testing.T) {
	setValidCommonEnv(t)
	t.Setenv("REDIS_URL", "redis://from-url:6379")
	t.Setenv("REDIS_ADDR", "from-addr:6380")

	cfg, err := LoadAPI()
	if err != nil {
		t.Fatalf("LoadAPI() error = %v", err)
	}

	if cfg.RedisURL != "redis://from-url:6379" {
		t.Fatalf("RedisURL = %q, want %q", cfg.RedisURL, "redis://from-url:6379")
	}
}

func TestLoadAPIUsesRedisAddrFallback(t *testing.T) {
	setValidCommonEnv(t)
	t.Setenv("REDIS_URL", "")
	t.Setenv("REDIS_ADDR", "redis:6379")

	cfg, err := LoadAPI()
	if err != nil {
		t.Fatalf("LoadAPI() error = %v", err)
	}

	if cfg.RedisURL != "redis://redis:6379" {
		t.Fatalf("RedisURL = %q, want %q", cfg.RedisURL, "redis://redis:6379")
	}
}

func setValidCommonEnv(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/event_pipeline?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("HTTP_ADDR", "127.0.0.1:8080")
}

func assertErrorContains(t *testing.T, err error, want string) {
	t.Helper()
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("error %q does not contain %q", err.Error(), want)
	}
}
