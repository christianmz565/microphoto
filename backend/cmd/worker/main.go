// Package main is the entry point for the worker service.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/christianmz565/microphoto/internal/worker"
	"github.com/christianmz565/microphoto/pkg/client/metrics"
	"github.com/christianmz565/microphoto/pkg/client/minio"
	"github.com/christianmz565/microphoto/pkg/client/redis"
)

func main() {
	workerID, err := os.Hostname()
	if err != nil || workerID == "" {
		workerID = "unknown-worker"
	}

	fmt.Printf("Microphoto Worker %s starting...\n", workerID)

	cfg := worker.NewConfig()

	m, err := metrics.InitMetrics("worker")
	if err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}

	metrics.StartMetricsServer(cfg.MetricsPort)

	rClient, err := redis.NewClient(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
	}

	mClient, err := minio.NewClient(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioSSL)
	if err != nil {
		_ = rClient.Close()

		log.Fatalf("Failed to initialize minio: %v", err)
	}
	defer func() { _ = rClient.Close() }()

	processor := worker.NewProcessor(rClient, mClient, m, workerID)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fmt.Println("Worker ready to process tasks...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker shutting down gracefully")
			return
		default:
			job, rawData, err := rClient.PopTaskReliable(ctx, workerID)
			if err != nil {
				if ctx.Err() != nil {
					return
				}

				log.Printf("Error popping task: %v", err)

				continue
			}

			log.Printf("Processing job %s (Type: %s, Parent: %s)", job.Id, job.Type, job.ParentId)

			if err := processor.HandleJob(ctx, job); err != nil {
				log.Printf("Error handling job %s: %v", job.Id, err)
				continue
			}

			if err := rClient.CompleteTask(ctx, workerID, rawData); err != nil {
				log.Printf("Error completing task %s: %v", job.Id, err)
			}

			log.Printf("Finished job %s", job.Id)
		}
	}
}
