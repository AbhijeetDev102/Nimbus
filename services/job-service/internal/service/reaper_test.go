package service

import (
	"context"
	"testing"
	"time"

	"github.com/AbhijeetDev102/Nimbus/services/job-service/internal/domain"
	"github.com/AbhijeetDev102/Nimbus/services/job-service/internal/infrastructure/repository"
	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func setupTestRepo(t *testing.T) (*repository.JobRepository, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test sqlite db: %v", err)
	}

	if err := db.AutoMigrate(&domain.Job{}, &domain.OutboxEvent{}); err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	return &repository.JobRepository{DB: db}, db
}

func TestJobReaper_ReapExpiredJobs_Retry(t *testing.T) {
	ctx := context.Background()
	repo, db := setupTestRepo(t)

	reaper := NewJobReaper(repo, 15*time.Second)

	jobID := uuid.New()
	workerID := uuid.New()
	expiredTime := time.Now().Add(-10 * time.Minute)
	now := time.Now()

	job := &domain.Job{
		ID:             jobID,
		JobType:        types.VideoTranscode,
		Status:         types.JobRunning,
		RetryCount:     0,
		MaxRetries:     3,
		WorkerID:       &workerID,
		LeaseExpiresAt: &expiredTime,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := db.Create(job).Error; err != nil {
		t.Fatalf("failed to seed job: %v", err)
	}

	// Run the Reaper loop
	reaper.Reap(ctx)

	// 1. Verify Job State in Database
	var updatedJob domain.Job
	if err := db.First(&updatedJob, "id = ?", jobID).Error; err != nil {
		t.Fatalf("failed to find job after reaping: %v", err)
	}

	if updatedJob.Status != types.JobQueued {
		t.Fatalf("expected status %s, got %s", types.JobQueued, updatedJob.Status)
	}
	if updatedJob.RetryCount != 1 {
		t.Fatalf("expected retry_count 1, got %d", updatedJob.RetryCount)
	}
	if updatedJob.WorkerID != nil {
		t.Fatalf("expected worker_id to be nil, got %v", updatedJob.WorkerID)
	}
	t.Log("[PASS] Step 1: Job was reset to QUEUED, retry count incremented, and worker_id cleared")

	// 2. Verify Outbox Event created for Debezium / Kafka
	var outboxEvent domain.OutboxEvent
	if err := db.First(&outboxEvent, "aggregate_id = ?", jobID).Error; err != nil {
		t.Fatalf("failed to find outbox event for reaped job: %v", err)
	}

	if outboxEvent.EventType != types.EventJobCreated {
		t.Fatalf("expected event type %s, got %s", types.EventJobCreated, outboxEvent.EventType)
	}
	t.Log("[PASS] Step 2: Transactional Outbox event 'JobCreated' emitted for Kafka dispatch")
}

func TestJobReaper_ReapExpiredJobs_MaxRetriesExhausted(t *testing.T) {
	ctx := context.Background()
	repo, db := setupTestRepo(t)

	reaper := NewJobReaper(repo, 15*time.Second)

	jobID := uuid.New()
	workerID := uuid.New()
	expiredTime := time.Now().Add(-10 * time.Minute)
	now := time.Now()

	// Seed job with retry_count = 2 and max_retries = 3 (this next attempt will exhaust retries)
	job := &domain.Job{
		ID:             jobID,
		JobType:        types.VideoTranscode,
		Status:         types.JobRunning,
		RetryCount:     2,
		MaxRetries:     3,
		WorkerID:       &workerID,
		LeaseExpiresAt: &expiredTime,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := db.Create(job).Error; err != nil {
		t.Fatalf("failed to seed job: %v", err)
	}

	// Run the Reaper
	reaper.Reap(ctx)

	// 1. Verify Job is marked FAILED
	var updatedJob domain.Job
	if err := db.First(&updatedJob, "id = ?", jobID).Error; err != nil {
		t.Fatalf("failed to find job after reaping: %v", err)
	}

	if updatedJob.Status != types.JobFailed {
		t.Fatalf("expected status %s, got %s", types.JobFailed, updatedJob.Status)
	}
	if updatedJob.CompletedAt == nil {
		t.Fatalf("expected completed_at timestamp to be set on failure")
	}
	t.Log("[PASS] Step 1: Job was permanently marked FAILED after exhausting retries")

	// 2. Verify Outbox Event is 'JobFailed'
	var outboxEvent domain.OutboxEvent
	if err := db.First(&outboxEvent, "aggregate_id = ?", jobID).Error; err != nil {
		t.Fatalf("failed to find outbox event for failed job: %v", err)
	}

	if outboxEvent.EventType != types.EventJobFailed {
		t.Fatalf("expected event type %s, got %s", types.EventJobFailed, outboxEvent.EventType)
	}
	t.Log("[PASS] Step 2: Transactional Outbox event 'JobFailed' emitted for Kafka dispatch")
}
