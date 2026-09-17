package video

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type FFmpegService struct {
}

func NewFFmpegService() *FFmpegService {
	return &FFmpegService{}
}

type ffprobeOutput struct {
	Streams []struct {
		CodecType string            `json:"codec_type"` // "video", "audio"
		CodecName string            `json:"codec_name"` // e.g. "h264"
		Width     int               `json:"width"`
		Height    int               `json:"height"`
		Duration  string            `json:"duration"`
		Tags      map[string]string `json:"tags"`
	} `json:"streams"`
	Format struct {
		Duration string            `json:"duration"` // ⚠️ ffprobe returns duration as a STRING e.g. "120.5000"
		BitRate  string            `json:"bit_rate"` // e.g. "4500000"
		Tags     map[string]string `json:"tags"`
	} `json:"format"`
}

func parseDurationToUs(timeStr string) int64 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0
	}
	hours, _ := strconv.ParseFloat(parts[0], 64)
	mins, _ := strconv.ParseFloat(parts[1], 64)
	secs, _ := strconv.ParseFloat(parts[2], 64)
	totalSecs := hours*3600 + mins*60 + secs
	return int64(totalSecs * 1000000)
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
		return nil, fmt.Errorf("failed to unmarshal the ffprobe output: %v", err)
	}

	for _, stream := range probeOutput.Streams {
		if stream.CodecType == "video" {
			duration, _ := strconv.ParseFloat(probeOutput.Format.Duration, 64)
			// Fallback 1: Stream-level duration (common in 4K MP4 / MOV)
			if duration <= 0 && stream.Duration != "" {
				duration, _ = strconv.ParseFloat(stream.Duration, 64)
			}
			// Fallback 2: Format tags (e.g. DURATION=00:01:23.456)
			if duration <= 0 && probeOutput.Format.Tags != nil {
				if dStr, ok := probeOutput.Format.Tags["DURATION"]; ok {
					duration = float64(parseDurationToUs(dStr)) / 1000000.0
				}
			}
			// Fallback 3: Stream tags
			if duration <= 0 && stream.Tags != nil {
				if dStr, ok := stream.Tags["DURATION"]; ok {
					duration = float64(parseDurationToUs(dStr)) / 1000000.0
				}
			}

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

type ProgressCallback func(percent float64, speed string, fps float64)

func (f *FFmpegService) Transcode(ctx context.Context, inputPath string, outputPath string, parameters VideoParameters, totalDuration float64, onProgress ProgressCallback) error {
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
		"-progress", "pipe:1", // <-- Instructs FFmpeg to stream statistics to stdout
		"-nostats", // <-- Suppresses noisy console animation
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
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get ffmpeg stdout pipe: %w", err)
	}
	// Start FFmpeg process asynchronously
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}
	// 4. Read FFmpeg progress stream line-by-line in real time
	scanner := bufio.NewScanner(stdoutPipe)
	var outTimeUs int64
	var speed string
	var fps float64
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			switch key {
			case "out_time_us":
				if val != "N/A" {
					if parsed, err := strconv.ParseInt(val, 10, 64); err == nil && parsed > 0 {
						outTimeUs = parsed
					}
				}
			case "out_time_ms":
				if outTimeUs == 0 && val != "N/A" {
					if parsed, err := strconv.ParseInt(val, 10, 64); err == nil && parsed > 0 {
						outTimeUs = parsed
					}
				}
			case "out_time":
				if outTimeUs == 0 && strings.Contains(val, ":") {
					outTimeUs = parseDurationToUs(val)
				}
			case "speed":
				speed = val
			case "fps":
				fps, _ = strconv.ParseFloat(val, 64)
			case "progress":
				if outTimeUs > 0 && onProgress != nil {
					outSeconds := float64(outTimeUs) / 1000000.0
					var percent float64
					if totalDuration > 0 {
						percent = (outSeconds / totalDuration) * 100.0
						if percent > 99.0 {
							percent = 99.0 // Cap at 99% until process finishes cleanly
						}
					}
					onProgress(percent, speed, fps)
				}
			}
		}
	}
	if err := scanner.Err(); err != nil {

		_ = cmd.Wait()
		return fmt.Errorf("failed reading ffmpeg progress output: %w", err)
	}
	// Wait for FFmpeg process to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg transcoding failed: %w", err)
	}
	// Emit final 100% completion
	if onProgress != nil {
		onProgress(100.0, speed, fps)
	}
	return nil
}
