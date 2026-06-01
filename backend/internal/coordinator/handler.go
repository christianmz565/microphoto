package coordinator

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/christianmz565/microphoto/pkg/client/metrics"
	"github.com/christianmz565/microphoto/proto/jobs"
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
	defer file.Close()

	jobTypeStr := r.FormValue("type")
	jobType := parseJobType(jobTypeStr)
	if jobType == jobs.JobType_UNKNOWN_TYPE {
		http.Error(w, "Invalid job type", http.StatusBadRequest)
		return
	}

	taskID, err := h.orchestrator.ProcessImage(r.Context(), file, header.Filename, jobType, header.Size)
	if err != nil {
		log.Printf("Error processing image: %v", err)
		http.Error(w, fmt.Sprintf("Error processing image: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"task_id": taskID,
	})
}

// parseJobType converts a string representation of a job type to the corresponding protobuf enum.
func parseJobType(s string) jobs.JobType {
	switch s {
	case "RESIZE":
		return jobs.JobType_RESIZE
	case "GRAYSCALE":
		return jobs.JobType_GRAYSCALE
	case "BLUR":
		return jobs.JobType_BLUR
	case "RECONSTRUCT":
		return jobs.JobType_RECONSTRUCT
	default:
		return jobs.JobType_UNKNOWN_TYPE
	}
}
