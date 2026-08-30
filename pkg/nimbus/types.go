package nimbus

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// JobType represents the type of workload to execute
type JobType string

const (
	TypeVideoTranscode JobType = "VIDEO_TRANSCODE"
	TypeImageResize    JobType = "IMAGE_RESIZE"
)

// JobStatus represents the state of a job in the lifecycle
type JobStatus string

const (
	JobQueued    JobStatus = "QUEUED"
	JobRunning   JobStatus = "RUNNING"
	JobCompleted JobStatus = "COMPLETED"
	JobFailed    JobStatus = "FAILED"
	JobCancelled JobStatus = "CANCELLED"
	JobRetrying  JobStatus = "RETRYING"
)

// Job represents a generic task submitted to the Nimbus platform
type Job struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	ResourceID       uuid.UUID      `json:"resource_id" gorm:"type:uuid;index;not null"`
	JobType          JobType        `json:"job_type"`
	Status           JobStatus      `json:"status"`
	RetryCount       int            `json:"retry_count"`
	MaxRetries       int            `json:"max_retries"`
	Parameters       datatypes.JSON `json:"parameters" gorm:"type:jsonb"`
	WorkerID         *uuid.UUID     `json:"worker_id,omitempty"`
	ErrorMessage     *string        `json:"error_message,omitempty"`
	OutputResourceID *uuid.UUID     `json:"output_resource_id,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	StartedAt        *time.Time     `json:"started_at,omitempty"`
	CompletedAt      *time.Time     `json:"completed_at,omitempty"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// ExecutionResult is returned by a JobHandler upon completion
type ExecutionResult struct {
	OutputResourceID *uuid.UUID     `json:"output_resource_id,omitempty"`
	Metadata         datatypes.JSON `json:"metadata,omitempty"`
}

// ProgressUpdate represents a real-time progress event
type ProgressUpdate struct {
	JobID    string         `json:"job_id"`
	Progress float64        `json:"progress"`           // Universal 0.0 to 100.0 (%)
	Message  string         `json:"message,omitempty"`  // e.g. "Processing chunk 3/10"
	Metadata map[string]any `json:"metadata,omitempty"` // {"fps": 30, "speed": "2.4x"} or {"current": 12, "total": 50}
	Status   string         `json:"status"`
}

// JobHandler is the public interface external developers implement for custom workloads
type JobHandler interface {
	Execute(ctx Context, job *Job) (*ExecutionResult, error)
}
