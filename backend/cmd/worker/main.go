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
	workerID, _ := os.Hostname()
	if workerID == "" {
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
	defer rClient.Close()

	mClient, err := minio.NewClient(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioSSL)
	if err != nil {
		log.Fatalf("Failed to initialize minio: %v", err)
	}

	processor := worker.NewProcessor(rClient, mClient, m, workerID)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	fmt.Println("Worker ready to process tasks...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker shutting down gracefully")
			return
		default:
			job, err := rClient.PopTaskReliable(ctx, "tasks")
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
			log.Printf("Finished job %s", job.Id)
		}
	}
}
