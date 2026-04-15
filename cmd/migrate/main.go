package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"

	"github.com/drwgrc/multi-tenant-event-pipeline/internal/observability"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const defaultMigrationsDir = "migrations"

type migrateConfig struct {
	DatabaseURL   string
	MigrationsDir string
}

func main() {
	logger := observability.NewLogger("migrate")

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("load migration config", slog.Any("error", err))
		os.Exit(1)
	}

	sourceURL, err := migrationSourceURL(cfg.MigrationsDir)
	if err != nil {
		logger.Error("resolve migrations directory", slog.Any("error", err))
		os.Exit(1)
	}

	runner, err := migrate.New(sourceURL, cfg.DatabaseURL)
	if err != nil {
		logger.Error("create migration runner", slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		sourceErr, databaseErr := runner.Close()
		if sourceErr != nil {
			logger.Error("close migration source", slog.Any("error", sourceErr))
		}
		if databaseErr != nil {
			logger.Error("close migration database", slog.Any("error", databaseErr))
		}
	}()

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "up":
		if err := runUp(logger, runner); err != nil {
			logger.Error("apply up migrations", slog.Any("error", err))
			os.Exit(1)
		}
	case "down":
		if err := runDown(logger, runner); err != nil {
			logger.Error("apply down migrations", slog.Any("error", err))
			os.Exit(1)
		}
	case "version":
		if err := printVersion(runner); err != nil {
			logger.Error("read migration version", slog.Any("error", err))
			os.Exit(1)
		}
	default:
		logger.Error("unknown migration command", slog.String("command", command))
		fmt.Fprintln(os.Stderr, "usage: go run ./cmd/migrate [up|down|version]")
		os.Exit(2)
	}
}

func loadConfig() (migrateConfig, error) {
	cfg := migrateConfig{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		MigrationsDir: os.Getenv("MIGRATIONS_DIR"),
	}

	if cfg.DatabaseURL == "" {
		return migrateConfig{}, errors.New("DATABASE_URL is required")
	}

	parsed, err := url.Parse(cfg.DatabaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return migrateConfig{}, errors.New("DATABASE_URL must be a valid URL")
	}

	if cfg.MigrationsDir == "" {
		cfg.MigrationsDir = defaultMigrationsDir
	}

	return cfg, nil
}

func migrationSourceURL(migrationsDir string) (string, error) {
	absDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		return "", fmt.Errorf("resolve migrations dir: %w", err)
	}

	info, err := os.Stat(absDir)
	if err != nil {
		return "", fmt.Errorf("stat migrations dir: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("migrations path %q is not a directory", absDir)
	}

	return (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(absDir),
	}).String(), nil
}

func runUp(logger *slog.Logger, runner *migrate.Migrate) error {
	err := runner.Up()
	if errors.Is(err, migrate.ErrNoChange) {
		logger.Info("migrations already up to date")
		return nil
	}
	if err != nil {
		return err
	}

	logger.Info("applied up migrations")
	return nil
}

func runDown(logger *slog.Logger, runner *migrate.Migrate) error {
	err := runner.Down()
	if errors.Is(err, migrate.ErrNoChange) {
		logger.Info("database already at base migration state")
		return nil
	}
	if err != nil {
		return err
	}

	logger.Info("reverted all migrations")
	return nil
}

func printVersion(runner *migrate.Migrate) error {
	version, dirty, err := runner.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		fmt.Println("version=none dirty=false")
		return nil
	}
	if err != nil {
		return err
	}

	fmt.Printf("version=%d dirty=%t\n", version, dirty)
	return nil
}
