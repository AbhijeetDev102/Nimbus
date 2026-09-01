package grpc_handler

import (
	"context"
	"encoding/json"
	"time"

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

	var parserID *uuid.UUID
	if resourceID != "" {

		value, err := uuid.Parse(resourceID)

		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid resource ID format: %v", err)

		}
		parserID = &value
	}
	if jobType == "" {
		return nil, status.Errorf(codes.InvalidArgument, "jobType is not given")
	}

	if !json.Valid(parameters) {
		return nil, status.Errorf(codes.InvalidArgument, "parameters must be valid JSON")
	}

	job, err := h.service.CreateJob(ctx, &domain.CreateJobRequest{
		ResourceID: parserID,
		JobType:    types.JobType(jobType),
		Parameters: datatypes.JSON(parameters),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create job: %v", err)
	}

	return &pb.CreateJobResponse{
		JobId:  job.ID.String(),
		Status: string(job.Status),
	}, nil

}

func (h *grpcHandler) GetJob(ctx context.Context, req *pb.GetJobRequest) (*pb.GetJobResponse, error) {
	jobId, err := uuid.Parse(req.JobId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid job ID: %v", err)
	}

	job, err := h.service.GetJob(ctx, jobId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "job not found: %v", err)
	}

	var errMsg *string
	if job.ErrorMessage != nil {
		errMsg = job.ErrorMessage
	}

	var ResourceID *string
	if job.ResourceID != nil {
		s := job.ResourceID.String()
		ResourceID = &s
	}

	var outputResourceID *string
	if job.OutputResourceID != nil {
		s := job.OutputResourceID.String()
		outputResourceID = &s
	}
	var startedAt *string
	if job.StartedAt != nil {
		s := job.StartedAt.Format(time.RFC3339)
		startedAt = &s
	}
	var completedAt *string
	if job.CompletedAt != nil {
		s := job.CompletedAt.Format(time.RFC3339)
		completedAt = &s
	}

	return &pb.GetJobResponse{
		JobId:            job.ID.String(),
		ResourceID:       ResourceID,
		JobType:          string(job.JobType),
		Status:           string(job.Status),
		RetryCount:       int32(job.RetryCount),
		MaxRetries:       int32(job.MaxRetries),
		ErrorMessage:     errMsg,
		OutputResourceID: outputResourceID,
		Parameters:       []byte(job.Parameters),
		Metadata:         []byte(job.Metadata),
		CreatedAt:        job.CreatedAt.Format(time.RFC3339),
		StartedAt:        startedAt,
		CompletedAt:      completedAt,
	}, nil
}

func mapJobToProto(job *domain.Job) *pb.GetJobResponse {
	var errMsg *string
	if job.ErrorMessage != nil {
		errMsg = job.ErrorMessage
	}

	var resourceID *string
	if job.ResourceID != nil {
		s := job.ResourceID.String()
		resourceID = &s
	}

	var outputResourceID *string
	if job.OutputResourceID != nil {
		s := job.OutputResourceID.String()
		outputResourceID = &s
	}

	var startedAt *string
	if job.StartedAt != nil {
		s := job.StartedAt.Format(time.RFC3339)
		startedAt = &s
	}

	var completedAt *string
	if job.CompletedAt != nil {
		s := job.CompletedAt.Format(time.RFC3339)
		completedAt = &s
	}

	return &pb.GetJobResponse{
		JobId:            job.ID.String(),
		ResourceID:       resourceID,
		JobType:          string(job.JobType),
		Status:           string(job.Status),
		RetryCount:       int32(job.RetryCount),
		MaxRetries:       int32(job.MaxRetries),
		ErrorMessage:     errMsg,
		OutputResourceID: outputResourceID,
		Parameters:       []byte(job.Parameters),
		Metadata:         []byte(job.Metadata),
		CreatedAt:        job.CreatedAt.Format(time.RFC3339),
		StartedAt:        startedAt,
		CompletedAt:      completedAt,
	}
}

func (h *grpcHandler) ListJobs(ctx context.Context, req *pb.ListJobsRequest) (*pb.ListJobsResponse, error) {
	var statusFilter *types.JobStatus
	if req.Status != nil && *req.Status != "" {
		st := types.JobStatus(*req.Status)
		statusFilter = &st
	}

	var jobTypeFilter *types.JobType
	if req.JobType != nil && *req.JobType != "" {
		jt := types.JobType(*req.JobType)
		jobTypeFilter = &jt
	}

	jobs, totalCount, err := h.service.ListJobs(ctx, &domain.ListJobsRequest{
		Limit:   int(req.GetLimit()),
		Offset:  int(req.GetOffset()),
		Status:  statusFilter,
		JobType: jobTypeFilter,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list jobs: %v", err)
	}

	protoJobs := make([]*pb.GetJobResponse, len(jobs))
	for i, j := range jobs {
		protoJobs[i] = mapJobToProto(j)
	}

	return &pb.ListJobsResponse{
		Jobs:       protoJobs,
		TotalCount: totalCount,
	}, nil
}

func (h *grpcHandler) GetJobStats(ctx context.Context, req *pb.GetJobStatsRequest) (*pb.GetJobStatsResponse, error) {
	stats, err := h.service.GetJobStats(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get job stats: %v", err)
	}

	return &pb.GetJobStatsResponse{
		Total:     stats.Total,
		Queued:    stats.Queued,
		Running:   stats.Running,
		Completed: stats.Completed,
		Failed:    stats.Failed,
	}, nil
}
