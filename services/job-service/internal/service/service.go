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

	jobJSON, err := json.Marshal(job)
	if err != nil {
		return nil, err
	}
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

func (s *jobService) GetJob(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	return s.repo.GetJob(ctx, id)
}

func (s *jobService) ListJobs(ctx context.Context, req *domain.ListJobsRequest) ([]*domain.Job, int64, error) {
	return s.repo.ListJobs(ctx, req)
}

func (s *jobService) GetJobStats(ctx context.Context) (*domain.JobStats, error) {
	return s.repo.GetJobStats(ctx)
}
