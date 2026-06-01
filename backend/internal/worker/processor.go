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
	"log"
	"maps"
	"math"
	"strconv"
	"time"

	"github.com/christianmz565/microphoto/pkg/client/metrics"
	"github.com/christianmz565/microphoto/pkg/client/minio"
	"github.com/christianmz565/microphoto/pkg/client/redis"
	"github.com/christianmz565/microphoto/pkg/model"
	jobs "github.com/christianmz565/microphoto/proto/jobs/v1"
	"github.com/google/uuid"
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
	case jobs.JobType_JOB_TYPE_SLICE:
		return p.handleSlice(ctx, job)
	case jobs.JobType_JOB_TYPE_RECONSTRUCT:
		return p.handleReconstruct(ctx, job)
	default:
		return p.handleProcess(ctx, job)
	}
}

func (p *Processor) handleSlice(ctx context.Context, job *jobs.Job) error {
	p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:   job.ParentId,
		Status:  "START_SLICING",
		Message: fmt.Sprintf("Worker %s started slicing for task %s", p.workerID, job.ParentId),
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

	W, H := img.Bounds().Dx(), img.Bounds().Dy()
	totalPixels := W * H
	const MaxPixelsPerSubtask = 1000000
	const DefaultAttempts = 3

	N := int(math.Ceil(float64(totalPixels) / float64(MaxPixelsPerSubtask)))
	if N <= 0 {
		N = 1
	}

	rowsPerSubtask := H / N
	if rowsPerSubtask <= 0 {
		rowsPerSubtask = 1
		N = H
	}

	var subtasks []*jobs.Job
	targetType := jobs.JobType(jobs.JobType_value[job.Parameters["target_type"]])

	radius := 0
	if targetType == jobs.JobType_JOB_TYPE_BLUR {
		if r, err := strconv.Atoi(job.Parameters["radius"]); err == nil {
			radius = r
		} else {
			radius = 1
		}
	}

	for i := 0; i < N; i++ {
		startY := i * rowsPerSubtask
		endY := (i + 1) * rowsPerSubtask
		if i == N-1 {
			endY = H
		}

		paddingTop := 0
		paddingBottom := 0
		if radius > 0 {
			if startY > 0 {
				paddingTop = radius
				if startY-paddingTop < 0 {
					paddingTop = startY
				}
			}
			if endY < H {
				paddingBottom = radius
				if endY+paddingBottom > H {
					paddingBottom = H - endY
				}
			}
		}

		rect := image.Rect(0, startY-paddingTop, W, endY+paddingBottom)
		subImg := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
		draw.Draw(subImg, subImg.Bounds(), img, rect.Min, draw.Src)

		fragmentPath := fmt.Sprintf("%s/sub_%d.png", job.ParentId, i)
		var buf bytes.Buffer
		if err := png.Encode(&buf, subImg); err != nil {
			return fmt.Errorf("encode fragment %d: %w", i, err)
		}

		_, err = p.minio.UploadObject(ctx, BucketName, fragmentPath, &buf, int64(buf.Len()), "image/png")
		if err != nil {
			return fmt.Errorf("upload fragment %d: %w", i, err)
		}

		subJobParams := make(map[string]string)
		maps.Copy(subJobParams, job.Parameters)
		subJobParams["index"] = fmt.Sprintf("%d", i)
		subJobParams["total_subtasks"] = fmt.Sprintf("%d", N)
		subJobParams["original_width"] = fmt.Sprintf("%d", W)
		subJobParams["original_height"] = fmt.Sprintf("%d", H)
		subJobParams["padding_top"] = fmt.Sprintf("%d", paddingTop)
		subJobParams["padding_bottom"] = fmt.Sprintf("%d", paddingBottom)

		subJobID := uuid.New().String()
		subJob := &jobs.Job{
			Id:                subJobID,
			Type:              targetType,
			Status:            jobs.JobStatus_JOB_STATUS_UNSPECIFIED,
			OriginalImagePath: fragmentPath,
			ParentId:          job.ParentId,
			Region: &jobs.Region{
				X:      0,
				Y:      0,
				Width:  int32(W),
				Height: int32(rect.Dy()),
			},
			Attempts:   DefaultAttempts,
			CreatedAt:  time.Now().Unix(),
			Timestamp:  time.Now().Unix(),
			Parameters: subJobParams,
		}
		subtasks = append(subtasks, subJob)
	}

	if err := p.redis.InitializeTask(ctx, job.ParentId, N, DefaultAttempts); err != nil {
		return fmt.Errorf("initialize redis task: %w", err)
	}

	if err := p.redis.PushTasksPipeline(ctx, subtasks); err != nil {
		return fmt.Errorf("push subtasks: %w", err)
	}

	p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:   job.ParentId,
		Status:  "END_SLICING",
		Message: fmt.Sprintf("Worker %s finished slicing for task %s", p.workerID, job.ParentId),
	})

	return nil
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
	case jobs.JobType_JOB_TYPE_GRAYSCALE:
		processed = ApplyGrayscale(subImg)
	case jobs.JobType_JOB_TYPE_BLUR:
		radius := 1
		if r, err := strconv.Atoi(job.Parameters["radius"]); err == nil {
			radius = r
		}
		processed = ApplyBlur(subImg, radius)
	case jobs.JobType_JOB_TYPE_BRIGHTNESS:
		factor := 1.0
		if f, err := strconv.ParseFloat(job.Parameters["factor"], 64); err == nil {
			factor = f
		}
		processed = ApplyBrightness(subImg, factor)
	case jobs.JobType_JOB_TYPE_RESIZE:
		targetWidth, _ := strconv.Atoi(job.Parameters["width"])
		targetHeight, _ := strconv.Atoi(job.Parameters["height"])
		originalWidth, _ := strconv.Atoi(job.Parameters["original_width"])
		originalHeight, _ := strconv.Atoi(job.Parameters["original_height"])

		if originalHeight > 0 && originalWidth > 0 && targetHeight > 0 && targetWidth > 0 {
			scaleY := float64(targetHeight) / float64(originalHeight)
			newFragHeight := int(float64(rect.Dy()) * scaleY)

			scaleX := float64(targetWidth) / float64(originalWidth)
			newFragWidth := int(float64(rect.Dx()) * scaleX)

			processed = ApplyResize(subImg, newFragWidth, newFragHeight)
		} else {
			processed = subImg
		}
	default:
		processed = subImg
	}

	paddingTop, _ := strconv.Atoi(job.Parameters["padding_top"])
	paddingBottom, _ := strconv.Atoi(job.Parameters["padding_bottom"])

	if paddingTop > 0 || paddingBottom > 0 {
		bounds := processed.Bounds()
		if job.Type == jobs.JobType_JOB_TYPE_RESIZE {
			originalHeight, _ := strconv.Atoi(job.Parameters["original_height"])
			targetHeight, _ := strconv.Atoi(job.Parameters["height"])
			scaleY := float64(targetHeight) / float64(originalHeight)
			paddingTop = int(float64(paddingTop) * scaleY)
			paddingBottom = int(float64(paddingBottom) * scaleY)
		}

		cropRect := image.Rect(0, paddingTop, bounds.Dx(), bounds.Dy()-paddingBottom)
		croppedImg := image.NewRGBA(image.Rect(0, 0, cropRect.Dx(), cropRect.Dy()))
		draw.Draw(croppedImg, croppedImg.Bounds(), processed, cropRect.Min, draw.Src)
		processed = croppedImg
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
			Type:              jobs.JobType_JOB_TYPE_RECONSTRUCT,
			ParentId:          job.ParentId,
			OriginalImagePath: job.OriginalImagePath,
			CreatedAt:         time.Now().Unix(),
			Timestamp:         time.Now().Unix(),
			Parameters:        job.Parameters,
		}

		err = p.redis.PushTask(ctx, reconstructJob)
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

	targetWidth := originalWidth
	targetHeight := originalHeight

	if w, err := strconv.Atoi(job.Parameters["width"]); err == nil && w > 0 {
		targetWidth = w
	}
	if h, err := strconv.Atoi(job.Parameters["height"]); err == nil && h > 0 {
		targetHeight = h
	}

	if originalWidth == 0 || originalHeight == 0 || totalSubtasks == 0 {
		return fmt.Errorf("invalid job parameters: W=%d, H=%d, N=%d", originalWidth, originalHeight, totalSubtasks)
	}

	finalImg := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	currentY := 0
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

		draw.Draw(finalImg, image.Rect(0, currentY, targetWidth, currentY+subImg.Bounds().Dy()), subImg, image.Point{}, draw.Src)
		currentY += subImg.Bounds().Dy()
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
