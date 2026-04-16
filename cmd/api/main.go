package main

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/drwgrc/multi-tenant-event-pipeline/internal/config"
	"github.com/drwgrc/multi-tenant-event-pipeline/internal/httpapi"
	"github.com/drwgrc/multi-tenant-event-pipeline/internal/middleware"
	"github.com/drwgrc/multi-tenant-event-pipeline/internal/observability"
	_ "github.com/lib/pq"
)

const readinessTimeout = 2 * time.Second

func main() {
	logger := observability.NewLogger("api")

	cfg, err := config.LoadAPI()
	if err != nil {
		logger.Error("load api config", slog.Any("error", err))
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	redisChecker, err := httpapi.NewRedisChecker(cfg.RedisURL)
	if err != nil {
		logger.Error("create redis readiness checker", slog.Any("error", err))
		os.Exit(1)
	}

	handler := newHandler(logger, httpapi.NewSQLChecker(db), redisChecker)
	server := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: handler,
	}

	logger.Info("api listening", slog.String("http_addr", cfg.HTTPAddr))

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("api server exited", slog.Any("error", err))
		os.Exit(1)
	}
}

func newHandler(logger *slog.Logger, databaseChecker httpapi.Checker, redisChecker httpapi.Checker) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/livez", httpapi.NewLivenessHandler())
	mux.Handle("/readyz", httpapi.NewReadinessHandler(readinessTimeout, databaseChecker, redisChecker))

	return middleware.RequestID(logger)(middleware.RequestLogging(logger)(mux))
}
