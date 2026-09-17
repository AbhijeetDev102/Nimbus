package nimbus

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/twmb/franz-go/pkg/kgo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Worker is the cloud-native execution engine that processes jobs
type Worker struct {
	config       Config
	db           *gorm.DB
	kafka        *kgo.Client
	redis        *redis.Client
	dispatcher   *Dispatcher
	startedAt    time.Time
	hostname     string
	currentJobID atomic.Pointer[string]
}

type WorkerHeartbeatPayload struct {
	WorkerID      string  `json:"workerId"`
	Hostname      string  `json:"hostname"`
	Status        string  `json:"status"`
	CurrentJobID  *string `json:"currentJobId,omitempty"`
	LastHeartbeat string  `json:"lastHeartbeat"`
	StartedAt     string  `json:"startedAt"`
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

	hostname, _ := os.Hostname()
	if h := os.Getenv("WORKER_NAME"); h != "" {
		hostname = h
	} else if hostname == "" {
		hostname = "worker-" + cfg.WorkerID.String()[:4]
	}

	return &Worker{
		config:     cfg,
		db:         db,
		kafka:      kafkaClient,
		redis:      redisClient,
		dispatcher: NewDispatcher(),
		startedAt:  time.Now(),
		hostname:   hostname,
	}, nil
}

// Register adds a custom handler for a given JobType
func (w *Worker) Register(jobType JobType, handler JobHandler) {
	w.dispatcher.Register(jobType, handler)
}

// PublishWorkerHeartbeat broadcasts the worker's status to Redis
func (w *Worker) PublishWorkerHeartbeat(ctx context.Context) {
	if w.redis == nil {
		return
	}

	var curJob *string
	status := "IDLE"
	if ptr := w.currentJobID.Load(); ptr != nil && *ptr != "" {
		curJob = ptr
		status = "BUSY"
	}

	startedStr := time.Now().Format(time.RFC3339)
	if !w.startedAt.IsZero() {
		startedStr = w.startedAt.Format(time.RFC3339)
	}
	hostname := w.hostname
	if hostname == "" {
		hostname = "worker-" + w.config.WorkerID.String()[:4]
	}

	payload := WorkerHeartbeatPayload{
		WorkerID:      w.config.WorkerID.String(),
		Hostname:      hostname,
		Status:        status,
		CurrentJobID:  curJob,
		LastHeartbeat: time.Now().Format(time.RFC3339),
		StartedAt:     startedStr,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	_ = w.redis.HSet(ctx, "nimbus:workers", w.config.WorkerID.String(), string(data)).Err()
}

func (w *Worker) runHeartbeatLoop(ctx context.Context) {
	if w.redis == nil {
		return
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// Initial heartbeat immediately
	w.PublishWorkerHeartbeat(ctx)

	for {
		select {
		case <-ctx.Done():
			w.markOffline()
			return
		case <-ticker.C:
			w.PublishWorkerHeartbeat(ctx)
		}
	}
}

func (w *Worker) markOffline() {
	if w.redis == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	payload := WorkerHeartbeatPayload{
		WorkerID:      w.config.WorkerID.String(),
		Hostname:      w.hostname,
		Status:        "OFFLINE",
		LastHeartbeat: time.Now().Format(time.RFC3339),
		StartedAt:     w.startedAt.Format(time.RFC3339),
	}
	if data, err := json.Marshal(payload); err == nil {
		_ = w.redis.HSet(ctx, "nimbus:workers", w.config.WorkerID.String(), string(data)).Err()
	}
}

// Close gracefully terminates Kafka and Redis connections
func (w *Worker) Close() {
	w.markOffline()
	if w.kafka != nil {
		w.kafka.Close()
	}
	if w.redis != nil {
		w.redis.Close()
	}
}

