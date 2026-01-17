package recording

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FFmpegOptions configures ffmpeg transcoding.
type FFmpegOptions struct {
	VideoCodec  string
	Preset      string
	CRF         int
	PixelFormat string
	ExtraArgs   []string
}

// VideoOptions configures video export using agg + ffmpeg.
type VideoOptions struct {
	Agg              AggOptions
	FFmpeg           FFmpegOptions
	TempDir          string
	KeepIntermediate bool
}

// ExportVideo renders a cast file into a video format.
// For .mp4 outputs, agg renders a temporary .webm and ffmpeg transcodes to mp4.
func ExportVideo(inputPath, outputPath string, options VideoOptions) error {
	if inputPath == "" || outputPath == "" {
		return fmt.Errorf("input and output paths are required")
	}
	switch strings.ToLower(filepath.Ext(outputPath)) {
	case ".mp4":
		return exportMP4(inputPath, outputPath, options)
	default:
		return ExportWithAgg(inputPath, outputPath, options.Agg)
	}
}

func exportMP4(inputPath, outputPath string, options VideoOptions) error {
	tmpFile, err := os.CreateTemp(options.TempDir, "fluffyui-*.webm")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	if !options.KeepIntermediate {
		defer func() {
			_ = os.Remove(tmpPath)
		}()
	}
	if err := ExportWithAgg(inputPath, tmpPath, options.Agg); err != nil {
		return err
	}
	return TranscodeWithFFmpeg(tmpPath, outputPath, options.FFmpeg)
}
