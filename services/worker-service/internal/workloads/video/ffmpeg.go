package video

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
)

type FFmpegService struct {
}

func NewFFmpegService() *FFmpegService {
	return &FFmpegService{}
}

type ffprobeOutput struct {
	Streams []struct {
		CodecType string `json:"codec_type"` // "video", "audio"
		CodecName string `json:"codec_name"` // e.g. "h264"
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"` // ⚠️ ffprobe returns duration as a STRING e.g. "120.5000"
		BitRate  string `json:"bit_rate"` // e.g. "4500000"
	} `json:"format"`
}

func (f *FFmpegService) Probe(ctx context.Context, inputPath string) (*VideoMetaData, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		inputPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed on %s: %w", inputPath, err)
	}

	var probeOutput ffprobeOutput

	if err := json.Unmarshal(output, &probeOutput); err != nil {
		return nil, fmt.Errorf("failed to unmarsher the ffprobeoutput : %v", err)
	}

	for _, stream := range probeOutput.Streams {
		if stream.CodecType == "video" {
			duration, _ := strconv.ParseFloat(probeOutput.Format.Duration, 64)

			bitrate, _ := strconv.ParseInt(probeOutput.Format.BitRate, 10, 64)

			return &VideoMetaData{
				DurationSeconds: duration,
				Width:           stream.Width,
				Height:          stream.Height,
				CodecName:       stream.CodecName,
				Bitrate:         bitrate,
			}, nil
		}
	}
	return nil, fmt.Errorf("no video stream found in %s", inputPath)
}

func (f *FFmpegService) Transcode(ctx context.Context, inputPath string, outputPath string, parameters VideoParameters) error {
	// 1. Determine the scaling filter based on requested resolution
	scaleFilter := "scale=-2:720" // default fallback
	switch parameters.Resolution {
	case "1080p":
		scaleFilter = "scale=-2:1080"
	case "720p":
		scaleFilter = "scale=-2:720"
	case "480p":
		scaleFilter = "scale=-2:480"
	case "360p":
		scaleFilter = "scale=-2:360"
	}

	// 2. Build the argument list
	args := []string{
		"-y",
		"-i", inputPath,
		"-vf", scaleFilter,
		"-c:v", "libx264",
		"-preset", "fast",
		"-c:a", "aac",
	}

	// If a custom bitrate was specified (e.g. "4000k"), apply it
	if parameters.Bitrate != "" {
		args = append(args, "-b:v", parameters.Bitrate)
	}

	// Append output destination path
	args = append(args, outputPath)

	// 3. Execute with context (so cancellation immediately terminates FFmpeg)
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	// CombinedOutput captures both stdout & stderr (FFmpeg outputs all logs to stderr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg transcoding failed: %w (output: %s)", err, string(output))
	}

	return nil
}
