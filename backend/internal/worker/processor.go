package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"maps"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/christianmz565/microphoto/pkg/client/metrics"
	"github.com/christianmz565/microphoto/pkg/client/minio"
	"github.com/christianmz565/microphoto/pkg/client/redis"
	"github.com/christianmz565/microphoto/pkg/model"
	jobs "github.com/christianmz565/microphoto/proto/jobs/v1"
	"github.com/google/uuid"
	"github.com/h2non/bimg"
	"golang.org/x/sync/errgroup"
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

type subTaskResult struct {
	job *jobs.Job
	err error
}

// HandleJob dispatches the job to the appropriate handler based on its type.
func (p *Processor) HandleJob(ctx context.Context, job *jobs.Job) error {
	startTime := time.Now()

	var err error

	defer func() {
		p.metrics.RecordTaskDuration(ctx, p.workerID, time.Since(startTime).Seconds())
		p.metrics.RecordTaskProcessed(ctx, p.workerID, job.Type.String())

		if err != nil {
			_ = p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
				JobID:    job.ParentId,
				WorkerID: p.workerID,
				Status:   "JOB_FAILED",
				Message:  fmt.Sprintf("Error in %s: %v", job.Type.String(), err),
			})
		}
	}()

	switch job.Type {
	case jobs.JobType_JOB_TYPE_SLICE:
		err = p.handleSlice(ctx, job)
	case jobs.JobType_JOB_TYPE_RECONSTRUCT:
		err = p.handleReconstruct(ctx, job)
	case jobs.JobType_JOB_TYPE_VIDEO_EXTRACT:
		err = p.handleVideoExtract(ctx, job)
	case jobs.JobType_JOB_TYPE_VIDEO_REASSEMBLE:
		err = p.handleVideoReassemble(ctx, job)
	case jobs.JobType_JOB_TYPE_RESIZE,
		jobs.JobType_JOB_TYPE_GRAYSCALE,
		jobs.JobType_JOB_TYPE_BLUR,
		jobs.JobType_JOB_TYPE_BRIGHTNESS,
		jobs.JobType_JOB_TYPE_UNSPECIFIED:
		err = p.handleProcess(ctx, job)
	}

	return err
}

func calcPadding(radius float64, startY, endY, height int) (paddingTop, paddingBottom int) {
	if radius <= 0 {
		return 0, 0
	}

	rInt := int(math.Ceil(radius))

	if startY > 0 {
		paddingTop = rInt
		if startY-paddingTop < 0 {
			paddingTop = startY
		}
	}

	if endY < height {
		paddingBottom = rInt
		if endY+paddingBottom > height {
			paddingBottom = height - endY
		}
	}

	return paddingTop, paddingBottom
}

func (p *Processor) handleSlice(ctx context.Context, job *jobs.Job) error {
	_ = p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:    job.ParentId,
		WorkerID: p.workerID,
		Status:   "SLICING",
		Progress: 0.05,
		Message:  "Dividiendo la imagen para procesamiento paralelo...",
	})

	reader, err := p.minio.DownloadObject(ctx, model.BucketName, job.OriginalImagePath)
	if err != nil {
		return fmt.Errorf("download original: %w", err)
	}
	defer reader.Close()

	buf, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read image: %w", err)
	}

	metadata, err := bimg.Metadata(buf)
	if err != nil {
		return fmt.Errorf("get image metadata: %w", err)
	}

	W, H := metadata.Size.Width, metadata.Size.Height
	totalPixels := W * H

	const (
		MaxPixelsPerSubtask = 1000000
		DefaultAttempts     = 3
	)

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

	radius := 0.0

	if targetType == jobs.JobType_JOB_TYPE_BLUR {
		if r, err := strconv.ParseFloat(job.Parameters["radius"], 64); err == nil {
			radius = r
		} else {
			radius = 1.0
		}
	}

	radius = getMaxBlurRadius(job.Parameters["effects"], radius)

	results := make([]subTaskResult, N)

	var g errgroup.Group
	g.SetLimit(N)

	for i := range N {
		g.Go(func() error {
			results[i] = p.createSliceSubtask(ctx, job, buf, i, N, W, H, rowsPerSubtask, targetType, radius)

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	for _, res := range results {
		if res.err != nil {
			return res.err
		}

		subtasks = append(subtasks, res.job)
	}

	if err := p.redis.InitializeTask(ctx, job.ParentId, N, DefaultAttempts); err != nil {
		return fmt.Errorf("initialize redis task: %w", err)
	}

	if err := p.redis.PushTasksPipeline(ctx, subtasks); err != nil {
		return fmt.Errorf("push subtasks: %w", err)
	}

	_ = p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:    job.ParentId,
		WorkerID: p.workerID,
		Status:   "PROCESSING",
		Progress: 0.10,
		Message:  fmt.Sprintf("Imagen dividida en %d fragmentos.", N),
	})

	return nil
}

func (p *Processor) createSliceSubtask(ctx context.Context, job *jobs.Job, buf []byte, i, n, w, h, rowsPerSubtask int, targetType jobs.JobType, radius float64) subTaskResult {
	startY := i * rowsPerSubtask

	endY := (i + 1) * rowsPerSubtask
	if i == n-1 {
		endY = h
	}

	paddingTop, paddingBottom := calcPadding(radius, startY, endY, h)

	fragmentHeight := (endY + paddingBottom) - (startY - paddingTop)

	fragmentBuf, err := ExtractRegion(buf, 0, startY-paddingTop, w, fragmentHeight)
	if err != nil {
		return subTaskResult{err: fmt.Errorf("extract fragment %d: %w", i, err)}
	}

	fragmentPath := job.ParentId + "/sub_" + strconv.Itoa(i) + ".png"

	_, err = p.minio.UploadObject(ctx, model.BucketName, fragmentPath, bytes.NewReader(fragmentBuf), int64(len(fragmentBuf)), "image/png")
	if err != nil {
		return subTaskResult{err: fmt.Errorf("upload fragment %d: %w", i, err)}
	}

	subJobParams := make(map[string]string)
	maps.Copy(subJobParams, job.Parameters)
	subJobParams["index"] = strconv.Itoa(i)
	subJobParams["total_subtasks"] = strconv.Itoa(n)
	subJobParams["original_width"] = strconv.Itoa(w)
	subJobParams["original_height"] = strconv.Itoa(h)
	subJobParams["padding_top"] = strconv.Itoa(paddingTop)
	subJobParams["padding_bottom"] = strconv.Itoa(paddingBottom)

	subJob := &jobs.Job{
		Id:                uuid.New().String(),
		Type:              targetType,
		Status:            jobs.JobStatus_JOB_STATUS_UNSPECIFIED,
		OriginalImagePath: fragmentPath,
		ParentId:          job.ParentId,
		Region: &jobs.Region{
			X:      0,
			Y:      0,
			Width:  int32(w),
			Height: int32(fragmentHeight),
		},
		Attempts:   3,
		CreatedAt:  time.Now().Unix(),
		Timestamp:  time.Now().Unix(),
		Parameters: subJobParams,
	}

	return subTaskResult{job: subJob}
}

func cropPadding(processed []byte, job *jobs.Job) ([]byte, error) {
	paddingTop, _ := strconv.Atoi(job.Parameters["padding_top"])
	paddingBottom, _ := strconv.Atoi(job.Parameters["padding_bottom"])

	if paddingTop <= 0 && paddingBottom <= 0 {
		return processed, nil
	}

	metadata, err := bimg.Metadata(processed)
	if err != nil {
		return nil, fmt.Errorf("get processed metadata: %w", err)
	}

	isResize, targetHeight := getResizeHeight(job)

	if isResize && targetHeight > 0 {
		originalHeight, _ := strconv.Atoi(job.Parameters["original_height"])
		scaleY := float64(targetHeight) / float64(originalHeight)
		paddingTop = int(float64(paddingTop) * scaleY)
		paddingBottom = int(float64(paddingBottom) * scaleY)
	}

	cropHeight := metadata.Size.Height - paddingTop - paddingBottom
	if cropHeight <= 0 {
		return processed, nil
	}

	return ExtractRegion(processed, 0, paddingTop, metadata.Size.Width, cropHeight)
}

func (p *Processor) handleProcess(ctx context.Context, job *jobs.Job) error {
	reader, err := p.minio.DownloadObject(ctx, model.BucketName, job.OriginalImagePath)
	if err != nil {
		return fmt.Errorf("download original: %w", err)
	}
	defer reader.Close()

	buf, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read fragment: %w", err)
	}

	var processed []byte

	if job.Parameters["is_segment"] == "true" {
		processed, err = p.ProcessVideoSegment(ctx, job, buf)
	} else {
		processed, err = p.applyEffectsPipeline(buf, job)
	}

	if err != nil {
		return fmt.Errorf("apply filter: %w", err)
	}

	if job.Parameters["is_segment"] != "true" {
		processed, err = cropPadding(processed, job)
		if err != nil {
			return fmt.Errorf("crop padding: %w", err)
		}
	}

	index := job.Parameters["index"]

	var resultPath string

	mimeType := "image/png"

	if job.Parameters["is_segment"] == "true" {
		idx, _ := strconv.Atoi(index)
		resultPath = fmt.Sprintf("%s/res_part_%03d.mp4", job.ParentId, idx)
		mimeType = "video/mp4"
	} else if job.Parameters["is_video"] == "true" {
		idx, _ := strconv.Atoi(index)
		resultPath = fmt.Sprintf("%s/res_frame_%06d.png", job.ParentId, idx)
	} else {
		resultPath = fmt.Sprintf("%s/res_sub_%s.png", job.ParentId, index)
	}

	_, err = p.minio.UploadObject(ctx, model.BucketName, resultPath, bytes.NewReader(processed), int64(len(processed)), mimeType)
	if err != nil {
		return fmt.Errorf("upload processed subtask: %w", err)
	}

	count, err := p.redis.DecrementCounter(ctx, job.ParentId)
	if err != nil {
		return fmt.Errorf("decrement counter: %w", err)
	}

	totalStr := job.Parameters["total_subtasks"]
	total, _ := strconv.Atoi(totalStr)
	completed := total - int(count)
	progress := 0.10 + (float64(completed)/float64(total))*0.80

	_ = p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:    job.ParentId,
		WorkerID: p.workerID,
		Status:   "PROCESSING",
		Progress: progress,
		Message:  fmt.Sprintf("Procesando fragmento %d de %d...", completed, total),
	})

	log.Printf("Task %s: subtasks remaining: %d", job.ParentId, count)

	if count <= 0 {
		if err := p.triggerReconstruction(ctx, job); err != nil {
			return err
		}
	}

	return nil
}

func (p *Processor) triggerReconstruction(ctx context.Context, job *jobs.Job) error {
	triggerKey := "{\"global\"}:reconstruct_triggered:" + job.ParentId

	ok, err := p.redis.SetNX(ctx, triggerKey, "1", 24*time.Hour)
	if err != nil {
		return fmt.Errorf("check reconstruction trigger: %w", err)
	}

	if !ok {
		return nil
	}

	var nextJob *jobs.Job
	if job.Parameters["is_video"] == "true" {
		nextJob = &jobs.Job{
			Id:                "reassemble-" + job.ParentId,
			Type:              jobs.JobType_JOB_TYPE_VIDEO_REASSEMBLE,
			ParentId:          job.ParentId,
			OriginalImagePath: job.OriginalImagePath,
			CreatedAt:         time.Now().Unix(),
			Timestamp:         time.Now().Unix(),
			Parameters:        job.Parameters,
		}
	} else {
		nextJob = &jobs.Job{
			Id:                "recon-" + job.ParentId,
			Type:              jobs.JobType_JOB_TYPE_RECONSTRUCT,
			ParentId:          job.ParentId,
			OriginalImagePath: job.OriginalImagePath,
			CreatedAt:         time.Now().Unix(),
			Timestamp:         time.Now().Unix(),
			Parameters:        job.Parameters,
		}
	}

	if err := p.redis.PushTask(ctx, nextJob); err != nil {
		return fmt.Errorf("push next task: %w", err)
	}

	log.Printf("Triggered next step for task %s", job.ParentId)

	return nil
}

func (p *Processor) handleReconstruct(ctx context.Context, job *jobs.Job) error {
	log.Printf("Starting reconstruction for task %s", job.ParentId)

	_ = p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:    job.ParentId,
		WorkerID: p.workerID,
		Status:   "RECONSTRUCTING",
		Progress: 0.95,
		Message:  "Reconstruyendo la imagen final...",
	})

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

	subImages := make([]image.Image, totalSubtasks)

	var g errgroup.Group
	g.SetLimit(totalSubtasks)

	for i := range totalSubtasks {
		g.Go(func() error {
			path := fmt.Sprintf("%s/res_sub_%d.png", job.ParentId, i)

			subReader, err := p.minio.DownloadObject(ctx, model.BucketName, path)
			if err != nil {
				return fmt.Errorf("download subtask %d: %w", i, err)
			}
			defer subReader.Close()

			subImg, _, err := image.Decode(subReader)
			if err != nil {
				return fmt.Errorf("decode subtask %d: %w", i, err)
			}

			subImages[i] = subImg

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	currentY := 0
	for _, subImg := range subImages {
		draw.Draw(finalImg, image.Rect(0, currentY, targetWidth, currentY+subImg.Bounds().Dy()), subImg, image.Point{}, draw.Src)
		currentY += subImg.Bounds().Dy()
	}

	finalPath := job.ParentId + "/final.png"

	var buf bytes.Buffer
	if err := png.Encode(&buf, finalImg); err != nil {
		return fmt.Errorf("encode final png: %w", err)
	}

	_, err := p.minio.UploadObject(ctx, model.BucketName, finalPath, &buf, int64(buf.Len()), "image/png")
	if err != nil {
		return fmt.Errorf("upload final image: %w", err)
	}

	_ = p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:     job.ParentId,
		WorkerID:  p.workerID,
		Status:    "JOB_COMPLETED",
		Progress:  1.0,
		Message:   "¡Procesamiento completado!",
		ResultURL: finalPath,
	})

	log.Printf("Completed reconstruction for task %s", job.ParentId)

	return nil
}

func (p *Processor) handleVideoExtract(ctx context.Context, job *jobs.Job) error {
	log.Printf("Starting video segment splitting for task %s", job.ParentId)

	_ = p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:    job.ParentId,
		WorkerID: p.workerID,
		Status:   "EXTRACTING",
		Progress: 0.05,
		Message:  "Dividiendo video en partes para procesamiento paralelo...",
	})

	reader, err := p.minio.DownloadObject(ctx, model.BucketName, job.OriginalImagePath)
	if err != nil {
		return fmt.Errorf("download video: %w", err)
	}
	defer reader.Close()

	buf, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read video: %w", err)
	}

	tmpVideo := fmt.Sprintf("/tmp/video_%s.mp4", job.ParentId)
	if err := os.WriteFile(tmpVideo, buf, 0o644); err != nil {
		return fmt.Errorf("write temp video: %w", err)
	}
	defer func() { _ = os.Remove(tmpVideo) }()

	width, height, fps, err := getVideoMetadata(ctx, tmpVideo)
	if err != nil {
		return fmt.Errorf("get video metadata: %w", err)
	}

	tmpPartsDir := "/tmp/parts_" + job.ParentId
	defer func() { _ = os.RemoveAll(tmpPartsDir) }()

	segmentDuration := 3
	if envVal := os.Getenv("SEGMENT_DURATION_SECONDS"); envVal != "" {
		if val, err := strconv.Atoi(envVal); err == nil && val > 0 {
			segmentDuration = val
		}
	}

	parts, err := SplitVideoIntoSegments(ctx, tmpVideo, tmpPartsDir, segmentDuration)
	if err != nil {
		return fmt.Errorf("split video into segments: %w", err)
	}

	log.Printf("Task %s: split video into %d parts (%dx%d) at %.2f FPS", job.ParentId, len(parts), width, height, fps)

	targetType := jobs.JobType(jobs.JobType_value[job.Parameters["target_type"]])

	const DefaultAttempts = 3

	subtasks := make([]*jobs.Job, len(parts))
	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(5) // Concurrently upload up to 5 segments

	for i, partPath := range parts {
		g.Go(func() error {
			partData, err := os.ReadFile(partPath)
			if err != nil {
				return fmt.Errorf("read segment part %d: %w", i, err)
			}

			partMinioPath := fmt.Sprintf("%s/part_%03d.mp4", job.ParentId, i)

			_, err = p.minio.UploadObject(gCtx, model.BucketName, partMinioPath, bytes.NewReader(partData), int64(len(partData)), "video/mp4")
			if err != nil {
				return fmt.Errorf("upload segment part %d: %w", i, err)
			}

			subJobParams := make(map[string]string)
			maps.Copy(subJobParams, job.Parameters)
			subJobParams["index"] = strconv.Itoa(i)
			subJobParams["total_subtasks"] = strconv.Itoa(len(parts))
			subJobParams["original_width"] = strconv.Itoa(width)
			subJobParams["original_height"] = strconv.Itoa(height)
			subJobParams["fps"] = fmt.Sprintf("%.3f", fps)
			subJobParams["is_segment"] = "true"
			subJobParams["padding_top"] = "0"
			subJobParams["padding_bottom"] = "0"

			subtasks[i] = &jobs.Job{
				Id:                uuid.New().String(),
				Type:              targetType,
				Status:            jobs.JobStatus_JOB_STATUS_UNSPECIFIED,
				OriginalImagePath: partMinioPath,
				ParentId:          job.ParentId,
				Region: &jobs.Region{
					X:      0,
					Y:      0,
					Width:  int32(width),
					Height: int32(height),
				},
				Attempts:   DefaultAttempts,
				CreatedAt:  time.Now().Unix(),
				Timestamp:  time.Now().Unix(),
				Parameters: subJobParams,
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if err := p.redis.InitializeTask(ctx, job.ParentId, len(subtasks), DefaultAttempts); err != nil {
		return fmt.Errorf("initialize redis task: %w", err)
	}

	if err := p.redis.PushTasksPipeline(ctx, subtasks); err != nil {
		return fmt.Errorf("push segment jobs: %w", err)
	}

	_ = p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:    job.ParentId,
		WorkerID: p.workerID,
		Status:   "PROCESSING",
		Progress: 0.10,
		Message:  fmt.Sprintf("Video dividido en %d partes para procesamiento paralelo.", len(subtasks)),
	})

	return nil
}

func (p *Processor) handleVideoReassemble(ctx context.Context, job *jobs.Job) error {
	log.Printf("Starting video reassembly for task %s", job.ParentId)

	_ = p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:    job.ParentId,
		WorkerID: p.workerID,
		Status:   "REASSEMBLING",
		Progress: 0.95,
		Message:  "Reensamblando video...",
	})

	totalParts, _ := strconv.Atoi(job.Parameters["total_subtasks"])

	tmpDir := "/tmp/reassemble_" + job.ParentId
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	var inputsContent strings.Builder

	// Download all processed segments
	for i := range totalParts {
		path := fmt.Sprintf("%s/res_part_%03d.mp4", job.ParentId, i)

		reader, err := p.minio.DownloadObject(ctx, model.BucketName, path)
		if err != nil {
			return fmt.Errorf("download segment part %d: %w", i, err)
		}

		partData, err := io.ReadAll(reader)
		reader.Close()

		if err != nil {
			return fmt.Errorf("read segment part %d: %w", i, err)
		}

		tmpPath := filepath.Join(tmpDir, fmt.Sprintf("part_%03d.mp4", i))
		if err := os.WriteFile(tmpPath, partData, 0o644); err != nil {
			return fmt.Errorf("write segment part %d: %w", i, err)
		}

		_, _ = fmt.Fprintf(&inputsContent, "file '%s'\n", tmpPath)
	}

	inputsTxtPath := filepath.Join(tmpDir, "inputs.txt")
	if err := os.WriteFile(inputsTxtPath, []byte(inputsContent.String()), 0o644); err != nil {
		return fmt.Errorf("write inputs.txt: %w", err)
	}

	// Concatenate segments
	tmpVideo := filepath.Join(tmpDir, "output.mp4")
	cmd := exec.CommandContext(ctx, "ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", inputsTxtPath, "-c", "copy", tmpVideo)

	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg concat: %w: %s", err, stderr.String())
	}

	videoData, err := os.ReadFile(tmpVideo)
	if err != nil {
		return fmt.Errorf("read output video: %w", err)
	}

	finalPath := job.ParentId + "/final.mp4"

	_, err = p.minio.UploadObject(ctx, model.BucketName, finalPath, bytes.NewReader(videoData), int64(len(videoData)), "video/mp4")
	if err != nil {
		return fmt.Errorf("upload final video: %w", err)
	}

	_ = p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:     job.ParentId,
		WorkerID:  p.workerID,
		Status:    "JOB_COMPLETED",
		Progress:  1.0,
		Message:   "¡Procesamiento de video completado!",
		ResultURL: finalPath,
	})

	log.Printf("Completed video reassembly for task %s", job.ParentId)

	return nil
}

func (p *Processor) ProcessVideoSegment(ctx context.Context, job *jobs.Job, inputSegmentData []byte) ([]byte, error) {
	tmpDir := "/tmp/proc_seg_" + job.Id
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpInputVideo := filepath.Join(tmpDir, "input.mp4")
	if err := os.WriteFile(tmpInputVideo, inputSegmentData, 0o644); err != nil {
		return nil, fmt.Errorf("write temp segment: %w", err)
	}

	tmpFramesDir := filepath.Join(tmpDir, "frames")

	frames, width, height, fps, err := ExtractFrames(ctx, tmpInputVideo, tmpFramesDir)
	if err != nil {
		return nil, fmt.Errorf("extract segment frames: %w", err)
	}

	_ = width
	_ = height

	workerConcurrency := 8
	if envVal := os.Getenv("WORKER_CONCURRENCY"); envVal != "" {
		if val, err := strconv.Atoi(envVal); err == nil && val > 0 {
			workerConcurrency = val
		}
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.SetLimit(workerConcurrency)

	for _, frame := range frames {
		g.Go(func() error {
			if gCtx.Err() != nil {
				return gCtx.Err()
			}

			frameData, err := os.ReadFile(frame.Path)
			if err != nil {
				return fmt.Errorf("read frame %d: %w", frame.Index, err)
			}

			processed, err := p.applyEffectsPipeline(frameData, job)
			if err != nil {
				return fmt.Errorf("process segment frame %d: %w", frame.Index, err)
			}

			if err := os.WriteFile(frame.Path, processed, 0o644); err != nil {
				return fmt.Errorf("write processed frame %d: %w", frame.Index, err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	tmpOutputVideo := filepath.Join(tmpDir, "output.mp4")
	if err := ReassembleVideo(ctx, tmpFramesDir, tmpOutputVideo, fps); err != nil {
		return nil, fmt.Errorf("reassemble processed segment: %w", err)
	}

	outputData, err := os.ReadFile(tmpOutputVideo)
	if err != nil {
		return nil, fmt.Errorf("read output segment: %w", err)
	}

	return outputData, nil
}

type pipelineEffect struct {
	Type   string            `json:"type"`
	Params map[string]string `json:"params"`
}

func (p *Processor) applyEffectsPipeline(data []byte, job *jobs.Job) ([]byte, error) {
	effectsJSON := job.Parameters["effects"]
	if effectsJSON == "" {
		return p.applySingleFilter(data, job.Type, job.Parameters, job.Region)
	}

	var effects []pipelineEffect
	if err := json.Unmarshal([]byte(effectsJSON), &effects); err != nil {
		log.Printf("Warning: failed to unmarshal effects JSON: %v. Falling back to single filter.", err)
		return p.applySingleFilter(data, job.Type, job.Parameters, job.Region)
	}

	current := data

	var err error

	for _, effect := range effects {
		var jobType jobs.JobType

		switch effect.Type {
		case "GRAYSCALE":
			jobType = jobs.JobType_JOB_TYPE_GRAYSCALE
		case "BLUR":
			jobType = jobs.JobType_JOB_TYPE_BLUR
		case "BRIGHTNESS":
			jobType = jobs.JobType_JOB_TYPE_BRIGHTNESS
		case "RESIZE":
			jobType = jobs.JobType_JOB_TYPE_RESIZE
		default:
			jobType = jobs.JobType_JOB_TYPE_UNSPECIFIED
		}

		current, err = p.applySingleFilter(current, jobType, effect.Params, job.Region)
		if err != nil {
			return nil, err
		}
	}

	return current, nil
}

func (p *Processor) applySingleFilter(data []byte, jobType jobs.JobType, params map[string]string, region *jobs.Region) ([]byte, error) {
	var (
		processed []byte
		err       error
	)

	switch jobType {
	case jobs.JobType_JOB_TYPE_GRAYSCALE:
		processed, err = ApplyGrayscale(data)
	case jobs.JobType_JOB_TYPE_BLUR:
		radius := 1.0
		if r, err := strconv.ParseFloat(params["radius"], 64); err == nil {
			radius = r
		}

		processed, err = ApplyBlur(data, radius)
	case jobs.JobType_JOB_TYPE_BRIGHTNESS:
		factor := 1.0
		if f, err := strconv.ParseFloat(params["factor"], 64); err == nil {
			factor = f
		}

		processed, err = ApplyBrightness(data, factor)
	case jobs.JobType_JOB_TYPE_RESIZE:
		targetWidth, _ := strconv.Atoi(params["width"])
		targetHeight, _ := strconv.Atoi(params["height"])
		originalWidth, _ := strconv.Atoi(params["original_width"])
		originalHeight, _ := strconv.Atoi(params["original_height"])

		if region != nil && originalHeight > 0 && originalWidth > 0 && targetHeight > 0 && targetWidth > 0 {
			scaleY := float64(targetHeight) / float64(originalHeight)
			newFragHeight := int(float64(region.Height) * scaleY)

			scaleX := float64(targetWidth) / float64(originalWidth)
			newFragWidth := int(float64(region.Width) * scaleX)

			processed, err = ApplyResize(data, newFragWidth, newFragHeight)
		} else if targetWidth > 0 && targetHeight > 0 {
			processed, err = ApplyResize(data, targetWidth, targetHeight)
		} else {
			processed = data
		}
	default:
		processed = data
	}

	return processed, err
}

func getMaxBlurRadius(effectsJSON string, currentRadius float64) float64 {
	if effectsJSON == "" {
		return currentRadius
	}

	var effects []pipelineEffect
	if err := json.Unmarshal([]byte(effectsJSON), &effects); err != nil {
		return currentRadius
	}

	radius := currentRadius

	for _, effect := range effects {
		if effect.Type != "BLUR" {
			continue
		}

		r, err := strconv.ParseFloat(effect.Params["radius"], 64)
		if err != nil {
			if radius < 1.0 {
				radius = 1.0
			}

			continue
		}

		if r > radius {
			radius = r
		}
	}

	return radius
}

func getResizeHeight(job *jobs.Job) (bool, int) {
	if job.Type == jobs.JobType_JOB_TYPE_RESIZE {
		h, _ := strconv.Atoi(job.Parameters["height"])
		return true, h
	}

	effectsJSON := job.Parameters["effects"]
	if effectsJSON == "" {
		return false, 0
	}

	var effects []pipelineEffect
	if err := json.Unmarshal([]byte(effectsJSON), &effects); err != nil {
		return false, 0
	}

	for _, effect := range effects {
		if effect.Type == "RESIZE" {
			h, _ := strconv.Atoi(effect.Params["height"])
			return true, h
		}
	}

	return false, 0
}
