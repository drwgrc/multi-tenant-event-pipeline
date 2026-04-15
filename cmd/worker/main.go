package main

import (
	"log"
	"time"

	"github.com/drwgrc/multi-tenant-event-pipeline/internal/config"
)

func main() {
	if _, err := config.LoadWorker(); err != nil {
		log.Fatalf("load worker config: %v", err)
	}

	log.Println("worker started")
	for {
		log.Println("worker tick")
		time.Sleep(5 * time.Second)
	}
}
