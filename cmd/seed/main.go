package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/drwgrc/multi-tenant-event-pipeline/internal/observability"
	_ "github.com/lib/pq"
)

const (
	defaultSeedTimeout = 10 * time.Second

	demoTenantSlug    = "demo"
	demoTenantName    = "Demo Tenant"
	demoAdminEmail    = "admin@demo.local"
	demoAdminPassword = "demo-admin-password"
	demoSourceName    = "demo-web"
	demoAPIKeyName    = "demo-ingest"
	demoAPIKeySecret  = "evt_demo_local_seed_7e9f6b4c2a1d"
	adminRole         = "admin"
)

type seedConfig struct {
	AppEnv        string
	DatabaseURL   string
	ForceNonLocal bool
}

type seedResult struct {
	TenantID     string
	UserID       string
	SourceID     string
	APIKeyID     string
	TenantSlug   string
	TenantName   string
	AdminEmail   string
	AdminRole    string
	SourceName   string
	APIKeyName   string
	APIKeyPrefix string
	APIKeySecret string
}

func main() {
	logger := observability.NewLogger("seed")

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("load seed config", slog.Any("error", err))
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), defaultSeedTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Error("ping database", slog.Any("error", err))
		os.Exit(1)
	}

	result, err := runSeed(ctx, db)
	if err != nil {
		logger.Error("seed database", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info(
		"seed complete",
		slog.String("tenant_id", result.TenantID),
		slog.String("user_id", result.UserID),
		slog.String("source_id", result.SourceID),
		slog.String("api_key_id", result.APIKeyID),
	)

	printResult(result)
}

func loadConfig() (seedConfig, error) {
	cfg := seedConfig{
		AppEnv:        strings.TrimSpace(os.Getenv("APP_ENV")),
		DatabaseURL:   strings.TrimSpace(os.Getenv("DATABASE_URL")),
		ForceNonLocal: strings.EqualFold(strings.TrimSpace(os.Getenv("SEED_ALLOW_NON_LOCAL_DATABASE")), "true"),
	}

	if cfg.AppEnv == "" {
		cfg.AppEnv = "development"
	}

	if cfg.DatabaseURL == "" {
		return seedConfig{}, errors.New("DATABASE_URL is required")
	}

	parsed, err := url.Parse(cfg.DatabaseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return seedConfig{}, errors.New("DATABASE_URL must be a valid URL")
	}

	if err := validateSeedTarget(cfg, parsed); err != nil {
		return seedConfig{}, err
	}

	return cfg, nil
}

func validateSeedTarget(cfg seedConfig, parsed *url.URL) error {
	if cfg.ForceNonLocal {
		return nil
	}

	if !isSafeSeedAppEnv(cfg.AppEnv) {
		return fmt.Errorf("refusing to seed when APP_ENV=%q; set SEED_ALLOW_NON_LOCAL_DATABASE=true to override intentionally", cfg.AppEnv)
	}

	host := parsed.Hostname()
	if !isSafeSeedHost(host) {
		return fmt.Errorf("refusing to seed non-local database host %q; set SEED_ALLOW_NON_LOCAL_DATABASE=true to override intentionally", host)
	}

	return nil
}

func isSafeSeedAppEnv(appEnv string) bool {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "development", "dev", "local", "test":
		return true
	default:
		return false
	}
}

func isSafeSeedHost(host string) bool {
	switch strings.ToLower(strings.TrimSpace(host)) {
	case "localhost", "127.0.0.1", "::1", "postgres":
		return true
	default:
		return false
	}
}

func runSeed(ctx context.Context, db *sql.DB) (seedResult, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return seedResult{}, fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result := seedResult{
		TenantSlug:   demoTenantSlug,
		TenantName:   demoTenantName,
		AdminEmail:   demoAdminEmail,
		AdminRole:    adminRole,
		SourceName:   demoSourceName,
		APIKeyName:   demoAPIKeyName,
		APIKeyPrefix: apiKeyPrefix(demoAPIKeySecret),
		APIKeySecret: demoAPIKeySecret,
	}

	if err = upsertTenant(ctx, tx, &result); err != nil {
		return seedResult{}, err
	}

	if err = upsertUser(ctx, tx, &result); err != nil {
		return seedResult{}, err
	}

	if err = upsertMembership(ctx, tx, result); err != nil {
		return seedResult{}, err
	}

	if err = upsertSource(ctx, tx, &result); err != nil {
		return seedResult{}, err
	}

	if err = upsertAPIKey(ctx, tx, &result); err != nil {
		return seedResult{}, err
	}

	if err = tx.Commit(); err != nil {
		return seedResult{}, fmt.Errorf("commit transaction: %w", err)
	}

	return result, nil
}

func upsertTenant(ctx context.Context, tx *sql.Tx, result *seedResult) error {
	err := tx.QueryRowContext(
		ctx,
		`INSERT INTO tenants (slug, name)
		VALUES ($1, $2)
		ON CONFLICT (slug) DO UPDATE
		SET name = EXCLUDED.name,
		    updated_at = NOW()
		RETURNING id`,
		result.TenantSlug,
		result.TenantName,
	).Scan(&result.TenantID)
	if err != nil {
		return fmt.Errorf("upsert tenant: %w", err)
	}

	return nil
}

func upsertUser(ctx context.Context, tx *sql.Tx, result *seedResult) error {
	err := tx.QueryRowContext(
		ctx,
		`INSERT INTO users (email, password_hash)
		VALUES ($1, crypt($2, gen_salt('bf')))
		ON CONFLICT ((LOWER(email))) DO UPDATE
		SET email = EXCLUDED.email,
		    password_hash = crypt($2, gen_salt('bf')),
		    updated_at = NOW()
		RETURNING id`,
		result.AdminEmail,
		demoAdminPassword,
	).Scan(&result.UserID)
	if err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}

	return nil
}

func upsertMembership(ctx context.Context, tx *sql.Tx, result seedResult) error {
	_, err := tx.ExecContext(
		ctx,
		`INSERT INTO memberships (tenant_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (tenant_id, user_id) DO UPDATE
		SET role = EXCLUDED.role`,
		result.TenantID,
		result.UserID,
		result.AdminRole,
	)
	if err != nil {
		return fmt.Errorf("upsert membership: %w", err)
	}

	return nil
}

func upsertSource(ctx context.Context, tx *sql.Tx, result *seedResult) error {
	err := tx.QueryRowContext(
		ctx,
		`INSERT INTO sources (tenant_id, name)
		VALUES ($1, $2)
		ON CONFLICT (tenant_id, name) DO UPDATE
		SET updated_at = NOW()
		RETURNING id`,
		result.TenantID,
		result.SourceName,
	).Scan(&result.SourceID)
	if err != nil {
		return fmt.Errorf("upsert source: %w", err)
	}

	return nil
}

func upsertAPIKey(ctx context.Context, tx *sql.Tx, result *seedResult) error {
	err := tx.QueryRowContext(
		ctx,
		`INSERT INTO api_keys (tenant_id, name, key_prefix, key_hash, revoked_at)
		VALUES ($1, $2, $3, $4, NULL)
		ON CONFLICT (tenant_id, name) DO UPDATE
		SET key_prefix = EXCLUDED.key_prefix,
		    key_hash = EXCLUDED.key_hash,
		    revoked_at = NULL
		RETURNING id`,
		result.TenantID,
		result.APIKeyName,
		result.APIKeyPrefix,
		apiKeyHash(result.APIKeySecret),
	).Scan(&result.APIKeyID)
	if err != nil {
		return fmt.Errorf("upsert api key: %w", err)
	}

	return nil
}

func apiKeyPrefix(secret string) string {
	const prefixLength = 12
	if len(secret) <= prefixLength {
		return secret
	}

	return secret[:prefixLength]
}

func apiKeyHash(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func printResult(result seedResult) {
	fmt.Printf(`seed complete
tenant_slug=%s
tenant_name=%s
admin_email=%s
admin_password=%s
admin_role=%s
source_name=%s
api_key_name=%s
api_key_prefix=%s
api_key_secret=%s
note=demo-only local bootstrap credentials
`, result.TenantSlug, result.TenantName, result.AdminEmail, demoAdminPassword, result.AdminRole, result.SourceName, result.APIKeyName, result.APIKeyPrefix, result.APIKeySecret)
}
