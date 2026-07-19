package httphandler

import (
	"encoding/json"
	"log"
	"net/http"

	grpcclient "github.com/AbhijeetDev102/Nimbus/services/api-gateway/internal/grpc_client"
	"github.com/AbhijeetDev102/Nimbus/services/api-gateway/pkg/types"
	jobPb "github.com/AbhijeetDev102/Nimbus/shared/proto/job"
	pb "github.com/AbhijeetDev102/Nimbus/shared/proto/resource"
	sharedTypes "github.com/AbhijeetDev102/Nimbus/shared/types"
)

type httpHandler struct {
	resourceGrpcClient *grpcclient.ResourceServiceClient
	jobGrpcClient      *grpcclient.JobServiceClient
}

func NewHttpHandler(resourceGrpcClient *grpcclient.ResourceServiceClient, jobGrpcClient *grpcclient.JobServiceClient) *httpHandler {
	return &httpHandler{
		resourceGrpcClient: resourceGrpcClient,
		jobGrpcClient:      jobGrpcClient,
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false) // <--- This fixes it!
	return encoder.Encode(data)
}

func (h *httpHandler) HandleUploadUrlRequest(w http.ResponseWriter, r *http.Request) {
	var requestBody sharedTypes.UploadUrlRequest

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "failed to parse json data ", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if requestBody.FileName == "" {
		http.Error(w, "File name is not specified", http.StatusBadRequest)
		return
	}

	if requestBody.FileName == "" {
		http.Error(w, "content type is not specified", http.StatusBadRequest)
		return
	}

	if requestBody.FileName == "" {
		http.Error(w, "file size is not specified", http.StatusBadRequest)
		return
	}
	if requestBody.ResourceType == "" {
		http.Error(w, "file size is not specified", http.StatusBadRequest)
		return
	}

	var response *pb.GeneratePSUrlResponse
	response, err := h.resourceGrpcClient.Client.GeneratePSUrl(r.Context(), &pb.GeneratePSUrlRequest{
		FileName:     requestBody.FileName,
		ContentType:  requestBody.ContentType,
		FileSize:     requestBody.FileSize,
		ResourceType: requestBody.ResourceType,
	})

	if err != nil {
		log.Printf("failed to generate presigned url: %v", err)

		http.Error(
			w,
			"failed to generate upload url",
			http.StatusInternalServerError,
		)

		return
	}

	if err := writeJSON(w, http.StatusCreated, sharedTypes.UploadUrlResponse{
		UploadUrl:  response.GetUploadUrl(),
		ObjectKey:  response.GetObjectKey(),
		ExpiresIn:  response.GetExpiresIn(),
		ResourceID: response.GetResourceID(),
	}); err != nil {
		log.Printf("Failed to writeJson to response in PUrl : %v", err)
	}

}

func (h *httpHandler) HandleCreateJobRequest(w http.ResponseWriter, r *http.Request) {
	var reqBody *types.CreateJobRequest

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "failed to parse json data ", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if reqBody.JobType == "" {
		http.Error(w, "job type is not specified", http.StatusBadRequest)
		return
	}
	if reqBody.ResourceID == "" {
		http.Error(w, "resource id is not specified", http.StatusBadRequest)
		return
	}
	if len(reqBody.Parameters) == 0 || !json.Valid(reqBody.Parameters) {
		http.Error(w, "parameters must be a valid JSON object", http.StatusBadRequest)
		return
	}

	response, err := h.jobGrpcClient.Client.CreateJob(r.Context(), &jobPb.CreateJobRequest{
		ResourceID: reqBody.ResourceID,
		JobType:    reqBody.JobType,
		Parameters: reqBody.Parameters,
	})

	if err != nil {
		log.Printf("failed to create job via gRPC: %v", err)
		http.Error(w, "failed to create job", http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, http.StatusCreated, types.CreateJobResponse{
		JobId:  response.GetJobId(),
		Status: response.GetStatus(),
	}); err != nil {
		log.Printf("Failed to write JSON response: %v", err)
	}

}
