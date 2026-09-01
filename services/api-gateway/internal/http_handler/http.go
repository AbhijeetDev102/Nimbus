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
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type httpHandler struct {
	resourceGrpcClient *grpcclient.ResourceServiceClient
	jobGrpcClient      *grpcclient.JobServiceClient
	redisClient        *redis.Client
}

func NewHttpHandler(resourceGrpcClient *grpcclient.ResourceServiceClient, jobGrpcClient *grpcclient.JobServiceClient, redisClient *redis.Client) *httpHandler {
	return &httpHandler{
		resourceGrpcClient: resourceGrpcClient,
		jobGrpcClient:      jobGrpcClient,
		redisClient:        redisClient,
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

	if requestBody.ContentType == "" {
		http.Error(w, "content type is not specified", http.StatusBadRequest)
		return
	}

	if requestBody.FileSize == 0 {
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
	var resourceID *string
	if reqBody.ResourceID != "" {
		resourceID = &reqBody.ResourceID
	}
	if len(reqBody.Parameters) == 0 || !json.Valid(reqBody.Parameters) {
		http.Error(w, "parameters must be a valid JSON object", http.StatusBadRequest)
		return
	}

	response, err := h.jobGrpcClient.Client.CreateJob(r.Context(), &jobPb.CreateJobRequest{
		ResourceID: resourceID,
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

func (h *httpHandler) HandleGetJob(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")
	if jobID == "" {
		http.Error(w, "job ID is required", http.StatusBadRequest)
		return
	}

	response, err := h.jobGrpcClient.Client.GetJob(r.Context(), &jobPb.GetJobRequest{
		JobId: jobID,
	})
	if err != nil {
		log.Printf("failed to get job via gRPC: %v", err)
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow cross-origin in development
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (h *httpHandler) HandleJobProgressWS(w http.ResponseWriter, r *http.Request) {
	// extract job ID from url path
	jobID := r.PathValue("id")
	if jobID == "" {
		http.Error(w, "job ID is required", http.StatusBadRequest)
		return
	}

	//Upgrade HTTP Connection to a WebSocket TCP Connection

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade websocket for job %s: %v", jobID, err)
		return
	}
	defer conn.Close()

	log.Printf("[WebSocket] Client connected for job %s", jobID)

	// subscribe to the specifc redis pub/sub channel for this job

	channelName := "job:progress:" + jobID
	pubsub := h.redisClient.Subscribe(r.Context(), channelName)

	defer pubsub.Close()

	ch := pubsub.Channel()

	//stream redis messgae directly to the websocket client
	for {
		select {
		case <-r.Context().Done():
			log.Printf("[WebSocket] Connection closed for job %s", jobID)
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			// Write the JSON payload down the WebSocket wire
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				log.Printf("[WebSocket] Client disconnected for job %s: %v", jobID, err)
				return // Client closed browser tab
			}
		}
	}
}
