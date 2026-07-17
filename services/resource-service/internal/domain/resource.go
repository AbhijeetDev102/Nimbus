package domain

import (
	"context"
	"time"

	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/google/uuid"
)

type ResourceStatus string

const (
	UploadRequested ResourceStatus = "UPLOAD_REQUESTED"

	Uploaded ResourceStatus = "UPLOADED"

	Processing ResourceStatus = "PROCESSING"

	Completed ResourceStatus = "COMPLETED"

	Failed ResourceStatus = "FAILED"
)

type Resource struct {
	ID               uuid.UUID
	OriginalFileName string
	ObjectKey        string
	SizeBytes        int64
	ContentType      string
	Status           ResourceStatus
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type StorageProvider interface {
	GeneratePSUrl(ctx context.Context, objectKey string, expireIn int64) (*types.UploadUrlResponse, error)
}

type ResourceRepository interface {
	CreateResource(ctx context.Context, resource *Resource) error
}

type ResourceService interface {
	GeneratePSUrl(ctx context.Context, req *types.UploadUrlRequest) (*types.UploadUrlResponse, error)
}
