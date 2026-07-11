package domain

import (
	"context"

	"github.com/AbhijeetDev102/Nimbus/shared/types"
)

type VideoRepository interface {
	GeneratePSUrl(ctx context.Context, objectKey string, expireIn int64) (*types.UploadUrlResponse, error)
}

type VideoService interface {
	GeneratePSUrl(ctx context.Context, req *types.UploadUrlRequest) (*types.UploadUrlResponse, error)
}
