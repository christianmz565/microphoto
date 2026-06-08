package coordinator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/christianmz565/microphoto/pkg/client/metrics"
	jobs "github.com/christianmz565/microphoto/proto/jobs/v1"
	"github.com/google/uuid"
)

// HTTPHandler manages the HTTP API for the coordinator.
type HTTPHandler struct {
	orchestrator  *Orchestrator
	metrics       *metrics.Metrics
	maxUploadSize int64
}

// NewHTTPHandler creates a new HTTPHandler instance.
func NewHTTPHandler(orch *Orchestrator, m *metrics.Metrics, maxUploadSize int64) *HTTPHandler {
	return &HTTPHandler{
		orchestrator:  orch,
		metrics:       m,
		maxUploadSize: maxUploadSize,
	}
}

// RegisterRoutes registers the HTTP routes with the provided mux.
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.HealthCheck)
	mux.HandleFunc("/api/v1/process", h.ProcessImage)
	mux.HandleFunc("/api/v1/result/", h.DownloadResult)
	mux.HandleFunc("/api/v1/events/", h.StreamEvents)
}

// DownloadResult serves the processed image for a given task ID.
func (h *HTTPHandler) DownloadResult(w http.ResponseWriter, r *http.Request) {
	taskID := r.URL.Path[len("/api/v1/result/"):]
	if taskID == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	reader, err := h.orchestrator.DownloadResult(r.Context(), taskID)
	if err != nil {
		log.Printf("Error downloading result for task %s: %v", taskID, err)
		http.Error(w, "Result not found or not ready", http.StatusNotFound)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "image/png")
	io.Copy(w, reader)
}

// HealthCheck provides a simple health check endpoint.
func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// StreamEvents handles SSE connections for task progress events.
// It first sends all historical events stored in Redis, then subscribes to the
// Pub/Sub channel for real-time updates.
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

	events, err := h.orchestrator.redis.GetProgressEvents(ctx, taskID)
	if err != nil {
		log.Printf("Error getting progress events for task %s: %v", taskID, err)
	}

	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}
		fmt.Fprintf(w, "data: %s\n\n", data)
	}
	flusher.Flush()

	pubsub, ch := h.orchestrator.redis.SubscribeProgress(ctx, taskID)
	defer pubsub.Close()

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
			fmt.Fprintf(w, "data: %s\n\n", msg.Payload)
			flusher.Flush()
		case <-keepalive.C:
			fmt.Fprintf(w, ": keepalive\n\n")
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

	err := r.ParseMultipartForm(h.maxUploadSize)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Image file is required", http.StatusBadRequest)
		return
	}

	jobTypeStr := r.FormValue("type")
	jobType := parseJobType(jobTypeStr)
	if jobType == jobs.JobType_JOB_TYPE_UNSPECIFIED {
		file.Close()
		http.Error(w, "Invalid job type", http.StatusBadRequest)
		return
	}

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

	data, err := io.ReadAll(file)
	file.Close()
	if err != nil {
		http.Error(w, "Failed to read image", http.StatusInternalServerError)
		return
	}

	taskID := uuid.New().String()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"task_id": taskID,
	})

	go func() {

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		err := h.orchestrator.ProcessImage(ctx, taskID, bytes.NewReader(data), header.Filename, jobType, int64(len(data)), params)
		if err != nil {
			log.Printf("Background processing failed for task %s: %v", taskID, err)
		}
	}()
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
