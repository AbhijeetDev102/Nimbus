package domain

import "context"

type ProgressUpdate struct {
	JobID    string  `json:"job_id"`
	Progress float64 `json:"progress"` // e.g. 45.2 (%)
	Speed    string  `json:"speed"`    // e.g. "2.4x"
	FPS      float64 `json:"fps"`
	Status   string  `json:"status"` // "RUNNING", "COMPLETED", "FAILED"
}

type ProgressPublisher interface {
	Publish(ctx context.Context, update *ProgressUpdate) error
}
