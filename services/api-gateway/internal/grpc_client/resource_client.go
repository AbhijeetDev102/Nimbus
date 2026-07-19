package grpcclient

import (
	"github.com/AbhijeetDev102/Nimbus/shared/env"
	pb "github.com/AbhijeetDev102/Nimbus/shared/proto/resource"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ResourceServiceClient struct {
	Client pb.ResourceServiceClient
	conn   *grpc.ClientConn
}

func NewResourceServiceClient() (*ResourceServiceClient, error) {
	resourceServiceUrl := env.GetString("RESOURCE_SERVICE_URL", "localhost:9093")
	if resourceServiceUrl == "" {
		resourceServiceUrl = "resource-service:9093"
	}

	conn, err := grpc.NewClient(resourceServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	ResourceClient := pb.NewResourceServiceClient(conn)

	return &ResourceServiceClient{
		Client: ResourceClient,
		conn:   conn,
	}, nil
}

func (c ResourceServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return
		}
	}
}
