package clipboard

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"m31labs.dev/fluffyui/terminal"
)

func TestOSC52Write(t *testing.T) {
	var buf bytes.Buffer
	caps := &terminal.TerminalCaps{OSC52Clipboard: true}
	c := NewOSC52Clipboard(&buf, caps)

	if err := c.Write("hello world"); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	got := buf.String()

	// Verify the escape sequence structure.
	if !strings.HasPrefix(got, "\033]52;c;") {
		t.Fatalf("expected OSC-52 prefix, got %q", got)
	}
	if !strings.HasSuffix(got, "\a") {
		t.Fatalf("expected BEL terminator, got %q", got)
	}

	// Extract and decode the base64 payload.
	payload := got[len("\033]52;c;") : len(got)-1]
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		t.Fatalf("base64 decode failed: %v", err)
	}
	if string(decoded) != "hello world" {
		t.Fatalf("decoded = %q, want %q", string(decoded), "hello world")
	}
}

func TestOSC52WriteEmpty(t *testing.T) {
	var buf bytes.Buffer
	caps := &terminal.TerminalCaps{OSC52Clipboard: true}
	c := NewOSC52Clipboard(&buf, caps)

	if err := c.Write(""); err != nil {
		t.Fatalf("Write empty failed: %v", err)
	}

	got := buf.String()

	// Empty string should produce a valid OSC-52 sequence with empty base64.
	encoded := base64.StdEncoding.EncodeToString([]byte(""))
	expected := "\033]52;c;" + encoded + "\a"
	if got != expected {
		t.Fatalf("got %q, want %q", got, expected)
	}
}

func TestOSC52WriteUnicode(t *testing.T) {
	var buf bytes.Buffer
	caps := &terminal.TerminalCaps{OSC52Clipboard: true}
	c := NewOSC52Clipboard(&buf, caps)

	text := "Hello, \u4e16\u754c! \U0001F4CB"
	if err := c.Write(text); err != nil {
		t.Fatalf("Write unicode failed: %v", err)
	}

	got := buf.String()
	payload := got[len("\033]52;c;") : len(got)-1]
	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		t.Fatalf("base64 decode failed: %v", err)
	}
	if string(decoded) != text {
		t.Fatalf("decoded = %q, want %q", string(decoded), text)
	}
}

func TestOSC52ReadReturnsError(t *testing.T) {
	var buf bytes.Buffer
	caps := &terminal.TerminalCaps{OSC52Clipboard: true}
	c := NewOSC52Clipboard(&buf, caps)

	_, err := c.Read()
	if err == nil {
		t.Fatal("expected Read to return error for OSC-52")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("expected 'not supported' in error, got: %v", err)
	}
}

func TestOSC52AvailableWithCaps(t *testing.T) {
	var buf bytes.Buffer

	// With OSC52 support.
	caps := &terminal.TerminalCaps{OSC52Clipboard: true}
	c := NewOSC52Clipboard(&buf, caps)
	if !c.Available() {
		t.Fatal("expected Available() to be true when caps indicate OSC52 support")
	}

	// Without OSC52 support.
	caps = &terminal.TerminalCaps{OSC52Clipboard: false}
	c = NewOSC52Clipboard(&buf, caps)
	if c.Available() {
		t.Fatal("expected Available() to be false when caps lack OSC52 support")
	}
}

func TestOSC52NilReceiver(t *testing.T) {
	var c *OSC52Clipboard
	if c.Available() {
		t.Fatal("nil receiver should not be available")
	}
}

func TestOSC52NilWriter(t *testing.T) {
	caps := &terminal.TerminalCaps{OSC52Clipboard: true}
	c := NewOSC52Clipboard(nil, caps)

	err := c.Write("test")
	if err == nil {
		t.Fatal("expected error when writer is nil")
	}
}

func TestOSC52MultipleWrites(t *testing.T) {
	var buf bytes.Buffer
	caps := &terminal.TerminalCaps{OSC52Clipboard: true}
	c := NewOSC52Clipboard(&buf, caps)

	texts := []string{"first", "second", "third"}
	for _, text := range texts {
		if err := c.Write(text); err != nil {
			t.Fatalf("Write %q failed: %v", text, err)
		}
	}

	// All sequences should be in the buffer.
	got := buf.String()
	for _, text := range texts {
		encoded := base64.StdEncoding.EncodeToString([]byte(text))
		seq := "\033]52;c;" + encoded + "\a"
		if !strings.Contains(got, seq) {
			t.Fatalf("buffer missing sequence for %q", text)
		}
	}
}

func TestOSC52AutoDetectCaps(t *testing.T) {
	// When nil caps are passed, DetectCaps() is called.
	// We just verify it doesn't panic.
	var buf bytes.Buffer
	c := NewOSC52Clipboard(&buf, nil)
	if c == nil {
		t.Fatal("expected non-nil clipboard")
	}
	// Available() should work without panic regardless of detected caps.
	_ = c.Available()
}
