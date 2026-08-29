package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"gorm.io/datatypes"
)

type Job struct {
	ID               uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	ResourceID       uuid.UUID       `json:"resource_id" gorm:"type:uuid;index;not null"`
	JobType          types.JobType   `json:"job_type"`
	Status           types.JobStatus `json:"status"`
	RetryCount       int             `json:"retry_count"`
	MaxRetries       int             `json:"max_retries"`
	WorkerID         *uuid.UUID      `json:"worker_id"`
	Parameters       datatypes.JSON  `json:"parameters" gorm:"type:jsonb"`
	ErrorMessage     *string         `json:"error_message"`
	OutputResourceID *uuid.UUID      `json:"output_resource_id"`
	CreatedAt        time.Time       `json:"created_at"`
	StartedAt        *time.Time      `json:"started_at"`
	CompletedAt      *time.Time      `json:"completed_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type ExecutionResult struct {
	OutputResourceID *uuid.UUID
	Metadata         datatypes.JSON
	Error            error
}
