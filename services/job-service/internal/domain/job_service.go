package domain

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type CreateJobRequest struct {
	ResourceID uuid.UUID
	JobType    JobType
	Parameters datatypes.JSON
}

type JobService interface {
	CreateJob(ctx context.Context, req *CreateJobRequest) (*Job, error)
}
