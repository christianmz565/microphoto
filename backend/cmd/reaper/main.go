package main

import (
	"fmt"
	"log"
	"time"

	"github.com/christianmz565/microphoto/pkg/client/metrics"
)

func main() {
	fmt.Println("Microphoto Reaper starting...")

	m, err := metrics.InitMetrics("reaper")
	if err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}
	_ = m // Use metrics later

	metrics.StartMetricsServer(9091)

	// Reaper loop
	for {
		fmt.Println("Reaper checking for stale jobs...")
		time.Sleep(30 * time.Second)
	}
}
