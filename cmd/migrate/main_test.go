package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantCommand string
		wantErr     string
	}{
		{
			name:        "defaults to up with no command",
			args:        []string{"migrate"},
			wantCommand: "up",
		},
		{
			name:        "accepts up",
			args:        []string{"migrate", "up"},
			wantCommand: "up",
		},
		{
			name:        "accepts down",
			args:        []string{"migrate", "down"},
			wantCommand: "down",
		},
		{
			name:        "accepts version",
			args:        []string{"migrate", "version"},
			wantCommand: "version",
		},
		{
			name:    "rejects unknown command",
			args:    []string{"migrate", "status"},
			wantErr: `unknown migration command "status"`,
		},
		{
			name:    "rejects down with step count",
			args:    []string{"migrate", "down", "1"},
			wantErr: "unexpected trailing arguments: 1",
		},
		{
			name:    "rejects down with typo",
			args:    []string{"migrate", "down", "typo"},
			wantErr: "unexpected trailing arguments: typo",
		},
		{
			name:    "rejects version with extra arg",
			args:    []string{"migrate", "version", "extra"},
			wantErr: "unexpected trailing arguments: extra",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCommand(tt.args)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("parseCommand() error = nil, want %q", tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("parseCommand() error = %q, want %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("parseCommand() error = %v", err)
			}
			if got != tt.wantCommand {
				t.Fatalf("parseCommand() = %q, want %q", got, tt.wantCommand)
			}
		})
	}
}

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
