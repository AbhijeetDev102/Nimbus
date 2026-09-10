package nimbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func (w *Worker) ClaimJob(ctx context.Context, jobID, workerID uuid.UUID) (bool, error) {
	now := time.Now()
	leaseExpiresAt := now.Add(w.config.LeaseDuration)
	result := w.db.WithContext(ctx).
		Model(&Job{}).
		Where("id = ?  AND (status = ? OR (status = ? AND lease_expires_at < ?))", jobID, JobQueued, JobRunning, now).
		Updates(map[string]interface{}{
			"status":           JobRunning,
			"worker_id":        workerID,
			"lease_expires_at": &leaseExpiresAt,
			"started_at":       &now,
			"updated_at":       now,
		})
	if result.Error != nil {
		return false, result.Error
	}
	// If RowsAffected == 1, THIS worker won the race and claimed the job!
	// If RowsAffected == 0, another worker already claimed it.
	return result.RowsAffected > 0, nil
}

func (w *Worker) CompleteJob(ctx context.Context, jobID uuid.UUID, outputResourceID *uuid.UUID, metadata datatypes.JSON, workerID uuid.UUID) error {
	result := w.db.WithContext(ctx).
		Model(&Job{}).
		Where("id=? AND worker_id = ? AND status=? ", jobID, workerID, JobRunning).
		Updates(map[string]interface{}{
			"status":             JobCompleted,
			"output_resource_id": outputResourceID,
			"completed_at":       time.Now(),
			"metadata":           metadata,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("cannot complete job %s: lease lost or stolen by another worker", jobID)
	}

	return nil
}

func (w *Worker) RetryJob(ctx context.Context, job *Job, reason string) error {
	now := time.Now()
	nextRetryCount := job.RetryCount + 1

	jobCopy := *job
	jobCopy.RetryCount = nextRetryCount
	jobCopy.Status = JobQueued
	jobCopy.WorkerID = nil
	jobCopy.ErrorMessage = &reason
	jobCopy.UpdatedAt = now

	jobJSON, err := json.Marshal(jobCopy)
	if err != nil {
		return err
	}

	outboxEvent := &OutboxEvent{
		ID:            uuid.New(),
		AggregateType: "job",
		AggregateID:   job.ID,
		EventType:     EventJobCreated,
		Payload:       datatypes.JSON(jobJSON),
		CreatedAt:     now,
	}

	return w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Reset job state to QUEUED so any worker can claim it
		if err := tx.Model(&Job{}).
			Where("id = ? AND status = ?", job.ID, JobRunning).
			Updates(map[string]interface{}{
				"status":        JobQueued,
				"retry_count":   nextRetryCount,
				"worker_id":     nil,
				"error_message": reason,
				"updated_at":    now,
			}).Error; err != nil {
			return err
		}

		// 2. Insert Outbox Event to emit back to Kafka via Debezium
		return tx.Create(outboxEvent).Error
	})
}

func (w *Worker) FailJob(ctx context.Context, jobID uuid.UUID, reason string, workerID uuid.UUID) error {
	now := time.Now()
	return w.db.WithContext(ctx).
		Model(&Job{}).
		Where("id = ? AND worker_id = ? AND status = ?", jobID, workerID, JobRunning).
		Updates(map[string]interface{}{
			"status":        JobFailed,
			"error_message": reason,
			"completed_at":  &now,
			"updated_at":    now,
		}).Error
}

func (w *Worker) Heartbeat(ctx context.Context, jobID, workerID uuid.UUID) (bool, error) {
	now := time.Now()
	leaseExpiresAt := now.Add(w.config.LeaseDuration)
	result := w.db.WithContext(ctx).
		Model(&Job{}).
		Where("id = ?  AND status = ? AND worker_id= ?", jobID, JobRunning, workerID).
		Updates(map[string]interface{}{
			"lease_expires_at": &leaseExpiresAt,
			"updated_at":       now,
		})
	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}
