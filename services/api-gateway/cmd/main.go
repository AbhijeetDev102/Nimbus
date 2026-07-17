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

	"github.com/AbhijeetDev102/Nimbus/services/api-gateway/internal/middleware"
	"github.com/AbhijeetDev102/Nimbus/shared/env"
)

var (
	httpAddr = env.GetString("GATEWAY_HTTP_ADDR", ":8081")
)

func main() {
	mux := http.NewServeMux()

	// resuable grpc client and connection
	grpcClient, err := grpcclient.NewResourceServiceClient()
	if err != nil {
		log.Printf("Failed to create resource service Grpc client : %v", err)
		return
	}

	defer grpcClient.Close()

	//passing the grpc client to http handler constructor
	httpHandler := httphandler.NewHttpHandler(grpcClient)

	mux.HandleFunc("POST /resource/upload-url", httpHandler.HandleUploadUrlRequest)
	mux.HandleFunc("POST /jobs/create", httpHandler.HandleCreateJobRequest)

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
