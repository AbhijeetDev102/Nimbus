package video

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/AbhijeetDev102/Nimbus/pkg/nimbus"
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
func (h *VideoHandler) Execute(ctx nimbus.Context, job *nimbus.Job) (*nimbus.ExecutionResult, error) {

	if job.ResourceID == nil {
		return nil, fmt.Errorf("video transcode workload requires a valid resource_id")
	}

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
	log.Printf("[Job %s] Downloading source video: %s", job.ID, inputKey)
	if err := h.storage.Download(ctx, inputKey, localInputPath); err != nil {
		return nil, fmt.Errorf("failed to download source video: %w", err)
	}

	log.Printf("[Job %s] Probing source video metadata...", job.ID)

	metadata, err := h.ffmpeg.Probe(ctx, localInputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to probe video metadata: %w", err)
	}

	//Previous implementation without Nimbus SDK
	// Throttle Redis progress events to at most once every 250ms
	// var lastPublished time.Time
	// onProgress := func(percent float64, speed string, fps float64) {
	// 	if time.Since(lastPublished) >= 250*time.Millisecond || percent >= 100.0 {
	// 		lastPublished = time.Now()
	// 		if h.publisher != nil {
	// 			_ = h.publisher.Publish(ctx, &domain.ProgressUpdate{
	// 				JobID:    job.ID.String(),
	// 				Progress: percent,
	// 				Speed:    speed,
	// 				FPS:      fps,
	// 				Status:   "RUNNING",
	// 			})
	// 		}
	// 	}
	// }

	//New implementation with Nimbus SDK
	onProgress := func(percent float64, speed string, fps float64) {
		ctx.ReportProgressDetails(percent, "Transcoding video", map[string]any{"speed": speed, "fps": fps})
	}

	log.Printf("[Job %s] Transcoding video to %s (%s)...", job.ID, videoParams.Resolution, videoParams.Codec)
	if err := h.ffmpeg.Transcode(ctx, localInputPath, localOutputPath, videoParams, metadata.DurationSeconds, onProgress); err != nil {
		return nil, fmt.Errorf("transcoding failed: %w", err)
	}

	// Probe the newly created output video for true output metadata
	outputMetadata, err := h.ffmpeg.Probe(ctx, localOutputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to probe output video metadata: %w", err)
	}

	outputResourceID := uuid.New()
	outputKey := fmt.Sprintf("resource/processed/%s.mp4", outputResourceID.String())

	log.Printf("[Job %s] Uploading processed video to MinIO: %s", job.ID, outputKey)
	if err := h.storage.Upload(ctx, outputKey, localOutputPath, "video/mp4"); err != nil {
		return nil, fmt.Errorf("failed to upload transcoded video: %w", err)
	}

	log.Printf("[Job %s] Video pipeline completed successfully!", job.ID)
	metadataJSON, _ := json.Marshal(outputMetadata)

	return &nimbus.ExecutionResult{
		OutputResourceID: &outputResourceID,
		Metadata:         datatypes.JSON(metadataJSON),
	}, nil

}
