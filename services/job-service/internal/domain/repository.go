package domain

import (
	"context"

	"github.com/google/uuid"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job *Job, event *OutboxEvent) error
	GetJob(ctx context.Context, id uuid.UUID) (*Job, error)
	ListJobs(ctx context.Context, req *ListJobsRequest) ([]*Job, int64, error)
	GetJobStats(ctx context.Context) (*JobStats, error)
}
