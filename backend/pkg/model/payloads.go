package model

type ProgressPayload struct {
	JobID     string  `json:"job_id"`
	Progress  float64 `json:"progress"`
	Status    string  `json:"status"`
	Message   string  `json:"message,omitempty"`
	ResultURL string  `json:"result_url,omitempty"`
}

type EventNotification struct {
	Type      string `json:"type"`
	Payload   any    `json:"payload"`
	Timestamp int64  `json:"timestamp"`
}
