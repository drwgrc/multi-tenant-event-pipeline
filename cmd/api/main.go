package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/drwgrc/multi-tenant-event-pipeline/internal/config"
	"github.com/drwgrc/multi-tenant-event-pipeline/internal/httpapi"
	"github.com/drwgrc/multi-tenant-event-pipeline/internal/middleware"
	"github.com/drwgrc/multi-tenant-event-pipeline/internal/observability"
	_ "github.com/lib/pq"
)

const (
	readinessTimeout  = 2 * time.Second
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

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
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("api listening", slog.String("http_addr", cfg.HTTPAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	select {
	case <-ctx.Done():
		logger.Info("api shutdown signal received")
	case err := <-serverErr:
		if err != nil {
			logger.Error("api server exited", slog.Any("error", err))
			os.Exit(1)
		}
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("api graceful shutdown failed", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("api shutdown complete")
}

func newHandler(logger *slog.Logger, databaseChecker httpapi.Checker, redisChecker httpapi.Checker) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/livez", httpapi.NewLivenessHandler())
	mux.Handle("/readyz", httpapi.NewReadinessHandler(readinessTimeout, databaseChecker, redisChecker))

	return middleware.RequestID(logger)(middleware.RequestLogging(logger)(mux))
}
