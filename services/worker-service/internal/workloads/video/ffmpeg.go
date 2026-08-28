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
