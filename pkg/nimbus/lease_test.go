package nimbus

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite db: %v", err)
	}

	if err := db.AutoMigrate(&Job{}); err != nil {
		t.Fatalf("failed to migrate Job schema: %v", err)
	}

	return db
}

func TestLease_StealingAndFencing(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)

	worker1ID := uuid.New()
	worker2ID := uuid.New()

	w1 := &Worker{
		config: Config{WorkerID: worker1ID, LeaseDuration: 30 * time.Second},
		db:     db,
	}
	w2 := &Worker{
		config: Config{WorkerID: worker2ID, LeaseDuration: 30 * time.Second},
		db:     db,
	}

	jobID := uuid.New()
	now := time.Now()
	job := &Job{
		ID:         jobID,
		JobType:    "TEST_WORKLOAD",
		Status:     JobQueued,
		RetryCount: 0,
		MaxRetries: 3,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := db.Create(job).Error; err != nil {
		t.Fatalf("failed to seed test job: %v", err)
	}

	// 1. Fresh Claim: Worker 1 claims QUEUED job
	claimed, err := w1.ClaimJob(ctx, jobID, worker1ID)
	if err != nil || !claimed {
		t.Fatalf("expected Worker 1 to claim queued job, got claimed=%v, err=%v", claimed, err)
	}
	t.Log("[PASS] Step 1: Worker 1 successfully claimed QUEUED job")

	// 2. Active Collision: Worker 2 tries to claim while Worker 1's lease is active
	claimed, err = w2.ClaimJob(ctx, jobID, worker2ID)
	if err != nil {
		t.Fatalf("unexpected error on collision check: %v", err)
	}
	if claimed {
		t.Fatalf("split-brain detected! Worker 2 claimed a job while Worker 1 held an active lease")
	}
	t.Log("[PASS] Step 2: Worker 2 was safely rejected from claiming an active lease")

	// 3. Heartbeat: Worker 1 extends its active lease
	renewed, err := w1.Heartbeat(ctx, jobID, worker1ID)
	if err != nil || !renewed {
		t.Fatalf("expected heartbeat renewal to succeed, got renewed=%v, err=%v", renewed, err)
	}
	t.Log("[PASS] Step 3: Worker 1 successfully renewed its lease via Heartbeat")

	// 4. Worker 1 crashes / lease expires: Artificially expire the lease in DB
	expiredTime := time.Now().Add(-10 * time.Minute)
	if err := db.Model(&Job{}).Where("id = ?", jobID).Update("lease_expires_at", expiredTime).Error; err != nil {
		t.Fatalf("failed to expire lease in test DB: %v", err)
	}

	// Worker 2 attempts to claim the expired job -> MUST STEAL LEASE
	stolen, err := w2.ClaimJob(ctx, jobID, worker2ID)
	if err != nil || !stolen {
		t.Fatalf("expected Worker 2 to steal expired lease, got stolen=%v, err=%v", stolen, err)
	}
	t.Log("[PASS] Step 4: Worker 2 successfully stole the expired lease from dead Worker 1")

	// 5. Split-Brain Fencing: Resurrected Worker 1 wakes up and attempts to CompleteJob
	err = w1.CompleteJob(ctx, jobID, nil, nil, worker1ID)
	if err == nil {
		t.Fatalf("expected CompleteJob to fail for stale Worker 1, but it succeeded!")
	}
	if !strings.Contains(err.Error(), "lease lost or stolen") {
		t.Fatalf("expected error message to mention lease lost or stolen, got: %v", err)
	}
	t.Logf("[PASS] Step 5: Fencing blocked resurrected Worker 1: %v", err)

	// 6. Active Worker 2 completes the job successfully
	err = w2.CompleteJob(ctx, jobID, nil, nil, worker2ID)
	if err != nil {
		t.Fatalf("expected active Worker 2 to complete job, got: %v", err)
	}

	var finalJob Job
	db.First(&finalJob, "id = ?", jobID)
	if finalJob.Status != JobCompleted {
		t.Fatalf("expected final status COMPLETED, got %s", finalJob.Status)
	}
	t.Log("[PASS] Step 6: Active Worker 2 completed job and state is COMPLETED")
}
