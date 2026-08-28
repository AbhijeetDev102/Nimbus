package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/AbhijeetDev102/Nimbus/services/worker-service/internal/infrastructure/events"
	"github.com/AbhijeetDev102/Nimbus/services/worker-service/internal/infrastructure/repository"
	"github.com/AbhijeetDev102/Nimbus/services/worker-service/internal/platform/dispatcher"
	"github.com/AbhijeetDev102/Nimbus/services/worker-service/internal/workloads/video"
	"github.com/AbhijeetDev102/Nimbus/shared/env"
	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	godotenv.Load()

	//postgres
	host := env.GetString("POSTGRES_HOSTNAME", "localhost")
	user := env.GetString("POSTGRES_USER", "postgres")
	password := env.GetString("POSTGRES_PASSWORD", "postgrespassword")
	dbName := env.GetString("POSTGRES_DB_NAME", "nimbus")
	dns := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable", host, user, password, dbName)

	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}

	repo := repository.NewPostgresJobRepository(db)

	// minio

	minioEndpoint := env.GetString("MINIO_ENDPOINT", "localhost:9000")
	minioAccessKey := env.GetString("MINIO_ACCESS_KEY", "minioadmin")
	minioSecretKey := env.GetString("MINIO_SECRET_KEY", "minioadmin")
	minioBucket := env.GetString("MINIO_BUCKET_NAME", "nimbus-bucket")
	minioSSL := env.GetBool("MINIO_USE_SSL", false)

	minioStorage, err := video.NewMinioInstance(minioEndpoint, minioAccessKey, minioSecretKey, minioSSL, minioBucket)
	if err != nil {
		log.Fatalf("Failed to init MinIO: %v", err)
	}
	ffmpegService := video.NewFFmpegService()
	videoHandler := video.NewVideoHandler(minioStorage, ffmpegService)

	//Dispatcher
	disp := dispatcher.NewDispatcher()
	disp.Register(types.VideoTranscode, videoHandler)

	workerID := uuid.New()
	kafkaBrokers := []string{env.GetString("KAFKA_BROKERS", "localhost:9092")}
	kafkaGroup := env.GetString("KAFKA_CONSUMER_GROUP", "nimbus-worker-group")
	kafkaTopic := env.GetString("KAFKA_JOB_TOPIC", "job.events")

	consumer, err := events.NewJobConsumer(kafkaBrokers, kafkaGroup, kafkaTopic, repo, disp, workerID)
	if err != nil {
		log.Fatalf("Failed to init Kafka consumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for SIGINT / SIGTERM
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[Worker %s] Starting Kafka consumer on topic '%s'...", workerID, kafkaTopic)
		if err := consumer.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("Consumer stopped with error: %v", err)
		}
	}()

	<-shutdown
	log.Println("Shutting down worker gracefully...")
	cancel()
	consumer.Close()
	log.Println("Worker exited cleanly.")

}
