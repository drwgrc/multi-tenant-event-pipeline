package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/drwgrc/multi-tenant-event-pipeline/internal/config"
	"github.com/drwgrc/multi-tenant-event-pipeline/internal/observability"
)

func main() {
	workerID := observability.NewCorrelationID()
	logger := observability.NewLogger("worker").With(observability.WorkerIDAttr(workerID))

	if _, err := config.LoadWorker(); err != nil {
		logger.Error("load worker config", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("worker started")
	for {
		logger.Info("worker tick")
		time.Sleep(5 * time.Second)
	}
}
