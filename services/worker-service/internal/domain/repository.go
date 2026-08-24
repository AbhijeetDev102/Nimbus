package domain

import (
	"context"

	"github.com/google/uuid"
)

type JobRepository interface {
	ClaimJob(ctx context.Context, jobID uuid.UUID, workerID uuid.UUID) (bool, error)

	CompleteJob(ctx context.Context, jobID uuid.UUID, outputResourceID *uuid.UUID) error

	FailJob(ctx context.Context, jobID uuid.UUID, reason string) error
}
