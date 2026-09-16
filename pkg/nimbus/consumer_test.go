package nimbus

import (
	"strings"
	"testing"
)

type panicHandler struct{}

func (h *panicHandler) Execute(ctx Context, job *Job) (*ExecutionResult, error) {
	panic("simulated fatal nil-pointer dereference")
}

func TestSafeExecute_PanicRecovery(t *testing.T) {
	w := &Worker{
		dispatcher: NewDispatcher(),
	}

	w.Register("POISON_JOB", &panicHandler{})

	job := &Job{
		JobType: "POISON_JOB",
	}

	result, err := w.safeExecute(nil, job)

	if err == nil {
		t.Fatalf("expected error due to panic, got nil")
	}

	if !strings.Contains(err.Error(), "panic in handler") {
		t.Fatalf("expected error message to mention 'panic in handler', got: %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result on panic, got: %v", result)
	}

	t.Logf("Successfully caught and isolated handler panic: %v", err)
}
