package nimbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
	"gorm.io/datatypes"
)

func (w *Worker) Start(ctx context.Context) error {
	for {
		//poll fetches new records arives or context is cancelled

		fetches := w.kafka.PollFetches(ctx)
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// handle broker/ fetch errors

		if errs := fetches.Errors(); len(errs) > 0 {
			for _, fetchErr := range errs {
				log.Printf("Kafka fetch error: %v", fetchErr.Err)
			}
			continue
		}

		//Iterate through fetched records
		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			w.processRecord(ctx, record)
		}
	}
}

func (w *Worker) processRecord(ctx context.Context, record *kgo.Record) {
	// check header
	var eventType string

	for _, h := range record.Headers {
		if h.Key == "eventType" {
			eventType = string(h.Value)
			break
		}
	}

	if eventType != string(EventJobCreated) {
		w.kafka.CommitRecords(ctx, record)
		return
	}

	//Unmarshal Payload (Debezium Outbox gave us clean JSON)

	var job Job

	// 1. Try unmarshaling directly (raw JSON)
	if err := json.Unmarshal(record.Value, &job); err != nil || job.ID == uuid.Nil {
		// 2. Fallback: check if wrapped in Kafka Connect Schema envelope
		var envelope struct {
			Payload string `json:"payload"`
		}
		if err := json.Unmarshal(record.Value, &envelope); err == nil && envelope.Payload != "" {
			_ = json.Unmarshal([]byte(envelope.Payload), &job)
		}
	}

	if job.ID == uuid.Nil {
		log.Printf("Failed to unmarshal valid job payload from Kafka message")
		w.kafka.CommitRecords(ctx, record)
		return
	}

	claimed, err := w.ClaimJob(ctx, job.ID, w.config.WorkerID)
	if err != nil {
		log.Printf("Failed to claim job %s: %v", job.ID, err)
		return
	}

	if !claimed {
		log.Printf("Job %s already claimed by another worker, skipping", job.ID)
		w.kafka.CommitRecords(ctx, record)
		return
	}

	// dispatch to workload handler (ffmpeg, etc)
	jobCtx := NewJobContext(ctx, job.ID.String(), newRedisPublisher(w.redis))
	result, err := w.safeExecute(jobCtx, &job)

	//Record Final State in PostgreSQL

	if err != nil {
		if job.RetryCount+1 < job.MaxRetries {
			log.Printf("[Job %s] Attempt %d/%d failed: %v. Re-queueing via Outbox...", job.ID, job.RetryCount+1, job.MaxRetries, err)
			_ = w.RetryJob(ctx, &job, err.Error())
		} else {
			log.Printf("[Job %s] Max retries (%d) exhausted. Failing permanently: %v", job.ID, job.MaxRetries, err)
			_ = w.FailJob(ctx, job.ID, err.Error())
		}
	} else {
		var outputID *uuid.UUID
		var metadata datatypes.JSON
		if result != nil {
			outputID = result.OutputResourceID
			metadata = result.Metadata
		}
		w.CompleteJob(ctx, job.ID, outputID, metadata)
	}

	// Commit kafka offset
	w.kafka.CommitRecords(ctx, record)
}

func (w *Worker) safeExecute(ctx Context, job *Job) (result *ExecutionResult, err error) {
	// Defer a recover function to catch panics in user handlers
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in handler: %v", r)
		}
	}()

	return w.dispatcher.Dispatch(ctx, job)
}
