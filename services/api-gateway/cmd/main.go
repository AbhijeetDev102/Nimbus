package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcclient "github.com/AbhijeetDev102/Nimbus/services/api-gateway/internal/grpc_client"
	httphandler "github.com/AbhijeetDev102/Nimbus/services/api-gateway/internal/http_handler"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/AbhijeetDev102/Nimbus/services/api-gateway/internal/middleware"
	"github.com/AbhijeetDev102/Nimbus/shared/env"
)

func main() {
	godotenv.Load()

	httpAddr := env.GetString("GATEWAY_HTTP_ADDR", ":8081")
	mux := http.NewServeMux()

	// resuable grpc client and connection
	resourceGrpcClient, err := grpcclient.NewResourceServiceClient()
	if err != nil {
		log.Printf("Failed to create resource service Grpc client : %v", err)
		return
	}

	defer resourceGrpcClient.Close()

	jobGrpcClient, err := grpcclient.NewJobServiceClient()
	if err != nil {
		log.Printf("Failed to create job service Grpc client : %v", err)
		return
	}

	defer jobGrpcClient.Close()

	// Redis Client for WebSocket Pub/Sub
	redisAddr := env.GetString("REDIS_ADDR", "localhost:6379")
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()

	//passing the grpc client to http handler constructor
	httpHandler := httphandler.NewHttpHandler(resourceGrpcClient, jobGrpcClient, redisClient)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /resource/upload-url", httpHandler.HandleUploadUrlRequest)
	mux.HandleFunc("GET /resource/{id}/download", httpHandler.HandleGetDownloadUrl)
	mux.HandleFunc("POST /jobs/create", httpHandler.HandleCreateJobRequest)
	mux.HandleFunc("GET /jobs/{id}", httpHandler.HandleGetJob)
	mux.HandleFunc("GET /ws/jobs/{id}", httpHandler.HandleJobProgressWS) // <-- Real-time WebSocket!

	server := &http.Server{
		Addr:    httpAddr,
		Handler: middleware.EnableCORS(mux),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Server Listening on %s", httpAddr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {

	case err := <-serverErrors:
		log.Printf("Error starting the server: %v", err)

	case sig := <-shutdown:
		log.Printf("Server is shutting down due to %v signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Could not stop the server gracefullt %v", err)
			server.Close()
		}
	}
}
