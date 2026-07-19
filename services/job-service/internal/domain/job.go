package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type JobType string

const (
	VideoTranscode JobType = "VIDEO_TRANSCODE"
	ImageResize    JobType = "IMAGE_RESIZE"
)

type JobStatus string

const (
	JobQueued    JobStatus = "QUEUED"
	JobRunning   JobStatus = "RUNNING"
	JobCompleted JobStatus = "COMPLETED"
	JobFailed    JobStatus = "FAILED"
	JobCancelled JobStatus = "CANCELLED"
	JobRetrying  JobStatus = "RETRYING"
)

type Job struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	ResourceID       uuid.UUID `gorm:"type:uuid;index;not null"` // Connects back to the Resource Service!
	JobType          JobType
	Status           JobStatus
	RetryCount       int
	MaxRetries       int
	WorkerID         *uuid.UUID
	Parameters       datatypes.JSON `gorm:"type:jsonb"` // Stores dynamic instructions (like what resolution to transcode to)
	OutputResourceID *uuid.UUID
	CreatedAt        time.Time
	StartedAt        *time.Time
	CompletedAt      *time.Time
	UpdatedAt        time.Time
}

type EventType string

const (
	EventJobCreated   EventType = "JobCreated"
	EventJobCompleted EventType = "JobCompleted"
	EventJobFailed    EventType = "JobFailed"
)

type OutboxEvent struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	AggregateType string         `gorm:"type:varchar(50);not null"` // What entity is this? (e.g., "Job")
	AggregateID   uuid.UUID      `gorm:"type:uuid;not null"`        // The UUID of the specific Job
	EventType     EventType      // What happened? (e.g., "JobCreated")
	Payload       datatypes.JSON `gorm:"type:jsonb;not null"` // The JSON data the Worker needs
	CreatedAt     time.Time
}
