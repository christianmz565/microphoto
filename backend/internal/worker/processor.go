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
	var err error

	defer func() {
		p.metrics.RecordTaskDuration(ctx, p.workerID, time.Since(startTime).Seconds())
		p.metrics.RecordTaskProcessed(ctx, p.workerID, job.Type.String())

		if err != nil {
			// Terminal error: notify the frontend
			p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
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
	default:
		err = p.handleProcess(ctx, job)
	}

	return err
}

func (p *Processor) handleSlice(ctx context.Context, job *jobs.Job) error {
	p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:    job.ParentId,
		WorkerID: p.workerID,
		Status:   "SLICING",
		Progress: 0.05,
		Message:  "Dividiendo la imagen para procesamiento paralelo...",
	})

	reader, err := p.minio.DownloadObject(ctx, BucketName, job.OriginalImagePath)
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

	radius := 0.0
	if targetType == jobs.JobType_JOB_TYPE_BLUR {
		if r, err := strconv.ParseFloat(job.Parameters["radius"], 64); err == nil {
			radius = r
		} else {
			radius = 1.0
		}
	}

	type subTaskResult struct {
		job *jobs.Job
		err error
	}
	resChan := make(chan subTaskResult, N)

	for i := 0; i < N; i++ {
		go func(i int) {
			startY := i * rowsPerSubtask
			endY := (i + 1) * rowsPerSubtask
			if i == N-1 {
				endY = H
			}

			paddingTop := 0
			paddingBottom := 0
			if radius > 0 {
				rInt := int(math.Ceil(radius))
				if startY > 0 {
					paddingTop = rInt
					if startY-paddingTop < 0 {
						paddingTop = startY
					}
				}
				if endY < H {
					paddingBottom = rInt
					if endY+paddingBottom > H {
						paddingBottom = H - endY
					}
				}
			}

			fragmentHeight := (endY + paddingBottom) - (startY - paddingTop)
			fragmentBuf, err := ExtractRegion(buf, 0, startY-paddingTop, W, fragmentHeight)
			if err != nil {
				resChan <- subTaskResult{err: fmt.Errorf("extract fragment %d: %w", i, err)}
				return
			}

			fragmentPath := fmt.Sprintf("%s/sub_%d.png", job.ParentId, i)
			_, err = p.minio.UploadObject(ctx, BucketName, fragmentPath, bytes.NewReader(fragmentBuf), int64(len(fragmentBuf)), "image/png")
			if err != nil {
				resChan <- subTaskResult{err: fmt.Errorf("upload fragment %d: %w", i, err)}
				return
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
					Height: int32(fragmentHeight),
				},
				Attempts:   DefaultAttempts,
				CreatedAt:  time.Now().Unix(),
				Timestamp:  time.Now().Unix(),
				Parameters: subJobParams,
			}
			resChan <- subTaskResult{job: subJob}
		}(i)
	}

	for i := 0; i < N; i++ {
		res := <-resChan
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

	p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:    job.ParentId,
		WorkerID: p.workerID,
		Status:   "PROCESSING",
		Progress: 0.10,
		Message:  fmt.Sprintf("Imagen dividida en %d fragmentos.", N),
	})

	return nil
}

func (p *Processor) handleProcess(ctx context.Context, job *jobs.Job) error {
	// We don't publish progress for every subtask start to avoid flooding,
	// but we could if needed. For now, we update on completion.

	reader, err := p.minio.DownloadObject(ctx, BucketName, job.OriginalImagePath)
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
		switch job.Type {
		case jobs.JobType_JOB_TYPE_GRAYSCALE:
			processed, err = ApplyGrayscale(buf)
		case jobs.JobType_JOB_TYPE_BLUR:
			radius := 1.0
			if r, err := strconv.ParseFloat(job.Parameters["radius"], 64); err == nil {
				radius = r
			}
			processed, err = ApplyBlur(buf, radius)
		case jobs.JobType_JOB_TYPE_BRIGHTNESS:
			factor := 1.0
			if f, err := strconv.ParseFloat(job.Parameters["factor"], 64); err == nil {
				factor = f
			}
			processed, err = ApplyBrightness(buf, factor)
		case jobs.JobType_JOB_TYPE_RESIZE:
			targetWidth, _ := strconv.Atoi(job.Parameters["width"])
			targetHeight, _ := strconv.Atoi(job.Parameters["height"])
			originalWidth, _ := strconv.Atoi(job.Parameters["original_width"])
			originalHeight, _ := strconv.Atoi(job.Parameters["original_height"])

			if originalHeight > 0 && originalWidth > 0 && targetHeight > 0 && targetWidth > 0 {
				scaleY := float64(targetHeight) / float64(originalHeight)
				newFragHeight := int(float64(job.Region.Height) * scaleY)

				scaleX := float64(targetWidth) / float64(originalWidth)
				newFragWidth := int(float64(job.Region.Width) * scaleX)

				processed, err = ApplyResize(buf, newFragWidth, newFragHeight)
			} else {
				processed = buf
			}
		default:
			processed = buf
		}
	}

	if err != nil {
		return fmt.Errorf("apply filter: %w", err)
	}

	if job.Parameters["is_segment"] != "true" {
		paddingTop, _ := strconv.Atoi(job.Parameters["padding_top"])
		paddingBottom, _ := strconv.Atoi(job.Parameters["padding_bottom"])

		if paddingTop > 0 || paddingBottom > 0 {
			metadata, err := bimg.Metadata(processed)
			if err != nil {
				return fmt.Errorf("get processed metadata: %w", err)
			}

			if job.Type == jobs.JobType_JOB_TYPE_RESIZE {
				originalHeight, _ := strconv.Atoi(job.Parameters["original_height"])
				targetHeight, _ := strconv.Atoi(job.Parameters["height"])
				scaleY := float64(targetHeight) / float64(originalHeight)
				paddingTop = int(float64(paddingTop) * scaleY)
				paddingBottom = int(float64(paddingBottom) * scaleY)
			}

			cropY := paddingTop
			cropHeight := metadata.Size.Height - paddingTop - paddingBottom
			if cropHeight > 0 {
				processed, err = ExtractRegion(processed, 0, cropY, metadata.Size.Width, cropHeight)
				if err != nil {
					return fmt.Errorf("crop padding: %w", err)
				}
			}
		}
	}

	index := job.Parameters["index"]
	var resultPath string
	var mimeType string = "image/png"
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

	_, err = p.minio.UploadObject(ctx, BucketName, resultPath, bytes.NewReader(processed), int64(len(processed)), mimeType)
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

	p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:    job.ParentId,
		WorkerID: p.workerID,
		Status:   "PROCESSING",
		Progress: progress,
		Message:  fmt.Sprintf("Procesando fragmento %d de %d...", completed, total),
	})

	log.Printf("Task %s: subtasks remaining: %d", job.ParentId, count)

	if count <= 0 {
		triggerKey := fmt.Sprintf(`{"global"}:reconstruct_triggered:%s`, job.ParentId)
		ok, err := p.redis.SetNX(ctx, triggerKey, "1", 24*time.Hour)
		if err != nil {
			return fmt.Errorf("check reconstruction trigger: %w", err)
		}

		if ok {
			var nextJob *jobs.Job
			if job.Parameters["is_video"] == "true" {
				nextJob = &jobs.Job{
					Id:                fmt.Sprintf("reassemble-%s", job.ParentId),
					Type:              jobs.JobType_JOB_TYPE_VIDEO_REASSEMBLE,
					ParentId:          job.ParentId,
					OriginalImagePath: job.OriginalImagePath,
					CreatedAt:         time.Now().Unix(),
					Timestamp:         time.Now().Unix(),
					Parameters:        job.Parameters,
				}
			} else {
				nextJob = &jobs.Job{
					Id:                fmt.Sprintf("recon-%s", job.ParentId),
					Type:              jobs.JobType_JOB_TYPE_RECONSTRUCT,
					ParentId:          job.ParentId,
					OriginalImagePath: job.OriginalImagePath,
					CreatedAt:         time.Now().Unix(),
					Timestamp:         time.Now().Unix(),
					Parameters:        job.Parameters,
				}
			}

			err = p.redis.PushTask(ctx, nextJob)
			if err != nil {
				return fmt.Errorf("push next task: %w", err)
			}
			log.Printf("Triggered next step for task %s", job.ParentId)
		}
	}

	return nil
}

func (p *Processor) handleReconstruct(ctx context.Context, job *jobs.Job) error {
	log.Printf("Starting reconstruction for task %s", job.ParentId)

	p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
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

	type subImgResult struct {
		index int
		img   image.Image
		err   error
	}
	resChan := make(chan subImgResult, totalSubtasks)

	for i := range totalSubtasks {
		go func(i int) {
			path := fmt.Sprintf("%s/res_sub_%d.png", job.ParentId, i)
			subReader, err := p.minio.DownloadObject(ctx, BucketName, path)
			if err != nil {
				resChan <- subImgResult{err: fmt.Errorf("download subtask %d: %w", i, err)}
				return
			}
			defer subReader.Close()

			subImg, _, err := image.Decode(subReader)
			if err != nil {
				resChan <- subImgResult{err: fmt.Errorf("decode subtask %d: %w", i, err)}
				return
			}
			resChan <- subImgResult{index: i, img: subImg}
		}(i)
	}

	subImages := make([]image.Image, totalSubtasks)
	for range totalSubtasks {
		res := <-resChan
		if res.err != nil {
			return res.err
		}
		subImages[res.index] = res.img
	}

	currentY := 0
	for _, subImg := range subImages {
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

	p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:    job.ParentId,
		WorkerID: p.workerID,
		Status:   "EXTRACTING",
		Progress: 0.05,
		Message:  "Dividiendo video en partes para procesamiento paralelo...",
	})

	reader, err := p.minio.DownloadObject(ctx, BucketName, job.OriginalImagePath)
	if err != nil {
		return fmt.Errorf("download video: %w", err)
	}
	defer reader.Close()

	buf, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read video: %w", err)
	}

	tmpVideo := fmt.Sprintf("/tmp/video_%s.mp4", job.ParentId)
	if err := os.WriteFile(tmpVideo, buf, 0644); err != nil {
		return fmt.Errorf("write temp video: %w", err)
	}
	defer os.Remove(tmpVideo)

	width, height, fps, err := getVideoMetadata(tmpVideo)
	if err != nil {
		return fmt.Errorf("get video metadata: %w", err)
	}

	tmpPartsDir := fmt.Sprintf("/tmp/parts_%s", job.ParentId)
	defer os.RemoveAll(tmpPartsDir)

	parts, err := SplitVideoIntoSegments(tmpVideo, tmpPartsDir, 3) // 3-second segments
	if err != nil {
		return fmt.Errorf("split video into segments: %w", err)
	}

	log.Printf("Task %s: split video into %d parts (%dx%d) at %.2f FPS", job.ParentId, len(parts), width, height, fps)

	targetType := jobs.JobType(jobs.JobType_value[job.Parameters["target_type"]])
	const DefaultAttempts = 3

	var subtasks []*jobs.Job
	for i, partPath := range parts {
		partData, err := os.ReadFile(partPath)
		if err != nil {
			return fmt.Errorf("read segment part %d: %w", i, err)
		}

		partMinioPath := fmt.Sprintf("%s/part_%03d.mp4", job.ParentId, i)
		_, err = p.minio.UploadObject(ctx, BucketName, partMinioPath, bytes.NewReader(partData), int64(len(partData)), "video/mp4")
		if err != nil {
			return fmt.Errorf("upload segment part %d: %w", i, err)
		}

		subJobParams := make(map[string]string)
		maps.Copy(subJobParams, job.Parameters)
		subJobParams["index"] = fmt.Sprintf("%d", i)
		subJobParams["total_subtasks"] = fmt.Sprintf("%d", len(parts))
		subJobParams["original_width"] = fmt.Sprintf("%d", width)
		subJobParams["original_height"] = fmt.Sprintf("%d", height)
		subJobParams["fps"] = fmt.Sprintf("%.3f", fps)
		subJobParams["is_segment"] = "true"
		subJobParams["padding_top"] = "0"
		subJobParams["padding_bottom"] = "0"

		subJob := &jobs.Job{
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
		subtasks = append(subtasks, subJob)
	}

	if err := p.redis.InitializeTask(ctx, job.ParentId, len(subtasks), DefaultAttempts); err != nil {
		return fmt.Errorf("initialize redis task: %w", err)
	}

	if err := p.redis.PushTasksPipeline(ctx, subtasks); err != nil {
		return fmt.Errorf("push segment jobs: %w", err)
	}

	p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
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

	p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
		JobID:    job.ParentId,
		WorkerID: p.workerID,
		Status:   "REASSEMBLING",
		Progress: 0.95,
		Message:  "Reensamblando video...",
	})

	totalParts, _ := strconv.Atoi(job.Parameters["total_subtasks"])

	tmpDir := fmt.Sprintf("/tmp/reassemble_%s", job.ParentId)
	defer os.RemoveAll(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}

	var inputsContent strings.Builder

	// Download all processed segments
	for i := 0; i < totalParts; i++ {
		path := fmt.Sprintf("%s/res_part_%03d.mp4", job.ParentId, i)
		reader, err := p.minio.DownloadObject(ctx, BucketName, path)
		if err != nil {
			return fmt.Errorf("download segment part %d: %w", i, err)
		}

		partData, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			return fmt.Errorf("read segment part %d: %w", i, err)
		}

		tmpPath := filepath.Join(tmpDir, fmt.Sprintf("part_%03d.mp4", i))
		if err := os.WriteFile(tmpPath, partData, 0644); err != nil {
			return fmt.Errorf("write segment part %d: %w", i, err)
		}

		inputsContent.WriteString(fmt.Sprintf("file '%s'\n", tmpPath))
	}

	inputsTxtPath := filepath.Join(tmpDir, "inputs.txt")
	if err := os.WriteFile(inputsTxtPath, []byte(inputsContent.String()), 0644); err != nil {
		return fmt.Errorf("write inputs.txt: %w", err)
	}

	// Concatenate segments
	tmpVideo := filepath.Join(tmpDir, "output.mp4")
	cmd := exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", inputsTxtPath, "-c", "copy", tmpVideo)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg concat: %w: %s", err, stderr.String())
	}

	// Upload final video
	videoData, err := os.ReadFile(tmpVideo)
	if err != nil {
		return fmt.Errorf("read output video: %w", err)
	}

	finalPath := fmt.Sprintf("%s/final.mp4", job.ParentId)
	_, err = p.minio.UploadObject(ctx, BucketName, finalPath, bytes.NewReader(videoData), int64(len(videoData)), "video/mp4")
	if err != nil {
		return fmt.Errorf("upload final video: %w", err)
	}

	p.redis.PublishProgress(ctx, job.ParentId, model.ProgressPayload{
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
	tmpDir := fmt.Sprintf("/tmp/proc_seg_%s", job.Id)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpInputVideo := filepath.Join(tmpDir, "input.mp4")
	if err := os.WriteFile(tmpInputVideo, inputSegmentData, 0644); err != nil {
		return nil, fmt.Errorf("write temp segment: %w", err)
	}

	tmpFramesDir := filepath.Join(tmpDir, "frames")
	frames, width, height, fps, err := ExtractFrames(tmpInputVideo, tmpFramesDir)
	if err != nil {
		return nil, fmt.Errorf("extract segment frames: %w", err)
	}

	_ = width  // not used directly, metadata inherited from parent
	_ = height

	// Process each frame
	for _, frame := range frames {
		frameData, err := os.ReadFile(frame.Path)
		if err != nil {
			return nil, fmt.Errorf("read frame %d: %w", frame.Index, err)
		}

		var processed []byte
		switch job.Type {
		case jobs.JobType_JOB_TYPE_GRAYSCALE:
			processed, err = ApplyGrayscale(frameData)
		case jobs.JobType_JOB_TYPE_BLUR:
			radius := 1.0
			if r, err := strconv.ParseFloat(job.Parameters["radius"], 64); err == nil {
				radius = r
			}
			processed, err = ApplyBlur(frameData, radius)
		case jobs.JobType_JOB_TYPE_BRIGHTNESS:
			factor := 1.0
			if f, err := strconv.ParseFloat(job.Parameters["factor"], 64); err == nil {
				factor = f
			}
			processed, err = ApplyBrightness(frameData, factor)
		case jobs.JobType_JOB_TYPE_RESIZE:
			targetWidth, _ := strconv.Atoi(job.Parameters["width"])
			targetHeight, _ := strconv.Atoi(job.Parameters["height"])
			if targetWidth > 0 && targetHeight > 0 {
				processed, err = ApplyResize(frameData, targetWidth, targetHeight)
			} else {
				processed = frameData
			}
		default:
			processed = frameData
		}

		if err != nil {
			return nil, fmt.Errorf("process segment frame %d: %w", frame.Index, err)
		}

		if err := os.WriteFile(frame.Path, processed, 0644); err != nil {
			return nil, fmt.Errorf("write processed frame %d: %w", frame.Index, err)
		}
	}

	tmpOutputVideo := filepath.Join(tmpDir, "output.mp4")
	if err := ReassembleVideo(tmpFramesDir, tmpOutputVideo, fps); err != nil {
		return nil, fmt.Errorf("reassemble processed segment: %w", err)
	}

	outputData, err := os.ReadFile(tmpOutputVideo)
	if err != nil {
		return nil, fmt.Errorf("read output segment: %w", err)
	}

	return outputData, nil
}
