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
	"github.com/christianmz565/microphoto/proto/jobs"
	"google.golang.org/protobuf/proto"
)

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
	defer rClient.Close()

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
				taskID := job.ParentId
				if taskID == "" {
					taskID = job.Id // fallback for jobs without parent
				}
				log.Printf("Detected timeout for job %s (task %s, worker %s)", job.Id, taskID, progressID)

				attempts, err := rClient.GetAttempts(ctx, taskID)
				if err != nil {
					log.Printf("Error getting attempts for task %s: %v", taskID, err)
					continue
				}

				if attempts > 0 {
					log.Printf("Rescheduling job %s. Attempts left: %d", job.Id, attempts-1)

					job.Timestamp = time.Now().Unix()
					newData, err := proto.Marshal(job)
					if err != nil {
						log.Printf("Error marshaling job %s for rescheduling: %v", job.Id, err)
						continue
					}

					if err := rClient.RescheduleTask(ctx, progressID, taskID, item, newData); err != nil {
						log.Printf("Error rescheduling task %s: %v", taskID, err)
						continue
					}

					m.RecordTaskTimeout(ctx, hostname)
				} else {
					log.Printf("Job %s reached max attempts. Failing task %s", job.Id, taskID)

					if err := rClient.CleanupFailedTask(ctx, progressID, taskID, item); err != nil {
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
				}
			}
		}
	}
}

func extractProgressID(key string) string {
	const prefix = `{"global"}:in_progress:`
	if strings.HasPrefix(key, prefix) {
		return key[len(prefix):]
	}
	return ""
}
