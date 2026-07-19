package domain

import (
	"context"
)

type JobRepository interface {
	CreateJob(ctx context.Context, job *Job, event *OutboxEvent) error
}
