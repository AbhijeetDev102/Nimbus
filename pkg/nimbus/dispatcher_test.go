package nimbus

import (
	"fmt"
	"sync"
	"testing"
)

type mockHandler struct{}

func (m *mockHandler) Execute(ctx Context, job *Job) (*ExecutionResult, error) {
	return &ExecutionResult{}, nil
}

func TestDispatcher_ConcurrentAccess(t *testing.T) {
	dispatcher := NewDispatcher()
	var wg sync.WaitGroup

	numGoroutines := 50
	operationsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				jobType := JobType(fmt.Sprintf("JOB_TYPE_%d", j%10))
				dispatcher.Register(jobType, &mockHandler{})
			}
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				jobType := JobType(fmt.Sprintf("JOB_TYPE_%d", j%10))
				job := &Job{JobType: jobType}

				_, _ = dispatcher.Dispatch(nil, job)
			}
		}(i)

	}
	wg.Wait()
}
