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
	"github.com/christianmz565/microphoto/pkg/model"
	jobs "github.com/christianmz565/microphoto/proto/jobs/v1"
	"github.com/google/uuid"
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
func (o *Orchestrator) ProcessImage(ctx context.Context, taskID string, file io.Reader, _ string, jobType jobs.JobType, size int64, params map[string]string) error {
	startTime := time.Now()

	path := taskID + "/original.png"

	_, err := o.minio.UploadObject(ctx, model.BucketName, path, file, size, "image/png")
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
	path := taskID + "/final.png"
	return o.minio.DownloadObject(ctx, model.BucketName, path)
}

// DownloadVideoResult downloads the final processed video from MinIO.
func (o *Orchestrator) DownloadVideoResult(ctx context.Context, taskID string) (io.ReadCloser, error) {
	path := taskID + "/final.mp4"
	return o.minio.DownloadObject(ctx, model.BucketName, path)
}

// ProcessVideo handles video upload and pushes a VIDEO_EXTRACT task to Redis.
func (o *Orchestrator) ProcessVideo(ctx context.Context, taskID string, file io.Reader, _ string, jobType jobs.JobType, size int64, params map[string]string) error {
	startTime := time.Now()

	path := taskID + "/video.mp4"

	_, err := o.minio.UploadObject(ctx, model.BucketName, path, file, size, "video/mp4")
	if err != nil {
		return fmt.Errorf("upload to minio: %w", err)
	}

	extractJob := &jobs.Job{
		Id:                uuid.New().String(),
		Type:              jobs.JobType_JOB_TYPE_VIDEO_EXTRACT,
		Status:            jobs.JobStatus_JOB_STATUS_UNSPECIFIED,
		OriginalImagePath: path,
		ParentId:          taskID,
		CreatedAt:         time.Now().Unix(),
		Timestamp:         time.Now().Unix(),
		Parameters: map[string]string{
			"target_type": jobType.String(),
			"fps":         params["fps"],
		},
	}

	err = o.redis.PushTask(ctx, extractJob)
	if err != nil {
		return fmt.Errorf("push video extract task: %w", err)
	}

	o.metrics.RecordTaskDuration(ctx, "coordinator", time.Since(startTime).Seconds())
	o.metrics.RecordTaskProcessed(ctx, "coordinator", "VIDEO_EXTRACT")

	return nil
}
