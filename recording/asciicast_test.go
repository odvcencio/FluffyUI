package recording

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
)

func TestAsciicastRecorder(t *testing.T) {
	var buf bytes.Buffer
	rec := NewAsciicastRecorderWriter(&buf, AsciicastOptions{Title: "Test"})

	now := time.Unix(0, 0)
	if err := rec.Start(2, 1, now); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	screen := runtime.NewBuffer(2, 1)
	screen.Set(0, 0, 'A', backend.DefaultStyle())
	if err := rec.Frame(screen, now.Add(150*time.Millisecond)); err != nil {
		t.Fatalf("frame failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected header + frame, got %d lines", len(lines))
	}

	var header map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &header); err != nil {
		t.Fatalf("header parse failed: %v", err)
	}
	if header["version"] != float64(2) {
		t.Fatalf("expected version 2, got %v", header["version"])
	}
	if !strings.Contains(lines[1], "A") {
		t.Fatalf("expected frame output to include rune")
	}
}
