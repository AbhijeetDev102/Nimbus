package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/AbhijeetDev102/Nimbus/services/job-service/internal/domain"
	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type JobReaper struct {
	repo     domain.JobRepository
	interval time.Duration
}

func NewJobReaper(repo domain.JobRepository, interval time.Duration) *JobReaper {
	if interval <= 0 {
		interval = 15 * time.Second
	}

	return &JobReaper{
		repo:     repo,
		interval: interval,
	}
}

func (r *JobReaper) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("[JobReaper] Background sweeper stopped.")
			return
		case <-ticker.C:
			r.Reap(ctx)
		}

	}
}

func (r *JobReaper) Reap(ctx context.Context) {
	expiredJobs, err := r.repo.FindExpiredRunningJobs(ctx, time.Now())

	if err != nil {
		log.Printf("[JobReaper] Error querying expired jobs: %v", err)
		return
	}

	if len(expiredJobs) == 0 {
		return
	}

	log.Printf("[JobReaper] Found %d expired running jobs. Reaping...", len(expiredJobs))

	for _, job := range expiredJobs {
		isFinalFailure := job.RetryCount+1 >= job.MaxRetries
		nextRetryCount := job.RetryCount + 1
		var eventType types.EventType
		jobCopy := *job
		if isFinalFailure {
			eventType = types.EventJobFailed
			jobCopy.Status = types.JobFailed
			nextRetryCount = job.RetryCount
		} else {
			eventType = types.EventJobCreated
			jobCopy.Status = types.JobQueued
			jobCopy.RetryCount = nextRetryCount
			jobCopy.WorkerID = nil
		}
		payloadJSON, err := json.Marshal(jobCopy)
		if err != nil {
			log.Printf("[JobReaper] Error marshaling payload for job %s: %v", job.ID, err)
			continue
		}
		outboxEvent := &domain.OutboxEvent{
			ID:            uuid.New(),
			AggregateType: "job",
			AggregateID:   job.ID,
			EventType:     eventType,
			Payload:       datatypes.JSON(payloadJSON),
			CreatedAt:     time.Now(),
		}
		if err := r.repo.ReapExpiredJob(ctx, job, nextRetryCount, outboxEvent, isFinalFailure); err != nil {
			log.Printf("[JobReaper] Failed to reap job %s: %v", job.ID, err)
			continue
		}
		if isFinalFailure {
			log.Printf("[JobReaper] Permanently failed expired job %s (max retries exhausted)", job.ID)
		} else {
			log.Printf("[JobReaper] Successfully re-queued expired job %s (retry %d/%d)", job.ID, nextRetryCount, job.MaxRetries)
		}
	}
}
