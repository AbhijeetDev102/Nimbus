package events

import (
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/AbhijeetDev102/Nimbus/services/worker-service/internal/domain"
	"github.com/AbhijeetDev102/Nimbus/services/worker-service/internal/platform/dispatcher"
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
