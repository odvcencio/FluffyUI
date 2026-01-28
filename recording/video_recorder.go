package recording

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/runtime"
)

// VideoRecorderOptions configures cast capture and video export.
type VideoRecorderOptions struct {
	Cast     AsciicastOptions
	Video    VideoOptions
	CastPath string
	KeepCast bool
}

// VideoRecorder captures a cast file and exports to a video on Close.
type VideoRecorder struct {
	cast       *AsciicastRecorder
	castPath   string
	outputPath string
	options    VideoOptions
	keepCast   bool
}

// NewVideoRecorder creates a recorder that exports to outputPath when closed.
func NewVideoRecorder(outputPath string, options VideoRecorderOptions) (*VideoRecorder, error) {
	if strings.TrimSpace(outputPath) == "" {
		return nil, fmt.Errorf("output path is required")
	}

	castPath := strings.TrimSpace(options.CastPath)
	keepCast := options.KeepCast
	if castPath == "" {
		tmp, err := os.CreateTemp(options.Video.TempDir, "fluffyui-*.cast")
		if err != nil {
			return nil, fmt.Errorf("create temp cast: %w", err)
		}
		castPath = tmp.Name()
		_ = tmp.Close()
	} else {
		keepCast = true
	}

	recorder, err := NewAsciicastRecorder(castPath, options.Cast)
	if err != nil {
		return nil, err
	}

	return &VideoRecorder{
		cast:       recorder,
		castPath:   castPath,
		outputPath: outputPath,
		options:    options.Video,
		keepCast:   keepCast,
	}, nil
}

// Start begins recording.
func (v *VideoRecorder) Start(width, height int, now time.Time) error {
	if v == nil || v.cast == nil {
		return nil
	}
	return v.cast.Start(width, height, now)
}

// Resize updates recording dimensions.
func (v *VideoRecorder) Resize(width, height int) error {
	if v == nil || v.cast == nil {
		return nil
	}
	return v.cast.Resize(width, height)
}

// Frame records a frame.
func (v *VideoRecorder) Frame(buffer *runtime.Buffer, now time.Time) error {
	if v == nil || v.cast == nil {
		return nil
	}
	return v.cast.Frame(buffer, now)
}

// Close finalizes the cast and exports the video.
func (v *VideoRecorder) Close() error {
	if v == nil {
		return nil
	}
	var err error
	if v.cast != nil {
		err = v.cast.Close()
	}
	if exportErr := ExportVideo(v.castPath, v.outputPath, v.options); exportErr != nil {
		err = exportErr
	}
	if !v.keepCast && v.castPath != "" {
		_ = os.Remove(v.castPath)
	}
	return err
}
