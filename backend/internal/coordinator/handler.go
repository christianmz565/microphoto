package coordinator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/christianmz565/microphoto/pkg/client/metrics"
	"github.com/christianmz565/microphoto/pkg/model"
	jobs "github.com/christianmz565/microphoto/proto/jobs/v1"
	"github.com/google/uuid"
)

type cachedPreview struct {
	data      []byte
	createdAt time.Time
}

// HTTPHandler manages the HTTP API for the coordinator.
type HTTPHandler struct {
	orchestrator  *Orchestrator
	metrics       *metrics.Metrics
	maxUploadSize int64
	wg            sync.WaitGroup
	previewCache  sync.Map // maps previewID (string) to cachedPreview
}

// NewHTTPHandler creates a new HTTPHandler instance.
func NewHTTPHandler(orch *Orchestrator, m *metrics.Metrics, maxUploadSize int64) *HTTPHandler {
	h := &HTTPHandler{
		orchestrator:  orch,
		metrics:       m,
		maxUploadSize: maxUploadSize,
	}

	h.wg.Go(func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()

			h.previewCache.Range(func(key, value any) bool {
				cp, ok := value.(cachedPreview)
				if ok && now.Sub(cp.createdAt) > 30*time.Minute {
					h.previewCache.Delete(key)
				}

				return true
			})
		}
	})

	return h
}

// RegisterRoutes registers the HTTP routes with the provided mux.
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.HealthCheck)
	mux.HandleFunc("/api/v1/process", h.ProcessImage)
	mux.HandleFunc("/api/v1/process-video", h.ProcessVideo)
	mux.HandleFunc("/api/v1/preview", h.PreviewImage)
	mux.HandleFunc("/api/v1/result/", h.DownloadResult)
	mux.HandleFunc("/api/v1/events/", h.StreamEvents)
}

// Wait blocks until all background goroutines complete.
func (h *HTTPHandler) Wait() {
	h.wg.Wait()
}

// DownloadResult serves the processed image or video for a given task ID.
func (h *HTTPHandler) DownloadResult(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Path[len("/api/v1/result/"):]
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	reader, err := h.orchestrator.DownloadVideoResult(r.Context(), taskID)
	if err != nil {
		reader, err = h.orchestrator.DownloadResult(r.Context(), taskID)
		if err != nil {
			log.Printf("Error downloading result for task %s: %v", taskID, err)
			http.Error(w, "Result not found or not ready", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "image/png")
	} else {
		w.Header().Set("Content-Type", "video/mp4")
	}
	defer reader.Close()

	if _, err := io.Copy(w, reader); err != nil {
		log.Printf("Error writing response for task %s: %v", taskID, err)
	}
}

// HealthCheck provides a simple health check endpoint.
func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte("OK")); err != nil {
		log.Printf("Error writing health check response: %v", err)
	}
}

// StreamEvents handles SSE connections for task progress events.
func (h *HTTPHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	taskID := strings.TrimPrefix(r.URL.Path, "/api/v1/events/")
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ctx := r.Context()

	pubsub, ch := h.orchestrator.redis.SubscribeProgress(ctx, taskID)
	defer func() { _ = pubsub.Close() }()

	events, err := h.orchestrator.redis.GetProgressEvents(ctx, taskID)
	if err != nil {
		log.Printf("Error getting progress events for task %s: %v", taskID, err)
	}

	var lastTimestamp int64

	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}

		_, _ = fmt.Fprintf(w, "data: %s\n\n", data)

		if event.Timestamp > lastTimestamp {
			lastTimestamp = event.Timestamp
		}
	}

	flusher.Flush()

	keepalive := time.NewTicker(15 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			var event model.ProgressPayload
			if err := json.Unmarshal([]byte(msg.Payload), &event); err == nil {
				if event.Timestamp <= lastTimestamp {
					continue
				}
			}

			_, _ = fmt.Fprintf(w, "data: %s\n\n", msg.Payload)

			flusher.Flush()
		case <-keepalive.C:
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")

			flusher.Flush()
		}
	}
}

// ProcessImage handles the multipart form upload of an image for processing.
func (h *HTTPHandler) ProcessImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > h.maxUploadSize {
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}

	jobTypeStr := r.FormValue("type")

	jobType := parseJobType(jobTypeStr)

	params := make(map[string]string)
	if r.FormValue("radius") != "" {
		params["radius"] = r.FormValue("radius")
	}

	if r.FormValue("factor") != "" {
		params["factor"] = r.FormValue("factor")
	}

	if r.FormValue("width") != "" {
		params["width"] = r.FormValue("width")
	}

	if r.FormValue("height") != "" {
		params["height"] = r.FormValue("height")
	}

	if r.FormValue("effects") != "" {
		params["effects"] = r.FormValue("effects")
	}

	tmpFile, err := os.CreateTemp("", "upload-img-*.tmp")
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}

	tmpPath := tmpFile.Name()
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	size, err := io.Copy(tmpFile, file)
	if err != nil {
		http.Error(w, "Failed to save upload", http.StatusInternalServerError)
		return
	}

	tmpFile.Close()
	tmpFile = nil

	taskID := uuid.New().String()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"task_id": taskID,
	}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

	h.wg.Go(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		defer os.Remove(tmpPath)

		f, err := os.Open(tmpPath)
		if err != nil {
			log.Printf("Background processing failed for task %s: failed to open temp file: %v", taskID, err)
			return
		}
		defer f.Close()

		err = h.orchestrator.ProcessImage(ctx, taskID, f, header.Filename, jobType, size, params)
		if err != nil {
			log.Printf("Background processing failed for task %s: %v", taskID, err)
			_ = h.orchestrator.redis.PublishProgress(context.Background(), taskID, model.ProgressPayload{
				JobID:     taskID,
				Status:    "JOB_FAILED",
				Message:   err.Error(),
				Timestamp: time.Now().UnixNano(),
			})
		}
	})
}

// ProcessVideo handles the multipart form upload of a video for processing.
func (h *HTTPHandler) ProcessVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Video file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > h.maxUploadSize {
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}

	jobTypeStr := r.FormValue("type")

	jobType := parseJobType(jobTypeStr)

	params := make(map[string]string)
	if r.FormValue("fps") != "" {
		params["fps"] = r.FormValue("fps")
	}

	if r.FormValue("radius") != "" {
		params["radius"] = r.FormValue("radius")
	}

	if r.FormValue("factor") != "" {
		params["factor"] = r.FormValue("factor")
	}

	if r.FormValue("width") != "" {
		params["width"] = r.FormValue("width")
	}

	if r.FormValue("height") != "" {
		params["height"] = r.FormValue("height")
	}

	if r.FormValue("effects") != "" {
		params["effects"] = r.FormValue("effects")
	}

	tmpFile, err := os.CreateTemp("", "upload-vid-*.tmp")
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}

	tmpPath := tmpFile.Name()
	defer func() {
		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}
	}()

	size, err := io.Copy(tmpFile, file)
	if err != nil {
		http.Error(w, "Failed to save upload", http.StatusInternalServerError)
		return
	}

	tmpFile.Close()
	tmpFile = nil

	taskID := uuid.New().String()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	if err := json.NewEncoder(w).Encode(map[string]string{
		"task_id": taskID,
	}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}

	h.wg.Go(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		defer os.Remove(tmpPath)

		f, err := os.Open(tmpPath)
		if err != nil {
			log.Printf("Background video processing failed for task %s: failed to open temp file: %v", taskID, err)
			return
		}
		defer f.Close()

		err = h.orchestrator.ProcessVideo(ctx, taskID, f, header.Filename, jobType, size, params)
		if err != nil {
			log.Printf("Background video processing failed for task %s: %v", taskID, err)
			_ = h.orchestrator.redis.PublishProgress(context.Background(), taskID, model.ProgressPayload{
				JobID:     taskID,
				Status:    "JOB_FAILED",
				Message:   err.Error(),
				Timestamp: time.Now().UnixNano(),
			})
		}
	})
}

// parseJobType converts a string representation of a job type to the corresponding protobuf enum.
func parseJobType(s string) jobs.JobType {
	switch s {
	case "RESIZE":
		return jobs.JobType_JOB_TYPE_RESIZE
	case "GRAYSCALE":
		return jobs.JobType_JOB_TYPE_GRAYSCALE
	case "BLUR":
		return jobs.JobType_JOB_TYPE_BLUR
	case "BRIGHTNESS":
		return jobs.JobType_JOB_TYPE_BRIGHTNESS
	case "RECONSTRUCT":
		return jobs.JobType_JOB_TYPE_RECONSTRUCT
	default:
		return jobs.JobType_JOB_TYPE_UNSPECIFIED
	}
}
