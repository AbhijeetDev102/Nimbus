package httphandler

import (
	"encoding/json"
	"log"
	"net/http"

	grpcclient "github.com/AbhijeetDev102/Nimbus/services/api-gateway/internal/grpc_client"
	"github.com/AbhijeetDev102/Nimbus/services/api-gateway/pkg/types"
	pb "github.com/AbhijeetDev102/Nimbus/shared/proto/video"
)

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func HandleUploadUrlRequest(w http.ResponseWriter, r *http.Request) {
	var requestBody types.UploadUrlRequest

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

	grpcClient, err := grpcclient.NewVideoServiceClient()
	if err != nil {
		log.Printf("Failed to create video service Grpc client : %v", err)
		return
	}

	defer grpcClient.Close()
	var response *pb.GeneratePSUrlResponse
	response, err = grpcClient.Client.GeneratePSUrl(r.Context(), &pb.GeneratePSUrlRequest{
		FileName:    requestBody.FileName,
		ContentType: requestBody.ContentType,
		FileSize:    requestBody.FileSize,
	})

	if err != nil {
		log.Printf("failed to generate Presigned Url : %v ", err)
	}

	if err := writeJSON(w, http.StatusCreated, types.UploadUrlResponse{
		UploadUrl: response.GetUploadUrl(),
		ObjectKey: response.GetObjectKey(),
		ExpiresIn: response.GetExpiresIn(),
	}); err != nil {
		log.Printf("Failed to writeJson to response in PUrl : %v", err)
	}

}

func HandleCreateJob(w http.ResponseWriter, r *http.Request) {

}
