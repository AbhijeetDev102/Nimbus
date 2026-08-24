package repository

import (
	"context"
	"time"

	"github.com/AbhijeetDev102/Nimbus/services/worker-service/internal/domain"
	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostgresJobRepository struct {
	db *gorm.DB
}

func NewPostgresJobRepository(db *gorm.DB) *PostgresJobRepository {
	return &PostgresJobRepository{db: db}
}
func (r *PostgresJobRepository) ClaimJob(ctx context.Context, jobID, workerID uuid.UUID) (bool, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&domain.Job{}).
		Where("id = ? AND status = ?", jobID, types.JobQueued).
		Updates(map[string]interface{}{
			"status":     types.JobRunning,
			"worker_id":  workerID,
			"started_at": &now,
			"updated_at": now,
		})
	if result.Error != nil {
		return false, result.Error
	}
	// If RowsAffected == 1, THIS worker won the race and claimed the job!
	// If RowsAffected == 0, another worker already claimed it.
	return result.RowsAffected > 0, nil
}

func (r *PostgresJobRepository) CompleteJob(ctx context.Context, jobID uuid.UUID, outputResourceID *uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Job{}).
		Where("id=? AND status=?", jobID, types.JobRunning).
		Updates(map[string]interface{}{
			"status":             types.JobCompleted,
			"output_resource_id": outputResourceID,
			"completed_at":       time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *PostgresJobRepository) FailJob(ctx context.Context, jobID uuid.UUID, reason string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Job{}).
		Where("id = ?", jobID).
		Updates(map[string]interface{}{
			"status":        types.JobFailed,
			"error_message": reason,
			"completed_at":  &now,
			"updated_at":    now,
		}).Error
}
