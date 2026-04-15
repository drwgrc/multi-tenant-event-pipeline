package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigDefaultsMigrationsDir(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/event_pipeline?sslmode=disable")
	t.Setenv("MIGRATIONS_DIR", "")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}

	if cfg.MigrationsDir != defaultMigrationsDir {
		t.Fatalf("loadConfig() migrations dir = %q, want %q", cfg.MigrationsDir, defaultMigrationsDir)
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

func TestMigrationSourceURLReturnsFileScheme(t *testing.T) {
	dir := t.TempDir()

	got, err := migrationSourceURL(dir)
	if err != nil {
		t.Fatalf("migrationSourceURL() error = %v", err)
	}

	wantSuffix := filepath.ToSlash(dir)
	if !strings.HasPrefix(got, "file://") {
		t.Fatalf("migrationSourceURL() = %q, want file:// prefix", got)
	}
	if !strings.HasSuffix(got, wantSuffix) {
		t.Fatalf("migrationSourceURL() = %q, want suffix %q", got, wantSuffix)
	}
}
