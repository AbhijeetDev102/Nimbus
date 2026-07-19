package grpcclient

import (
	"github.com/AbhijeetDev102/Nimbus/shared/env"
	pb "github.com/AbhijeetDev102/Nimbus/shared/proto/job"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type JobServiceClient struct {
	Client pb.JobServiceClient
	conn   *grpc.ClientConn
}

func NewJobServiceClient() (*JobServiceClient, error) {
	jobServiceUrl := env.GetString("JOB_SERVICE_URL", "localhost:9094")
	if jobServiceUrl == "" {
		jobServiceUrl = "job-service:9094"
	}

	conn, err := grpc.NewClient(jobServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	JobClient := pb.NewJobServiceClient(conn)

	return &JobServiceClient{
		Client: JobClient,
		conn:   conn,
	}, nil
}

func (c JobServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return
		}
	}
}
