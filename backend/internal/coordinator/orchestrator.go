package coordinator

import (
	"context"
	"fmt"
	"io"
	"maps"
	"time"

	"github.com/christianmz565/microphoto/pkg/client/metrics"
	"github.com/christianmz565/microphoto/pkg/client/minio"
	"github.com/christianmz565/microphoto/pkg/client/redis"
	jobs "github.com/christianmz565/microphoto/proto/jobs/v1"
	"github.com/google/uuid"
)

const (
	BucketName = "microphoto"
)

// Orchestrator coordinates the image processing tasks by splitting images into subtasks and managing their lifecycle.
type Orchestrator struct {
	redis   *redis.Client
	minio   *minio.Client
	metrics *metrics.Metrics
}

// NewOrchestrator creates a new Orchestrator instance.
func NewOrchestrator(r *redis.Client, m *minio.Client, mt *metrics.Metrics) *Orchestrator {
	return &Orchestrator{
		redis:   r,
		minio:   m,
		metrics: mt,
	}
}

// ProcessImage handles the initial image upload and pushes a SLICE task to Redis.
func (o *Orchestrator) ProcessImage(ctx context.Context, taskID string, file io.Reader, filename string, jobType jobs.JobType, size int64, params map[string]string) error {
	startTime := time.Now()

	path := fmt.Sprintf("%s/original.png", taskID)
	_, err := o.minio.UploadObject(ctx, BucketName, path, file, size, "image/png")
	if err != nil {
		return fmt.Errorf("upload to minio: %w", err)
	}

	jobParams := map[string]string{
		"target_type": jobType.String(),
	}
	maps.Copy(jobParams, params)

	sliceJob := &jobs.Job{
		Id:                uuid.New().String(),
		Type:              jobs.JobType_JOB_TYPE_SLICE,
		Status:            jobs.JobStatus_JOB_STATUS_UNSPECIFIED,
		OriginalImagePath: path,
		ParentId:          taskID,
		CreatedAt:         time.Now().Unix(),
		Timestamp:         time.Now().Unix(),
		Parameters:        jobParams,
	}

	err = o.redis.PushTask(ctx, sliceJob)
	if err != nil {
		return fmt.Errorf("push slice task: %w", err)
	}

	o.metrics.RecordTaskDuration(ctx, "coordinator", time.Since(startTime).Seconds())
	o.metrics.RecordTaskProcessed(ctx, "coordinator", jobType.String())

	return nil
}

// DownloadResult downloads the final processed image from MinIO.
func (o *Orchestrator) DownloadResult(ctx context.Context, taskID string) (io.ReadCloser, error) {
	path := fmt.Sprintf("%s/final.png", taskID)
	return o.minio.DownloadObject(ctx, BucketName, path)
}
