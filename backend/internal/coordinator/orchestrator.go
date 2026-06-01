package coordinator

import (
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"time"

	"github.com/christianmz565/microphoto/pkg/client/metrics"
	"github.com/christianmz565/microphoto/pkg/client/minio"
	"github.com/christianmz565/microphoto/pkg/client/redis"
	"github.com/christianmz565/microphoto/proto/jobs"
	"github.com/google/uuid"
)

const (
	MaxPixelsPerSubtask = 1000000
	DefaultAttempts     = 3
	BucketName          = "microphoto"
)

type Orchestrator struct {
	redis   *redis.Client
	minio   *minio.Client
	metrics *metrics.Metrics
}

func NewOrchestrator(r *redis.Client, m *minio.Client, mt *metrics.Metrics) *Orchestrator {
	return &Orchestrator{
		redis:   r,
		minio:   m,
		metrics: mt,
	}
}

func (o *Orchestrator) ProcessImage(ctx context.Context, file io.ReadSeeker, filename string, jobType jobs.JobType, size int64) (string, error) {
	startTime := time.Now()
	taskID := uuid.New().String()

	cfg, _, err := image.DecodeConfig(file)
	if err != nil {
		return "", fmt.Errorf("decode image config: %w", err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("seek to start: %w", err)
	}

	path := fmt.Sprintf("%s/original.png", taskID)
	_, err = o.minio.UploadObject(ctx, BucketName, path, file, size, "image/png")
	if err != nil {
		return "", fmt.Errorf("upload to minio: %w", err)
	}

	W, H := cfg.Width, cfg.Height
	totalPixels := W * H
	N := int(math.Ceil(float64(totalPixels) / float64(MaxPixelsPerSubtask)))
	if N <= 0 {
		N = 1
	}

	rowsPerSubtask := H / N
	if rowsPerSubtask <= 0 {
		rowsPerSubtask = 1
		N = H
	}

	err = o.redis.InitializeTask(ctx, taskID, N, DefaultAttempts)
	if err != nil {
		return "", fmt.Errorf("initialize redis task: %w", err)
	}

	for i := 0; i < N; i++ {
		startY := i * rowsPerSubtask
		endY := (i + 1) * rowsPerSubtask
		if i == N-1 {
			endY = H
		}

		subJobID := uuid.New().String()
		job := &jobs.Job{
			Id:                subJobID,
			Type:              jobType,
			Status:            jobs.JobStatus_PENDING,
			OriginalImagePath: path,
			ParentId:          taskID,
			Region: &jobs.Region{
				X:      0,
				Y:      int32(startY),
				Width:  int32(W),
				Height: int32(endY - startY),
			},
			Attempts:  DefaultAttempts,
			CreatedAt: time.Now().Unix(),
		}

		err = o.redis.PushTask(ctx, taskID, job)
		if err != nil {
			return "", fmt.Errorf("push task: %w", err)
		}
	}

	o.metrics.RecordTaskDuration(ctx, "coordinator", time.Since(startTime).Seconds())
	o.metrics.RecordTaskProcessed(ctx, "coordinator", jobType.String())

	return taskID, nil
}
