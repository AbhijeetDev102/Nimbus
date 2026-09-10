package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job *Job, event *OutboxEvent) error
	GetJob(ctx context.Context, id uuid.UUID) (*Job, error)
	ListJobs(ctx context.Context, req *ListJobsRequest) ([]*Job, int64, error)
	GetJobStats(ctx context.Context) (*JobStats, error)
	FindExpiredRunningJobs(ctx context.Context, now time.Time) ([]*Job, error)
	ReapExpiredJob(ctx context.Context, job *Job, nextRetryCount int, outboxEvent *OutboxEvent, isFinalFailure bool) error
}
