package domain

import (
	"time"

	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Job struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	ResourceID       uuid.UUID `gorm:"type:uuid;index;not null"` // Connects back to the Resource Service!
	JobType          types.JobType
	Status           types.JobStatus
	RetryCount       int
	MaxRetries       int
	WorkerID         *uuid.UUID
	Parameters       datatypes.JSON `gorm:"type:jsonb"` // Stores dynamic instructions (like what resolution to transcode to)
	ErrorMesssage    *string
	OutputResourceID *uuid.UUID
	CreatedAt        time.Time
	StartedAt        *time.Time
	CompletedAt      *time.Time
	UpdatedAt        time.Time
}

type OutboxEvent struct {
	ID            uuid.UUID       `gorm:"type:uuid;primaryKey"`
	AggregateType string          `gorm:"type:varchar(50);not null"` // What entity is this? (e.g., "Job")
	AggregateID   uuid.UUID       `gorm:"type:uuid;not null"`        // The UUID of the specific Job
	EventType     types.EventType // What happened? (e.g., "JobCreated")
	Payload       datatypes.JSON  `gorm:"type:jsonb;not null"` // The JSON data the Worker needs
	CreatedAt     time.Time
}
