package service

import (
	"context"

	"github.com/AbhijeetDev102/Nimbus/services/resource-service/internal/domain"
	"github.com/AbhijeetDev102/Nimbus/shared/env"
	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/google/uuid"
)

type Service struct {
	storage  domain.StorageProvider
	metadata domain.ResourceRepository
}

func NewService(storgae domain.StorageProvider, metadata domain.ResourceRepository) *Service {
	return &Service{
		storage:  storgae,
		metadata: metadata,
	}
}

func (s *Service) GeneratePSUrl(ctx context.Context, req *types.UploadUrlRequest) (*types.UploadUrlResponse, error) {
	randString := uuid.NewString()
	ID, err := uuid.Parse(randString)
	if err != nil {
		return nil, err
	}

	objectKey := "resource/original/" + randString + ".mp4"

	resourceMetadata := &domain.Resource{
		ID:              ID,
		Name:            req.FileName,
		ResourceType:    domain.ResourceType(req.ResourceType),
		StorageProvider: domain.MinIO,
		Bucket:          env.GetString("MINIO_BUCKET_NAME", ""),
		ObjectKey:       objectKey,
		SizeBytes:       req.FileSize,
		ContentType:     req.ContentType,
		Status:          domain.UploadRequested,
	}

	if err := s.metadata.CreateResource(ctx, resourceMetadata); err != nil {
		return nil, err
	}

	response, err := s.storage.GeneratePSUrl(ctx, objectKey, int64(600))
	if err != nil {
		return nil, err
	}
	return response, nil
}
