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

type JobService interface {
	CreateJob(ctx context.Context, req *CreateJobRequest) (*Job, error)
	GetJob(ctx context.Context, id uuid.UUID) (*Job, error)
}
