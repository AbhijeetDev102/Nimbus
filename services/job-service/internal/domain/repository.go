package domain

import (
	"context"

	"github.com/google/uuid"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job *Job, event *OutboxEvent) error
	GetJob(ctx context.Context, id uuid.UUID) (*Job, error)
}
