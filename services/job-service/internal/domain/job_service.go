package domain

import (
	"context"

	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type CreateJobRequest struct {
	ResourceID *uuid.UUID
	JobType    types.JobType
	Parameters datatypes.JSON
}

type ListJobsRequest struct {
	Limit   int
	Offset  int
	Status  *types.JobStatus
	JobType *types.JobType
}

type JobService interface {
	CreateJob(ctx context.Context, req *CreateJobRequest) (*Job, error)
	GetJob(ctx context.Context, id uuid.UUID) (*Job, error)
	ListJobs(ctx context.Context, req *ListJobsRequest) ([]*Job, int64, error)
	GetJobStats(ctx context.Context) (*JobStats, error)
}
