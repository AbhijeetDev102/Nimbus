package domain

import (
	"context"
	"time"

	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/google/uuid"
)

type Resource struct {
	ID               uuid.UUID
	OriginalFileName string
	ObjectKey        string
	SizeBytes        int64
	ContentType      string
	Status           string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type StorageRepository interface {
	GeneratePSUrl(ctx context.Context, objectKey string, expireIn int64) (*types.UploadUrlResponse, error)
}

type MetaDataRepository interface {
	CreateResource(ctx context.Context, resource *Resource) error
}

type ResourceService interface {
	GeneratePSUrl(ctx context.Context, req *types.UploadUrlRequest) (*types.UploadUrlResponse, error)
}
