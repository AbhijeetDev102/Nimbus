package service

import (
	"context"
	"encoding/json"

	"github.com/AbhijeetDev102/Nimbus/services/job-service/internal/domain"
	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type jobService struct {
	repo domain.JobRepository
}

func NewJobService(repo domain.JobRepository) *jobService {
	return &jobService{
		repo: repo,
	}
}

func (s *jobService) CreateJob(ctx context.Context, req *domain.CreateJobRequest) (*domain.Job, error) {

	jobId := uuid.New()
	eventId := uuid.New()

	job := &domain.Job{
		ID:         jobId,
		ResourceID: req.ResourceID,
		JobType:    req.JobType,
		Status:     types.JobQueued,
		RetryCount: 0,
		MaxRetries: 3,
		Parameters: req.Parameters,
	}

	jobJSON, _ := json.Marshal(job)
	outBoxEvent := &domain.OutboxEvent{
		ID:            eventId,
		AggregateType: "job",
		AggregateID:   jobId,
		EventType:     types.EventJobCreated,
		Payload:       datatypes.JSON(jobJSON),
	}

	Joberr := s.repo.CreateJob(ctx, job, outBoxEvent)

	if Joberr != nil {
		return nil, Joberr
	}
	return job, nil

}
