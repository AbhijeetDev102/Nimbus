package nimbus

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Worker is the cloud-native execution engine that processes jobs
type Worker struct {
	config     Config
	db         *gorm.DB
	kafka      *kgo.Client
	redis      *redis.Client
	dispatcher *Dispatcher
}

// NewWorker initializes the complete Nimbus worker runtime
func NewWorker(cfg Config) (*Worker, error) {
	// 1. Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("nimbus worker: failed to connect to postgres: %w", err)
	}
	// 2. Connect to Kafka
	kafkaClient, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.KafkaBrokers...),
		kgo.ConsumerGroup(cfg.KafkaGroup),
		kgo.ConsumeTopics(cfg.KafkaTopic),
		kgo.DisableAutoCommit(),
		kgo.BlockRebalanceOnPoll(),
	)
	if err != nil {
		return nil, fmt.Errorf("nimbus worker: failed to connect to kafka: %w", err)
	}
	// 3. Connect to Redis Pub/Sub
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
	return &Worker{
		config:     cfg,
		db:         db,
		kafka:      kafkaClient,
		redis:      redisClient,
		dispatcher: NewDispatcher(),
	}, nil
}

// Register adds a custom handler for a given JobType
func (w *Worker) Register(jobType JobType, handler JobHandler) {
	w.dispatcher.Register(jobType, handler)
}

// Close gracefully terminates Kafka and Redis connections
func (w *Worker) Close() {
	if w.kafka != nil {
		w.kafka.Close()
	}
	if w.redis != nil {
		w.redis.Close()
	}
}
