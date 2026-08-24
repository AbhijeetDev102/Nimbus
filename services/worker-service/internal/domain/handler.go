package domain

import "context"

type JobHandler interface {
	Execute(ctx context.Context, job *Job) (*ExecutionResult, error)
}
