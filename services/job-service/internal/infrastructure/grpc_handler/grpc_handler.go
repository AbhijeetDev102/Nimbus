package grpc_handler

import (
	"context"
	"encoding/json"

	"github.com/AbhijeetDev102/Nimbus/services/job-service/internal/domain"
	pb "github.com/AbhijeetDev102/Nimbus/shared/proto/job"
	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/datatypes"
)

type grpcHandler struct {
	pb.UnimplementedJobServiceServer
	service domain.JobService
}

func NewGrpcHandler(server *grpc.Server, service domain.JobService) *grpcHandler {
	handler := &grpcHandler{
		service: service,
	}

	pb.RegisterJobServiceServer(server, handler)
	return handler
}

func (h *grpcHandler) CreateJob(ctx context.Context, req *pb.CreateJobRequest) (*pb.CreateJobResponse, error) {
	resourceID := req.GetResourceID()
	jobType := req.GetJobType()
	parameters := req.GetParameters()

	if resourceID == "" {
		return nil, status.Errorf(codes.InvalidArgument, "resourceID is not given")
	}
	if jobType == "" {
		return nil, status.Errorf(codes.InvalidArgument, "resourceID is not given")
	}

	if !json.Valid(parameters) {
		return nil, status.Errorf(codes.InvalidArgument, "parameters must be valid JSON")
	}

	parserID, err := uuid.Parse(resourceID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resource ID format: %v", err)

	}
	job, err := h.service.CreateJob(ctx, &domain.CreateJobRequest{
		ResourceID: parserID,
		JobType:    types.JobType(jobType),
		Parameters: datatypes.JSON(parameters),
	})

	return &pb.CreateJobResponse{
		JobId:  job.ID.String(),
		Status: string(job.Status),
	}, nil

}
