package nimbus

import (
	"fmt"
	"sync"
)

// Dispatcher manages registered JobHandlers and routes incoming jobs dynamically
type Dispatcher struct {
	handlers map[JobType]JobHandler
	mu       sync.RWMutex
}

// NewDispatcher creates a new handler registry
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[JobType]JobHandler),
	}
}

// Register binds a JobType to a specific JobHandler
func (d *Dispatcher) Register(jobType JobType, handler JobHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[jobType] = handler
}

// Dispatch executes the appropriate handler for the incoming job
func (d *Dispatcher) Dispatch(ctx Context, job *Job) (*ExecutionResult, error) {
	d.mu.RLock()
	handler, exists := d.handlers[job.JobType]
	d.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unsupported job type: %s (no handler registered)", job.JobType)
	}

	return handler.Execute(ctx, job)
}
