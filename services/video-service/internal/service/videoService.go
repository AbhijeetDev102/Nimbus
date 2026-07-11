package service

import (
	"context"

	"github.com/AbhijeetDev102/Nimbus/services/video-service/internal/domain"
	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/google/uuid"
)

type Service struct {
	repo domain.VideoRepository
}

func NewSerice(repo domain.VideoRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GeneratePSUrl(ctx context.Context, req *types.UploadUrlRequest) (*types.UploadUrlResponse, error) {
	randString := uuid.NewString()

	objectKey := "video/original/" + randString + ".mp4"

	response, err := s.repo.GeneratePSUrl(ctx, objectKey, int64(600))
	if err != nil {
		return nil, err
	}
	return response, nil
}
