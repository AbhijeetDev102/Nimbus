package main

import (
	"context"
	"fmt"
	"log"
	"net"

	grpchandler "github.com/AbhijeetDev102/Nimbus/services/video-service/internal/infrastructure/grpc_handler"
	"github.com/AbhijeetDev102/Nimbus/services/video-service/internal/infrastructure/repository"
	"github.com/AbhijeetDev102/Nimbus/shared/env"
	"github.com/joho/godotenv"

	"github.com/AbhijeetDev102/Nimbus/services/video-service/internal/service"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, relying on system environment variables")
	}
	ctx := context.Background()

	// 1. Connection Configurations
	// Note: Use the API port (usually 9000), NOT the Console Web UI port (usually 9001).
	endpoint := env.GetString("MINIO_ENDPOINT", "localhost:9000")
	accessKeyID := env.GetString("MINIO_ACCESS_KEY", "")     // Replace with your self-hosted access key
	secretAccessKey := env.GetString("MINIO_SECRET_KEY", "") // Replace with your self-hosted secret key
	useSSL := env.GetBool("MINIO_USE_SSL", false)            // Set to true if your self-hosted instance uses HTTPS

	// 2. Initialize the MinIO Client
	bucketName := env.GetString("MINIO_BUCKET_NAME", "video-bucket")
	minioInstance, err := repository.NewMinioInstance(endpoint, accessKeyID, secretAccessKey, useSSL, bucketName)
	if err != nil {
		log.Fatalln("Failed to initialize MinIO client:", err)
	}
	fmt.Println("Successfully connected to MinIO instance.")

	//postgress Instance creation

	hostName := env.GetString("POSTGRES_HOSTNAME", "localhost")
	user := env.GetString("POSTGRES_USER", "")
	password := env.GetString("POSTGRES_PASSWORD", "")
	db_name := env.GetString("POSTGRES_DB_NAME", "")
	dns := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable", hostName, user, password, db_name)

	postgresInstance, err := repository.NewPostgresInstance(dns)
	if err != nil {
		log.Fatalln("Failed to connect to Postgres:", err)
	}
	log.Println("SUccessfully connected to Postgres instance")
	srv := service.NewSerice(minioInstance, postgresInstance)

	// grpc server initialization
	grpcServer := grpc.NewServer()
	grpchandler.NewGrpcHandler(grpcServer, srv)

	// 3. Checking if the bucket exsits or not if now creat one with the given name
	// 4. Create Bucket if it doesn't exist
	err = minioInstance.EnsureBucketExists(ctx)

	if err != nil {
		log.Fatalln("Failed to ensure video bucket exists:", err)
	}

	fmt.Printf("Successfully ensured bucket '%s' is ready.\n", bucketName)

	lis, err := net.Listen("tcp", ":9093")

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down the server...")
	grpcServer.GracefulStop()

}
