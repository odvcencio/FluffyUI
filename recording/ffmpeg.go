package recording

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// TranscodeWithFFmpeg uses ffmpeg to transcode a video file.
func TranscodeWithFFmpeg(inputPath, outputPath string, options FFmpegOptions) error {
	if inputPath == "" || outputPath == "" {
		return fmt.Errorf("input and output paths are required")
	}
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		return fmt.Errorf("ffmpeg not found: %w", err)
	}
	args := []string{"-i", inputPath, "-an"}

	codec := options.VideoCodec
	if codec == "" {
		codec = "libx264"
	}
	args = append(args, "-c:v", codec)

	if options.Preset != "" {
		args = append(args, "-preset", options.Preset)
	}
	if options.CRF > 0 {
		args = append(args, "-crf", strconv.Itoa(options.CRF))
	}
	if options.PixelFormat != "" {
		args = append(args, "-pix_fmt", options.PixelFormat)
	} else {
		args = append(args, "-pix_fmt", "yuv420p")
	}
	if strings.HasSuffix(strings.ToLower(outputPath), ".mp4") {
		args = append(args, "-movflags", "+faststart")
	}
	if len(options.ExtraArgs) > 0 {
		args = append(args, options.ExtraArgs...)
	}
	args = append(args, outputPath)

	cmd := exec.Command(ffmpeg, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w (%s)", err, string(output))
	}
	return nil
}
