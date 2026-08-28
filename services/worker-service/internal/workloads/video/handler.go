package video

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AbhijeetDev102/Nimbus/services/worker-service/internal/domain"
	"github.com/google/uuid"

	"gorm.io/datatypes"
)

type VideoHandler struct {
	storage *MinioInstance
	ffmpeg  *FFmpegService
}

func NewVideoHandler(storage *MinioInstance, ffmpeg *FFmpegService) *VideoHandler {
	return &VideoHandler{
		storage: storage,
		ffmpeg:  ffmpeg,
	}
}
func (h *VideoHandler) Execute(ctx context.Context, job *domain.Job) (*domain.ExecutionResult, error) {
	tempDir, err := os.MkdirTemp("", "videoFile-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create the temp downlaod folder :%v", err)
	}

	defer os.RemoveAll(tempDir)

	fmt.Printf("Temporary directory created at: %s\n", tempDir)

	var videoParams VideoParameters

	if err := json.Unmarshal(job.Parameters, &videoParams); err != nil {
		return nil, fmt.Errorf("failed to parse job parameter to videoParameter : %v", err)
	}

	localInputPath := filepath.Join(tempDir, "input.mp4")
	localOutputPath := filepath.Join(tempDir, "output.mp4")

	inputKey := fmt.Sprintf("resource/original/%s.mp4", job.ResourceID.String())
	if err := h.storage.Download(ctx, inputKey, localInputPath); err != nil {
		return nil, fmt.Errorf("failed to download source video: %w", err)
	}

	metadata, err := h.ffmpeg.Probe(ctx, localInputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to probe video metadata: %w", err)
	}

	if err := h.ffmpeg.Transcode(ctx, localInputPath, localOutputPath, videoParams); err != nil {
		return nil, fmt.Errorf("transcoding failed: %w", err)
	}

	outputResourceID := uuid.New()
	outputKey := fmt.Sprintf("resource/processed/%s.mp4", outputResourceID.String())

	if err := h.storage.Upload(ctx, outputKey, localOutputPath, "video/mp4"); err != nil {
		return nil, fmt.Errorf("failed to upload transcoded video: %w", err)
	}

	metadataJSON, _ := json.Marshal(metadata)

	return &domain.ExecutionResult{
		OutputResourceID: &outputResourceID,
		Metadata:         datatypes.JSON(metadataJSON),
		Error:            nil,
	}, nil

}
