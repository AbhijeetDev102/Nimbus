package grpchandler

import (
	"context"

	"github.com/AbhijeetDev102/Nimbus/services/resource-service/internal/domain"
	pb "github.com/AbhijeetDev102/Nimbus/shared/proto/resource"
	"github.com/AbhijeetDev102/Nimbus/shared/types"
	"github.com/google/uuid"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcHandler struct {
	pb.UnimplementedResourceServiceServer
	service domain.ResourceService
}

func NewGrpcHandler(server *grpc.Server, service domain.ResourceService) *grpcHandler {
	handler := &grpcHandler{
		service: service,
	}

	pb.RegisterResourceServiceServer(server, handler)
	return handler

}

func (h *grpcHandler) GeneratePSUrl(ctx context.Context, req *pb.GeneratePSUrlRequest) (*pb.GeneratePSUrlResponse, error) {
	fileName := req.GetFileName()
	fileSize := req.GetFileSize()
	contentType := req.GetContentType()
	resourceType := req.GetResourceType()

	if fileName == "" {
		return nil, status.Error(codes.InvalidArgument, "File name is not given")
	}
	if fileSize == 0 {
		return nil, status.Error(codes.InvalidArgument, "File size is not given")
	}
	if contentType == "" {
		return nil, status.Error(codes.InvalidArgument, "Content type is not given")
	}
	if resourceType == "" {
		return nil, status.Error(codes.InvalidArgument, "Resource type is not given")
	}

	response, err := h.service.GeneratePSUrl(ctx, &types.UploadUrlRequest{
		ContentType:  contentType,
		FileName:     fileName,
		FileSize:     fileSize,
		ResourceType: resourceType,
	})

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate presigned url : %v", err)
	}
	return &pb.GeneratePSUrlResponse{
		UploadUrl:  response.UploadUrl,
		ObjectKey:  response.ObjectKey,
		ExpiresIn:  response.ExpiresIn,
		ResourceID: response.ResourceID,
	}, nil
}

func (h *grpcHandler) GetDownloadUrl(ctx context.Context, req *pb.GetDownloadUrlRequest) (*pb.GetDownloadUrlResponse, error) {
	resourceIDStr := req.GetResourceID()
	if resourceIDStr == "" {
		return nil, status.Error(codes.InvalidArgument, "resourceID is required")
	}

	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid resourceID format: %v", err)
	}

	expiresIn := req.GetExpiresIn()
	if expiresIn <= 0 {
		expiresIn = 3600
	}

	downloadURL, err := h.service.GetDownloadUrl(ctx, resourceID, expiresIn)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate download url: %v", err)
	}

	return &pb.GetDownloadUrlResponse{
		DownloadUrl: downloadURL,
		ResourceID:  resourceIDStr,
		ExpiresIn:   expiresIn,
	}, nil
}
