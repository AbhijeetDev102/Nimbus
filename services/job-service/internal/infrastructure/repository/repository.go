package repository

import (
	"context"

	"github.com/AbhijeetDev102/Nimbus/services/job-service/internal/domain"
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
