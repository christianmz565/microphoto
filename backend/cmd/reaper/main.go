// Package main is the entry point for the reaper service.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/christianmz565/microphoto/internal/reaper"
	"github.com/christianmz565/microphoto/pkg/client/metrics"
	"github.com/christianmz565/microphoto/pkg/client/redis"
	"github.com/christianmz565/microphoto/pkg/model"
	jobs "github.com/christianmz565/microphoto/proto/jobs/v1"
	"google.golang.org/protobuf/proto"
)

const inProgressPrefix = `{"global"}:in_progress:`

func main() {
	fmt.Println("Microphoto Reaper starting...")

	cfg := reaper.NewConfig()

	m, err := metrics.InitMetrics("reaper")
	if err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}

	metrics.StartMetricsServer(cfg.MetricsPort)

	rClient, err := redis.NewClient(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
	}
	defer func() { _ = rClient.Close() }()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(time.Duration(cfg.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "reaper-unknown"
	}

	log.Printf("Reaper active. Interval: %ds, Timeout: %ds", cfg.IntervalSeconds, cfg.GlobalTimeoutSeconds)

	for {
		select {
		case <-ctx.Done():
			log.Println("Reaper shutting down...")
			return
		case <-ticker.C:
			runReaperCycle(ctx, rClient, m, cfg, hostname)
		}
	}
}

func runReaperCycle(ctx context.Context, rClient *redis.Client, m *metrics.Metrics, cfg *reaper.Config, hostname string) {
	keys, err := rClient.ScanInProgressKeys(ctx, `{"global"}:in_progress:*`)
	if err != nil {
		log.Printf("Error scanning keys: %v", err)
		return
	}

	for _, key := range keys {
		progressID := extractProgressID(key)
		if progressID == "" {
			continue
		}

		items, err := rClient.GetListItems(ctx, key)
		if err != nil {
			log.Printf("Error getting items for key %s: %v", key, err)
			continue
		}

		for _, item := range items {
			job := &jobs.Job{}
			if err := proto.Unmarshal(item, job); err != nil {
				log.Printf("Error unmarshaling job from %s: %v", key, err)
				continue
			}

			now := time.Now().Unix()
			if now-job.Timestamp > cfg.GlobalTimeoutSeconds {
				if err := handleTimeout(ctx, rClient, m, job, item, progressID, hostname); err != nil {
					log.Printf("Error handling timeout for job %s: %v", job.Id, err)
				}
			}
		}
	}
}

func handleTimeout(ctx context.Context, rClient *redis.Client, m *metrics.Metrics, job *jobs.Job, rawData []byte, progressID, hostname string) error {
	taskID := job.ParentId
	if taskID == "" {
		taskID = job.Id
	}

	log.Printf("Detected timeout for job %s (task %s, worker %s)", job.Id, taskID, progressID)

	attempts, err := rClient.GetAttempts(ctx, taskID)
	if err != nil {
		return fmt.Errorf("getting attempts for task %s: %w", taskID, err)
	}

	if attempts > 0 {
		return rescheduleJob(ctx, rClient, m, job, rawData, progressID, taskID, hostname)
	}

	return failJob(ctx, rClient, m, job, rawData, progressID, taskID, hostname)
}

func rescheduleJob(ctx context.Context, rClient *redis.Client, m *metrics.Metrics, job *jobs.Job, rawData []byte, progressID, taskID, hostname string) error {
	log.Printf("Rescheduling job %s. Attempts left: %d", job.Id, job.Attempts-1)

	job.Timestamp = time.Now().Unix()

	newData, err := proto.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshaling job %s for rescheduling: %w", job.Id, err)
	}

	if err := rClient.RescheduleTask(ctx, progressID, taskID, rawData, newData); err != nil {
		return fmt.Errorf("rescheduling task %s: %w", taskID, err)
	}

	m.RecordTaskTimeout(ctx, hostname)

	return nil
}

func failJob(ctx context.Context, rClient *redis.Client, m *metrics.Metrics, job *jobs.Job, rawData []byte, progressID, taskID, hostname string) error {
	log.Printf("Job %s reached max attempts. Failing task %s", job.Id, taskID)

	if err := rClient.CleanupFailedTask(ctx, progressID, taskID, rawData); err != nil {
		log.Printf("Error cleaning up failed task %s: %v", taskID, err)
	}

	payload := model.ProgressPayload{
		JobID:   taskID,
		Status:  "JOB_FAILED",
		Message: "Max retry attempts reached or worker timeout",
	}
	if err := rClient.PublishProgress(ctx, taskID, payload); err != nil {
		log.Printf("Error publishing failure for task %s: %v", taskID, err)
	}

	m.RecordTaskTimeout(ctx, hostname)

	return nil
}

func extractProgressID(key string) string {
	if strings.HasPrefix(key, inProgressPrefix) {
		return key[len(inProgressPrefix):]
	}

	return ""
}
