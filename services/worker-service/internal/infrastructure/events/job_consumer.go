package events

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/AbhijeetDev102/Nimbus/services/worker-service/internal/domain"
	"github.com/AbhijeetDev102/Nimbus/services/worker-service/internal/platform/dispatcher"
	"github.com/AbhijeetDev102/Nimbus/shared/types"
)

type JobConsumer struct {
	client     *kgo.Client
	repo       domain.JobRepository
	dispatcher *dispatcher.Dispatcher
	workerID   uuid.UUID
}

func NewJobConsumer(
	brokers []string,
	groupID string,
	topic string,
	repo domain.JobRepository,
	disp *dispatcher.Dispatcher,
	workerID uuid.UUID,
) (*JobConsumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topic),
		kgo.DisableAutoCommit(),
		kgo.BlockRebalanceOnPoll(),
	)
	if err != nil {
		return nil, err
	}
	return &JobConsumer{
		client:     client,
		repo:       repo,
		dispatcher: disp,
		workerID:   workerID,
	}, nil
}

func (c *JobConsumer) Start(ctx context.Context) error {
	for {
		//poll fetches new records arives or context is cancelled

		fetches := c.client.PollFetches(ctx)
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
			c.processRecord(ctx, record)
		}
	}
}

func (c *JobConsumer) processRecord(ctx context.Context, record *kgo.Record) {
	// check header
	var eventType string

	for _, h := range record.Headers {
		if h.Key == "eventType" {
			eventType = string(h.Value)
			break
		}
	}

	if eventType != string(types.EventJobCreated) {
		c.client.CommitRecords(ctx, record)
		return
	}

	//Unmarshal Payload (Debezium Outbox gave us clean JSON)

	var job domain.Job

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
		c.client.CommitRecords(ctx, record)
		return
	}

	claimed, err := c.repo.ClaimJob(ctx, job.ID, c.workerID)
	if err != nil {
		log.Printf("Failed to claim job %s: %v", job.ID, err)
		return
	}

	if !claimed {
		log.Printf("Job %s already claimed by another worker, skipping", job.ID)
		c.client.CommitRecords(ctx, record)
		return
	}

	// dispatch to workload handler (ffmpeg, etc)

	result, err := c.dispatcher.Dispatch(ctx, &job)

	//Record Final State in PostgreSQL

	if err != nil || (result != nil && result.Error != nil) {
		errMsg := "execution failed"
		if err != nil {
			errMsg = err.Error()
		} else if result.Error != nil {
			errMsg = result.Error.Error()
		}
		c.repo.FailJob(ctx, job.ID, errMsg)
	} else {
		var outputID *uuid.UUID
		if result != nil {
			outputID = result.OutputResourceID
		}
		c.repo.CompleteJob(ctx, job.ID, outputID)
	}

	// Commit kafka offset
	c.client.CommitRecords(ctx, record)
}

func (c *JobConsumer) Close() {
	c.client.Close()
}
