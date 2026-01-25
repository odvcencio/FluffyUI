package fur

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestHandlerBasic(t *testing.T) {
	var buf bytes.Buffer
	console := New(WithOutput(&buf), WithNoColor(), WithWidth(80))

	handler := &Handler{
		opts:    HandlerOpts{Level: slog.LevelInfo},
		console: console,
	}

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	err := handler.Handle(context.Background(), record)

	if err != nil {
		t.Fatalf("Handle returned error: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("expected message in output, got %q", output)
	}
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected INFO level in output, got %q", output)
	}
}

func TestHandlerWithTime(t *testing.T) {
	var buf bytes.Buffer
	console := New(WithOutput(&buf), WithNoColor(), WithWidth(80))

	handler := &Handler{
		opts:    HandlerOpts{Level: slog.LevelInfo, ShowTime: true, TimeFormat: "15:04:05"},
		console: console,
	}

	record := slog.NewRecord(time.Date(2026, 1, 24, 14, 30, 45, 0, time.UTC), slog.LevelInfo, "timed", 0)
	_ = handler.Handle(context.Background(), record)

	output := buf.String()
	if !strings.Contains(output, "14:30:45") {
		t.Errorf("expected timestamp in output, got %q", output)
	}
}

func TestHandlerEnabled(t *testing.T) {
	handler := NewHandler(HandlerOpts{Level: slog.LevelWarn})

	if handler.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("Debug should not be enabled when level is Warn")
	}
	if handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("Info should not be enabled when level is Warn")
	}
	if !handler.Enabled(context.Background(), slog.LevelWarn) {
		t.Error("Warn should be enabled when level is Warn")
	}
	if !handler.Enabled(context.Background(), slog.LevelError) {
		t.Error("Error should be enabled when level is Warn")
	}
}

func TestHandlerWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	console := New(WithOutput(&buf), WithNoColor(), WithWidth(80))

	handler := &Handler{
		opts:    HandlerOpts{Level: slog.LevelInfo},
		console: console,
	}

	withAttrs := handler.WithAttrs([]slog.Attr{
		slog.String("key", "value"),
	})

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "with attrs", 0)
	_ = withAttrs.Handle(context.Background(), record)

	output := buf.String()
	if !strings.Contains(output, "key=") {
		t.Errorf("expected attr key in output, got %q", output)
	}
}

func TestHandlerWithGroup(t *testing.T) {
	var buf bytes.Buffer
	console := New(WithOutput(&buf), WithNoColor(), WithWidth(80))

	handler := &Handler{
		opts:    HandlerOpts{Level: slog.LevelInfo},
		console: console,
	}

	withGroup := handler.WithGroup("mygroup")
	withAttrs := withGroup.WithAttrs([]slog.Attr{
		slog.String("field", "data"),
	})

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "grouped", 0)
	_ = withAttrs.Handle(context.Background(), record)

	output := buf.String()
	if !strings.Contains(output, "mygroup.field=") {
		t.Errorf("expected grouped attr in output, got %q", output)
	}
}

func TestHandlerPrettyAttrs(t *testing.T) {
	var buf bytes.Buffer
	console := New(WithOutput(&buf), WithNoColor(), WithWidth(80))

	handler := &Handler{
		opts:    HandlerOpts{Level: slog.LevelInfo, Pretty: true},
		console: console,
	}

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "pretty", 0)
	record.AddAttrs(slog.String("attr1", "val1"), slog.Int("attr2", 42))
	_ = handler.Handle(context.Background(), record)

	output := buf.String()
	if !strings.Contains(output, "attr1=") {
		t.Errorf("expected attr1 in output, got %q", output)
	}
}

func TestHandlerNil(t *testing.T) {
	var h *Handler
	err := h.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0))
	if err != nil {
		t.Errorf("nil handler should return nil error, got %v", err)
	}
}

func TestNewHandler(t *testing.T) {
	handler := NewHandler(HandlerOpts{
		ShowTime:   true,
		ShowSource: true,
		Pretty:     true,
	})

	if handler == nil {
		t.Fatal("NewHandler returned nil")
	}
	if handler.opts.TimeFormat != "15:04:05" {
		t.Errorf("expected default time format, got %q", handler.opts.TimeFormat)
	}
}
