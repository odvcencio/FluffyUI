package recording

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/compositor"
	"github.com/odvcencio/fluffyui/runtime"
)

func TestANSIEncoderEncodeEmpty(t *testing.T) {
	encoder := NewANSIEncoder()
	if got := encoder.Encode(nil, false); got != "" {
		t.Fatalf("expected empty output for nil buffer")
	}

	buf := runtime.NewBuffer(2, 1)
	if got := encoder.Encode(buf, false); got != "" {
		t.Fatalf("expected empty output for clean buffer")
	}
}

func TestANSIEncoderEncodeFull(t *testing.T) {
	encoder := NewANSIEncoder()
	buf := runtime.NewBuffer(2, 1)
	buf.Set(0, 0, 'A', backend.DefaultStyle())

	out := encoder.Encode(buf, true)
	if !strings.Contains(out, "A") {
		t.Fatalf("expected output to contain rendered rune")
	}
	if !strings.Contains(out, compositor.ANSIClearScreen) {
		t.Fatalf("expected clear screen sequence")
	}
	if !strings.Contains(out, compositor.ANSICursorHome) {
		t.Fatalf("expected cursor home sequence")
	}
}

func TestExportWithAggErrors(t *testing.T) {
	if err := ExportWithAgg("", "out.webm", AggOptions{}); err == nil {
		t.Fatalf("expected error for empty input path")
	}
	if err := ExportWithAgg("in.cast", "", AggOptions{}); err == nil {
		t.Fatalf("expected error for empty output path")
	}
	t.Setenv("PATH", "")
	if err := ExportWithAgg("in.cast", "out.webm", AggOptions{}); err == nil || !strings.Contains(err.Error(), "agg not found") {
		t.Fatalf("expected agg not found error, got %v", err)
	}
}

func TestTranscodeWithFFmpegErrors(t *testing.T) {
	if err := TranscodeWithFFmpeg("", "out.mp4", FFmpegOptions{}); err == nil {
		t.Fatalf("expected error for empty input path")
	}
	if err := TranscodeWithFFmpeg("in.webm", "", FFmpegOptions{}); err == nil {
		t.Fatalf("expected error for empty output path")
	}
	t.Setenv("PATH", "")
	if err := TranscodeWithFFmpeg("in.webm", "out.mp4", FFmpegOptions{}); err == nil || !strings.Contains(err.Error(), "ffmpeg not found") {
		t.Fatalf("expected ffmpeg not found error, got %v", err)
	}
}

func TestExportVideoTempCleanup(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("PATH", "")

	err := ExportVideo("input.cast", filepath.Join(tmp, "out.mp4"), VideoOptions{TempDir: tmp})
	if err == nil || !strings.Contains(err.Error(), "agg not found") {
		t.Fatalf("expected agg not found error, got %v", err)
	}
	entries, err := os.ReadDir(tmp)
	if err != nil {
		t.Fatalf("read temp dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected temp dir to be empty, got %d entries", len(entries))
	}
}

func TestVideoRecorderKeepsCastWhenSpecified(t *testing.T) {
	tmp := t.TempDir()
	castPath := filepath.Join(tmp, "session.cast")
	outputPath := filepath.Join(tmp, "out.mp4")

	recorder, err := NewVideoRecorder(outputPath, VideoRecorderOptions{
		CastPath: castPath,
		Video:    VideoOptions{TempDir: tmp},
	})
	if err != nil {
		t.Fatalf("new recorder: %v", err)
	}

	buf := runtime.NewBuffer(2, 1)
	buf.Set(0, 0, 'A', backend.DefaultStyle())
	if err := recorder.Start(2, 1, time.Now()); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := recorder.Resize(3, 1); err != nil {
		t.Fatalf("resize: %v", err)
	}
	if err := recorder.Frame(buf, time.Now()); err != nil {
		t.Fatalf("frame: %v", err)
	}

	t.Setenv("PATH", "")
	if err := recorder.Close(); err == nil {
		t.Fatalf("expected error from export with missing agg")
	}
	if _, err := os.Stat(castPath); err != nil {
		t.Fatalf("expected cast file to remain, got %v", err)
	}
}

func TestVideoRecorderRemovesTempCast(t *testing.T) {
	tmp := t.TempDir()
	outputPath := filepath.Join(tmp, "out.webm")
	recorder, err := NewVideoRecorder(outputPath, VideoRecorderOptions{
		Video: VideoOptions{TempDir: tmp},
	})
	if err != nil {
		t.Fatalf("new recorder: %v", err)
	}
	before, _ := filepath.Glob(filepath.Join(tmp, "*.cast"))
	if len(before) == 0 {
		t.Fatalf("expected temp cast file to exist")
	}

	t.Setenv("PATH", "")
	_ = recorder.Close()

	after, _ := filepath.Glob(filepath.Join(tmp, "*.cast"))
	if len(after) != 0 {
		t.Fatalf("expected temp cast file removed, got %d", len(after))
	}
}
