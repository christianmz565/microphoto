package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/christianmz565/microphoto/pkg/client/metrics"
)

func main() {
	fmt.Println("Microphoto Coordinator starting...")

	m, err := metrics.InitMetrics("coordinator")
	if err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}
	_ = m // Use metrics later

	metrics.StartMetricsServer(9090)

	// Simple health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
