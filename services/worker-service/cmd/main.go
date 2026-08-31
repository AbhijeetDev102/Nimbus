package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/AbhijeetDev102/Nimbus/pkg/nimbus"
	"github.com/AbhijeetDev102/Nimbus/services/worker-service/internal/workloads/video"
	"github.com/AbhijeetDev102/Nimbus/shared/env"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	// Initialize the Nimbus Worker Engine from Environment
	worker, err := nimbus.NewWorker(nimbus.ConfigFromEnv())
	if err != nil {
		log.Fatalf("Failed to initialize Nimbus worker: %v", err)
	}
	defer worker.Close()

	// Initialize the Video Trancoder Workload

	minioStorage, err := video.NewMinioInstance(
		env.GetString("MINIO_ENDPOINT", "localhost:9000"),
		env.GetString("MINIO_ACCESS_KEY", "minioadmin"),
		env.GetString("MINIO_SECRET_KEY", "minioadmin"),
		env.GetBool("MINIO_USE_SSL", false),
		env.GetString("MINIO_BUCKET_NAME", "nimbus-bucket"),
	)

	if err != nil {
		log.Fatalf("Failed to initialize MinIO: %v", err)
	}

	ffmpegService := video.NewFFmpegService()
	videoHandler := video.NewVideoHandler(minioStorage, ffmpegService)

	// Register the Workload with the Worker Engine

	worker.Register(nimbus.TypeVideoTranscode, videoHandler)

	// Start the Worker with Gracefull Shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel()
	log.Printf("Starting Nimbus Video Worker...")
	if err := worker.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Worker terminated with error: %v", err)
	}
	log.Println("Worker exited cleanly.")

}
