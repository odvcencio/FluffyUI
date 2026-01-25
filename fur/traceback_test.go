package fur

import (
	"errors"
	"strings"
	"testing"
)

func TestTracebackBasic(t *testing.T) {
	err := errors.New("something went wrong")
	r := Traceback(err)
	lines := r.Render(80)

	text := extractText(lines)
	if !strings.Contains(text, "something went wrong") {
		t.Error("expected error message in output")
	}
	if !strings.Contains(text, "Traceback") {
		t.Error("expected Traceback header")
	}
}

func TestTracebackNil(t *testing.T) {
	r := Traceback(nil)
	lines := r.Render(80)

	if len(lines) != 0 {
		t.Error("expected no output for nil error")
	}
}

func TestTracebackWithOpts(t *testing.T) {
	err := errors.New("test error")
	r := TracebackWith(err, TracebackOpts{
		Context:   2,
		MaxFrames: 5,
	})
	lines := r.Render(80)

	text := extractText(lines)
	if !strings.Contains(text, "test error") {
		t.Error("expected error message")
	}
}

func TestWrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original)

	if wrapped == nil {
		t.Fatal("Wrap returned nil")
	}
	if !strings.Contains(wrapped.Error(), "original error") {
		t.Error("wrapped error should contain original message")
	}
}

func TestWrapNil(t *testing.T) {
	wrapped := Wrap(nil)
	if wrapped != nil {
		t.Error("Wrap(nil) should return nil")
	}
}

func TestWrapMsg(t *testing.T) {
	original := errors.New("connection refused")
	wrapped := WrapMsg(original, "failed to connect")

	if wrapped == nil {
		t.Fatal("WrapMsg returned nil")
	}
	msg := wrapped.Error()
	if !strings.Contains(msg, "failed to connect") {
		t.Error("expected wrapper message")
	}
	if !strings.Contains(msg, "connection refused") {
		t.Error("expected original message")
	}
}

func TestWrapMsgNil(t *testing.T) {
	wrapped := WrapMsg(nil, "message")
	if wrapped != nil {
		t.Error("WrapMsg(nil, ...) should return nil")
	}
}

func TestTraceErrorUnwrap(t *testing.T) {
	original := errors.New("root cause")
	wrapped := Wrap(original)

	unwrapped := errors.Unwrap(wrapped)
	if unwrapped != original {
		t.Error("Unwrap should return original error")
	}
}

func TestWrapIdempotent(t *testing.T) {
	original := errors.New("error")
	wrapped1 := Wrap(original)
	wrapped2 := Wrap(wrapped1)

	// Second wrap should return same error since it already has stack
	if wrapped2 != wrapped1 {
		t.Error("Wrap should be idempotent for errors with stack")
	}
}
