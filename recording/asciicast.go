package recording

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/odvcencio/fluffy-ui/runtime"
)

// AsciicastOptions configures asciicast recording.
type AsciicastOptions struct {
	Title string
	Env   map[string]string
}

// AsciicastRecorder writes asciicast v2 recordings.
type AsciicastRecorder struct {
	mu       sync.Mutex
	writer   io.Writer
	closers  []io.Closer
	started  bool
	start    time.Time
	width    int
	height   int
	fullNext bool
	options  AsciicastOptions
	encoder  *ANSIEncoder
}

// NewAsciicastRecorder creates a recorder writing to path.
func NewAsciicastRecorder(path string, options AsciicastOptions) (*AsciicastRecorder, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return nil, err
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	var writer io.Writer = file
	closers := []io.Closer{file}
	if strings.HasSuffix(path, ".gz") {
		gz := gzip.NewWriter(file)
		writer = gz
		closers = append([]io.Closer{gz}, closers...)
	}
	recorder := NewAsciicastRecorderWriter(writer, options)
	recorder.closers = closers
	return recorder, nil
}

// NewAsciicastRecorderWriter creates a recorder writing to writer.
func NewAsciicastRecorderWriter(writer io.Writer, options AsciicastOptions) *AsciicastRecorder {
	return &AsciicastRecorder{
		writer:  writer,
		options: options,
		encoder: NewANSIEncoder(),
	}
}

// Start writes the asciicast header.
func (a *AsciicastRecorder) Start(width, height int, now time.Time) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.started {
		return nil
	}
	a.started = true
	a.start = now
	a.width = width
	a.height = height
	a.fullNext = true
	return a.writeHeaderLocked()
}

// Resize updates the recording dimensions.
func (a *AsciicastRecorder) Resize(width, height int) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.width = width
	a.height = height
	a.fullNext = true
	return nil
}

// Frame records a frame from the buffer.
func (a *AsciicastRecorder) Frame(buffer *runtime.Buffer, now time.Time) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.started {
		a.started = true
		a.start = now
		if buffer != nil {
			a.width, a.height = buffer.Size()
		}
		a.fullNext = true
		if err := a.writeHeaderLocked(); err != nil {
			return err
		}
	}
	if buffer == nil || a.encoder == nil {
		return nil
	}
	full := a.fullNext
	a.fullNext = false
	frame := a.encoder.Encode(buffer, full)
	if frame == "" {
		return nil
	}
	delta := now.Sub(a.start).Seconds()
	payload := []any{delta, "o", frame}
	return writeJSONLine(a.writer, payload)
}

// Close closes the recorder writer.
func (a *AsciicastRecorder) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	var err error
	for _, closer := range a.closers {
		if closeErr := closer.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	a.closers = nil
	return err
}

func (a *AsciicastRecorder) writeHeaderLocked() error {
	header := map[string]any{
		"version":   2,
		"width":     a.width,
		"height":    a.height,
		"timestamp": a.start.Unix(),
	}
	if a.options.Title != "" {
		header["title"] = a.options.Title
	}
	env := a.options.Env
	if env == nil {
		env = map[string]string{}
		if term := os.Getenv("TERM"); term != "" {
			env["TERM"] = term
		}
		if shell := os.Getenv("SHELL"); shell != "" {
			env["SHELL"] = shell
		}
	}
	if len(env) > 0 {
		header["env"] = env
	}
	return writeJSONLine(a.writer, header)
}

func writeJSONLine(writer io.Writer, value any) error {
	if writer == nil {
		return nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = writer.Write(append(payload, '\n'))
	return err
}
