package main

import (
	"context"
	"fmt"
	"log"
	"net"

	grpchandler "github.com/AbhijeetDev102/Nimbus/services/job-service/internal/infrastructure/grpc_handler"
	"github.com/AbhijeetDev102/Nimbus/services/job-service/internal/infrastructure/repository"
	"github.com/AbhijeetDev102/Nimbus/services/job-service/internal/service"
	"github.com/AbhijeetDev102/Nimbus/shared/env"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, relying on system environment variables")
	}
	ctx := context.Background()

	hostName := env.GetString("POSTGRES_HOSTNAME", "localhost")
	user := env.GetString("POSTGRES_USER", "")
	password := env.GetString("POSTGRES_PASSWORD", "")
	db_name := env.GetString("POSTGRES_DB_NAME", "")
	dns := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable", hostName, user, password, db_name)

	postgresInstance, err := repository.NewJobRepository(dns)
	if err != nil {
		log.Fatalln("Failed to connect to Postgres:", err)
	}
	log.Println("SUccessfully connected to Postgres instance")
	srv := service.NewJobService(postgresInstance)

	// grpc server initialization
	grpcServer := grpc.NewServer()
	grpchandler.NewGrpcHandler(grpcServer, srv)

	lis, err := net.Listen("tcp", ":9094")

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down the server...")
	grpcServer.GracefulStop()

}
