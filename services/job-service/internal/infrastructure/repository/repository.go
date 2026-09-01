package repository

import (
	"context"

	"github.com/AbhijeetDev102/Nimbus/services/job-service/internal/domain"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type JobRepository struct {
	DB *gorm.DB
}

func NewJobRepository(dns string) (*JobRepository, error) {
	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&domain.Job{})
	db.AutoMigrate(&domain.OutboxEvent{})

	return &JobRepository{
		DB: db,
	}, nil
}

func (r *JobRepository) CreateJob(ctx context.Context, job *domain.Job, event *domain.OutboxEvent) error {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(job).Error; err != nil {
			return err
		}

		if err := tx.Create(event).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *JobRepository) GetJob(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	var job *domain.Job

	if err := r.DB.WithContext(ctx).First(&job, "id=?", id).Error; err != nil {
		return nil, err
	}

	return job, nil
}

func (r *JobRepository) ListJobs(ctx context.Context, req *domain.ListJobsRequest) ([]*domain.Job, int64, error) {
	var jobs []*domain.Job
	var totalCount int64

	// 1. Build base query
	query := r.DB.WithContext(ctx).Model(&domain.Job{})

	// 2. Apply optional filters
	if req.Status != nil && *req.Status != "" {
		query = query.Where("status = ?", *req.Status)
	}
	if req.JobType != nil && *req.JobType != "" {
		query = query.Where("job_type = ?", *req.JobType)
	}

	// 3. Count total matching rows for pagination
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// 4. Safe pagination limits
	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	// 5. Fetch the page ordered by newest first
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&jobs).Error; err != nil {
		return nil, 0, err
	}

	return jobs, totalCount, nil
}

func (r *JobRepository) GetJobStats(ctx context.Context) (*domain.JobStats, error) {
	var stats domain.JobStats

	if err := r.DB.WithContext(ctx).Model(&domain.Job{}).Select(`
	COUNT(*) AS total,
            COUNT(*) FILTER (WHERE status = 'QUEUED') AS queued,
            COUNT(*) FILTER (WHERE status = 'RUNNING') AS running,
            COUNT(*) FILTER (WHERE status = 'COMPLETED') AS completed,
            COUNT(*) FILTER (WHERE status = 'FAILED') AS failed
	`).Scan(&stats).Error; err != nil {
		return nil, err
	}
	return &stats, nil
}
