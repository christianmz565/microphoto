package worker

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"log"
	"strconv"
	"time"

	"github.com/christianmz565/microphoto/pkg/client/metrics"
	"github.com/christianmz565/microphoto/pkg/client/minio"
	"github.com/christianmz565/microphoto/pkg/client/redis"
	"github.com/christianmz565/microphoto/pkg/model"
	"github.com/christianmz565/microphoto/proto/jobs"
)

const (
	BucketName = "microphoto"
)

// Processor handles the image processing tasks.
type Processor struct {
	redis    *redis.Client
	minio    *minio.Client
	metrics  *metrics.Metrics
	workerID string
}

// NewProcessor creates a new Processor instance.
func NewProcessor(r *redis.Client, m *minio.Client, mt *metrics.Metrics, workerID string) *Processor {
	return &Processor{
		redis:    r,
		minio:    m,
		metrics:  mt,
		workerID: workerID,
	}
}

// HandleJob dispatches the job to the appropriate handler based on its type.
func (p *Processor) HandleJob(ctx context.Context, job *jobs.Job) error {
	startTime := time.Now()
	defer func() {
		p.metrics.RecordTaskDuration(ctx, p.workerID, time.Since(startTime).Seconds())
		p.metrics.RecordTaskProcessed(ctx, p.workerID, job.Type.String())
	}()

	switch job.Type {
	case jobs.JobType_RECONSTRUCT:
		return p.handleReconstruct(ctx, job)
	default:
		return p.handleProcess(ctx, job)
	}
}

func (p *Processor) handleProcess(ctx context.Context, job *jobs.Job) error {
	p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:   job.ParentId,
		Status:  "START_SUBTASK",
		Message: fmt.Sprintf("Worker %s started subtask %s", p.workerID, job.Id),
	})

	reader, err := p.minio.DownloadObject(ctx, BucketName, job.OriginalImagePath)
	if err != nil {
		return fmt.Errorf("download original: %w", err)
	}
	defer reader.Close()

	img, _, err := image.Decode(reader)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}

	rect := image.Rect(
		int(job.Region.X),
		int(job.Region.Y),
		int(job.Region.X+job.Region.Width),
		int(job.Region.Y+job.Region.Height),
	)

	subImg := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(subImg, subImg.Bounds(), img, rect.Min, draw.Src)

	var processed image.Image
	switch job.Type {
	case jobs.JobType_GRAYSCALE:
		processed = ApplyGrayscale(subImg)
	case jobs.JobType_BLUR:
		processed = ApplyBlur(subImg)
	default:
		processed = subImg
	}

	index := job.Parameters["index"]
	resultPath := fmt.Sprintf("%s/res_sub_%s.png", job.ParentId, index)

	var buf bytes.Buffer
	if err := png.Encode(&buf, processed); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}

	_, err = p.minio.UploadObject(ctx, BucketName, resultPath, &buf, int64(buf.Len()), "image/png")
	if err != nil {
		return fmt.Errorf("upload processed subtask: %w", err)
	}

	p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:   job.ParentId,
		Status:  "END_SUBTASK",
		Message: fmt.Sprintf("Worker %s finished subtask %s", p.workerID, job.Id),
	})

	count, err := p.redis.DecrementCounter(ctx, job.ParentId)
	if err != nil {
		return fmt.Errorf("decrement counter: %w", err)
	}

	log.Printf("Task %s: subtasks remaining: %d", job.ParentId, count)

	if count == 0 {
		reconstructJob := &jobs.Job{
			Id:                fmt.Sprintf("recon-%s", job.ParentId),
			Type:              jobs.JobType_RECONSTRUCT,
			ParentId:          job.ParentId,
			OriginalImagePath: job.OriginalImagePath,
			CreatedAt:         time.Now().Unix(),
			Parameters:        job.Parameters,
		}

		err = p.redis.PushTask(ctx, "tasks", reconstructJob)
		if err != nil {
			return fmt.Errorf("push reconstruct task: %w", err)
		}
		log.Printf("Triggered reconstruction for task %s", job.ParentId)
	}

	return nil
}

func (p *Processor) handleReconstruct(ctx context.Context, job *jobs.Job) error {
	log.Printf("Starting reconstruction for task %s", job.ParentId)

	totalSubtasks, _ := strconv.Atoi(job.Parameters["total_subtasks"])
	originalHeight, _ := strconv.Atoi(job.Parameters["original_height"])
	originalWidth, _ := strconv.Atoi(job.Parameters["original_width"])

	if originalWidth == 0 || originalHeight == 0 || totalSubtasks == 0 {
		return fmt.Errorf("invalid job parameters: W=%d, H=%d, N=%d", originalWidth, originalHeight, totalSubtasks)
	}

	finalImg := image.NewRGBA(image.Rect(0, 0, originalWidth, originalHeight))
	rowsPerSubtask := originalHeight / totalSubtasks
	if rowsPerSubtask <= 0 {
		rowsPerSubtask = 1
	}

	for i := range totalSubtasks {
		path := fmt.Sprintf("%s/res_sub_%d.png", job.ParentId, i)
		subReader, err := p.minio.DownloadObject(ctx, BucketName, path)
		if err != nil {
			return fmt.Errorf("download subtask %d: %w", i, err)
		}

		subImg, _, err := image.Decode(subReader)
		subReader.Close()
		if err != nil {
			return fmt.Errorf("decode subtask %d: %w", i, err)
		}

		startY := i * rowsPerSubtask
		draw.Draw(finalImg, image.Rect(0, startY, originalWidth, startY+subImg.Bounds().Dy()), subImg, image.Point{}, draw.Src)
	}

	finalPath := fmt.Sprintf("%s/final.png", job.ParentId)
	var buf bytes.Buffer
	if err := png.Encode(&buf, finalImg); err != nil {
		return fmt.Errorf("encode final png: %w", err)
	}

	_, err := p.minio.UploadObject(ctx, BucketName, finalPath, &buf, int64(buf.Len()), "image/png")
	if err != nil {
		return fmt.Errorf("upload final image: %w", err)
	}

	p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:     job.ParentId,
		Status:    "JOB_COMPLETED",
		Message:   "Image processing completed successfully",
		ResultURL: finalPath,
	})

	log.Printf("Completed reconstruction for task %s", job.ParentId)
	return nil
}
