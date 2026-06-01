package main

import (
	"fmt"
	"log"
	"time"

	"github.com/christianmz565/microphoto/pkg/client/metrics"
)

func main() {
	fmt.Println("Microphoto Worker starting...")

	m, err := metrics.InitMetrics("worker")
	if err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}
	_ = m // Use metrics later

	metrics.StartMetricsServer(9092)

	// Worker loop
	for {
		fmt.Println("Worker waiting for jobs...")
		time.Sleep(10 * time.Second)
	}
}
