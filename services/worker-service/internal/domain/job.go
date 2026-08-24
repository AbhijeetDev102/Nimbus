package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/AbhijeetDev102/Nimbus/shared/types"
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
	OutputResourceID *uuid.UUID
	CreatedAt        time.Time
	StartedAt        *time.Time
	CompletedAt      *time.Time
	UpdatedAt        time.Time
}
