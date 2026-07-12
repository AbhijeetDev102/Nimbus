package grpcclient

import (
	"github.com/AbhijeetDev102/Nimbus/shared/env"
	pb "github.com/AbhijeetDev102/Nimbus/shared/proto/video"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type VideoServiceClient struct {
	Client pb.VideoServiceClient
	conn   *grpc.ClientConn
}

func NewVideoServiceClient() (*VideoServiceClient, error) {
	videoServiceUrl := env.GetString("VIDEO_SERVICE_URL", "localhost:9093")
	if videoServiceUrl == "" {
		videoServiceUrl = "video-service:9093"
	}

	conn, err := grpc.NewClient(videoServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	VideoClient := pb.NewVideoServiceClient(conn)

	return &VideoServiceClient{
		Client: VideoClient,
		conn:   conn,
	}, nil
}

func (c VideoServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return
		}
	}
}
