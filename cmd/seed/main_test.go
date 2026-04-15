package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestLoadConfigDefaultsDatabaseURL(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/event_pipeline?sslmode=disable")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if cfg.DatabaseURL == "" {
		t.Fatal("loadConfig() database URL = empty, want populated value")
	}

	if cfg.AppEnv != "development" {
		t.Fatalf("loadConfig() app env = %q, want %q", cfg.AppEnv, "development")
	}
}

func TestLoadConfigRejectsMissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := loadConfig()
	if err == nil {
		t.Fatal("loadConfig() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "DATABASE_URL is required") {
		t.Fatalf("loadConfig() error = %q, want DATABASE_URL is required", err)
	}
}

func TestLoadConfigRejectsInvalidDatabaseURL(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "not-a-url")

	_, err := loadConfig()
	if err == nil {
		t.Fatal("loadConfig() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "DATABASE_URL must be a valid URL") {
		t.Fatalf("loadConfig() error = %q, want valid URL error", err)
	}
}

func TestLoadConfigRejectsUnsafeAppEnvWithoutOverride(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/event_pipeline?sslmode=disable")
	t.Setenv("SEED_ALLOW_NON_LOCAL_DATABASE", "")

	_, err := loadConfig()
	if err == nil {
		t.Fatal("loadConfig() error = nil, want error")
	}

	if !strings.Contains(err.Error(), `refusing to seed when APP_ENV="production"`) {
		t.Fatalf("loadConfig() error = %q, want app env safety error", err)
	}
}

func TestLoadConfigRejectsUnsafeHostWithoutOverride(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@db.internal:5432/event_pipeline?sslmode=disable")
	t.Setenv("SEED_ALLOW_NON_LOCAL_DATABASE", "")

	_, err := loadConfig()
	if err == nil {
		t.Fatal("loadConfig() error = nil, want error")
	}

	if !strings.Contains(err.Error(), `refusing to seed non-local database host "db.internal"`) {
		t.Fatalf("loadConfig() error = %q, want host safety error", err)
	}
}

func TestLoadConfigAllowsUnsafeTargetWithExplicitOverride(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@db.internal:5432/event_pipeline?sslmode=disable")
	t.Setenv("SEED_ALLOW_NON_LOCAL_DATABASE", "true")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if !cfg.ForceNonLocal {
		t.Fatal("loadConfig() ForceNonLocal = false, want true")
	}
}

func TestAPIKeyPrefixUsesLeadingCharacters(t *testing.T) {
	got := apiKeyPrefix(demoAPIKeySecret)

	if got != "evt_demo_loc" {
		t.Fatalf("apiKeyPrefix() = %q, want %q", got, "evt_demo_loc")
	}
}

func TestAPIKeyHashUsesSHA256Hex(t *testing.T) {
	sum := sha256.Sum256([]byte(demoAPIKeySecret))
	want := hex.EncodeToString(sum[:])

	if got := apiKeyHash(demoAPIKeySecret); got != want {
		t.Fatalf("apiKeyHash() = %q, want %q", got, want)
	}
}
