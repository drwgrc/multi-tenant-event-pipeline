package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestLoadConfigDefaultsDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/event_pipeline?sslmode=disable")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if cfg.DatabaseURL == "" {
		t.Fatal("loadConfig() database URL = empty, want populated value")
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
	t.Setenv("DATABASE_URL", "not-a-url")

	_, err := loadConfig()
	if err == nil {
		t.Fatal("loadConfig() error = nil, want error")
	}

	if !strings.Contains(err.Error(), "DATABASE_URL must be a valid URL") {
		t.Fatalf("loadConfig() error = %q, want valid URL error", err)
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
