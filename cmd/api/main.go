package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"github.com/drwgrc/multi-tenant-event-pipeline/internal/config"
	"github.com/drwgrc/multi-tenant-event-pipeline/internal/middleware"
	"github.com/drwgrc/multi-tenant-event-pipeline/internal/observability"
)

func main() {
	logger := observability.NewLogger("api")

	cfg, err := config.LoadAPI()
	if err != nil {
		logger.Error("load api config", slog.Any("error", err))
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := middleware.RequestID(logger)(middleware.RequestLogging(logger)(mux))
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
