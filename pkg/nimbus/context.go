package nimbus

import (
	"context"
	"sync"
	"time"
)

// Context wraps standard Go context and provides Nimbus execution helpers
type Context interface {
	context.Context
	ReportProgress(percent float64)
	ReportProgressDetails(percent float64, message string, metadata map[string]any)
}

type jobContext struct {
	context.Context
	jobID         string
	publisher     ProgressPublisher
	lastPublished time.Time
	throttleMs    time.Duration
	mu            sync.Mutex
}

type ProgressPublisher interface {
	Publish(ctx context.Context, update *ProgressUpdate) error
}

func NewJobContext(ctx context.Context, jobID string, publisher ProgressPublisher) Context {
	return &jobContext{
		Context:    ctx,
		jobID:      jobID,
		publisher:  publisher,
		throttleMs: 250 * time.Millisecond, // Automatically throttles to at most 4 updates/sec
	}
}

func (c *jobContext) ReportProgress(percent float64) {
	c.ReportProgressDetails(percent, "", nil)
}
func (c *jobContext) ReportProgressDetails(percent float64, message string, metadata map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if percent > 100.0 {
		percent = 100.0
	} else if percent < 0.0 {
		percent = 0.0
	}
	if time.Since(c.lastPublished) >= c.throttleMs || percent >= 100.0 {
		c.lastPublished = time.Now()
		if c.publisher != nil {
			_ = c.publisher.Publish(c.Context, &ProgressUpdate{
				JobID:    c.jobID,
				Progress: percent,
				Message:  message,
				Metadata: metadata,
				Status:   "RUNNING",
			})
		}
	}
}
