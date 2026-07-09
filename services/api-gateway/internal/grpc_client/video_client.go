package grpcclient

import (
	"os"

	pb "github.com/AbhijeetDev102/Nimbus/shared/proto/video"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type videoServiceClient struct {
	Client pb.VideoServiceClient
	conn   *grpc.ClientConn
}

func NewVideoServiceClient() (*videoServiceClient, error) {
	videoServiceUrl := os.Getenv("VIDEO_SERVICE_URL")
	if videoServiceUrl == "" {
		videoServiceUrl = "video-service:9093"
	}

	conn, err := grpc.NewClient(videoServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return nil, err
	}

	VideoClient := pb.NewVideoServiceClient(conn)

	return &videoServiceClient{
		Client: VideoClient,
		conn:   conn,
	}, nil
}

func (c videoServiceClient) Close() {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return
		}
	}
}
