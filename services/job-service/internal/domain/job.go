package domain

import (
	"time"

	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Job struct {
	ID               uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	ResourceID       *uuid.UUID      `json:"resource_id,omitempty" gorm:"type:uuid;index"`
	JobType          types.JobType   `json:"job_type"`
	Status           types.JobStatus `json:"status"`
	RetryCount       int             `json:"retry_count"`
	MaxRetries       int             `json:"max_retries"`
	WorkerID         *uuid.UUID      `json:"worker_id,omitempty"`
	Parameters       datatypes.JSON  `json:"parameters" gorm:"type:jsonb"`
	ErrorMessage     *string         `json:"error_message,omitempty"`
	OutputResourceID *uuid.UUID      `json:"output_resource_id,omitempty"`
	Metadata         datatypes.JSON  `json:"metadata,omitempty" gorm:"type:jsonb"`

	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type OutboxEvent struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey"`
	AggregateType string          `gorm:"type:varchar(50);not null"` // What entity is this? (e.g., "Job")
	AggregateID   uuid.UUID       `gorm:"type:uuid;not null"`        // The UUID of the specific Job
	EventType     types.EventType // What happened? (e.g., "JobCreated")
	Payload       datatypes.JSON  `gorm:"type:jsonb;not null"` // The JSON data the Worker needs
	CreatedAt     time.Time
}

type JobStats struct {
	Total     int64 `json:"total"`
	Queued    int64 `json:"queued"`
	Running   int64 `json:"running"`
	Completed int64 `json:"completed"`
	Failed    int64 `json:"failed"`
}
