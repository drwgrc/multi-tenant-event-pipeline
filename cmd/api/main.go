package main

import (
	"log"
	"net/http"

	"github.com/drwgrc/multi-tenant-event-pipeline/internal/config"
)

func main() {
	cfg, err := config.LoadAPI()
	if err != nil {
		log.Fatalf("load api config: %v", err)
	}

	http.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Printf("api listening on %s", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, nil))
}
