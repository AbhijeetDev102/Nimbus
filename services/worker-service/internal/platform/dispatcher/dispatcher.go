package dispatcher

import (
	"context"
	"fmt"

	"github.com/AbhijeetDev102/Nimbus/services/worker-service/internal/domain"
	"github.com/AbhijeetDev102/Nimbus/shared/types"
)

type Dispatcher struct {
	handlers map[types.JobType]domain.JobHandler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[types.JobType]domain.JobHandler),
	}
}

func (d *Dispatcher) Register(jobType types.JobType, handler domain.JobHandler) {
	d.handlers[jobType] = handler
}

func (d *Dispatcher) Dispatch(ctx context.Context, job *domain.Job) (*domain.ExecutionResult, error) {
	value, ok := d.handlers[job.JobType]
	if !ok {
		return nil, fmt.Errorf("unsupported job type: %s", job.JobType)
	}

	return value.Execute(ctx, job)
}
