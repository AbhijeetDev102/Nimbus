package nimbus

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (w *Worker) ClaimJob(ctx context.Context, jobID, workerID uuid.UUID) (bool, error) {
	now := time.Now()
	result := w.db.WithContext(ctx).
		Model(&Job{}).
		Where("id = ? AND status = ?", jobID, JobQueued).
		Updates(map[string]interface{}{
			"status":     JobRunning,
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

func (w *Worker) CompleteJob(ctx context.Context, jobID uuid.UUID, outputResourceID *uuid.UUID) error {
	result := w.db.WithContext(ctx).
		Model(&Job{}).
		Where("id=? AND status=?", jobID, JobRunning).
		Updates(map[string]interface{}{
			"status":             JobCompleted,
			"output_resource_id": outputResourceID,
			"completed_at":       time.Now(),
		})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (w *Worker) FailJob(ctx context.Context, jobID uuid.UUID, reason string) error {
	now := time.Now()
	return w.db.WithContext(ctx).
		Model(&Job{}).
		Where("id = ?", jobID).
		Updates(map[string]interface{}{
			"status":        JobFailed,
			"error_message": reason,
			"completed_at":  &now,
			"updated_at":    now,
		}).Error
}
