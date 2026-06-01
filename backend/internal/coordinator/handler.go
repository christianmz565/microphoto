package coordinator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/christianmz565/microphoto/pkg/client/metrics"
	"github.com/christianmz565/microphoto/proto/jobs"
)

type HTTPHandler struct {
	orchestrator  *Orchestrator
	metrics       *metrics.Metrics
	maxUploadSize int64
}

func NewHTTPHandler(orch *Orchestrator, m *metrics.Metrics, maxUploadSize int64) *HTTPHandler {
	return &HTTPHandler{
		orchestrator:  orch,
		metrics:       m,
		maxUploadSize: maxUploadSize,
	}
}

func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.HealthCheck)
	mux.HandleFunc("/api/v1/process", h.ProcessImage)
}

func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

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
