package clipboard

import (
	"bytes"
	"testing"
)

func TestAutoClipboardFallbackChain(t *testing.T) {
	// With nil writer, OSC-52 is skipped. If no platform tools are
	// available, it falls back to memory clipboard.
	ac := NewAutoClipboard(nil)
	if ac == nil {
		t.Fatal("expected non-nil auto clipboard")
	}
	if !ac.Available() {
		t.Fatal("auto clipboard should always be available")
	}

	// Write and read should work via fallback.
	if err := ac.Write("test"); err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	got, err := ac.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if got != "test" {
		t.Fatalf("Read = %q, want %q", got, "test")
	}
}

func TestAutoClipboardWithWriter(t *testing.T) {
	var buf bytes.Buffer
	ac := NewAutoClipboard(&buf)
	if ac == nil {
		t.Fatal("expected non-nil auto clipboard")
	}
	if !ac.Available() {
		t.Fatal("auto clipboard should always be available")
	}

	// Primary should not be nil.
	if ac.Primary() == nil {
		t.Fatal("expected non-nil primary clipboard")
	}

	// Reader should not be nil.
	if ac.Reader() == nil {
		t.Fatal("expected non-nil reader clipboard")
	}
}

func TestAutoClipboardNilReceiver(t *testing.T) {
	var ac *AutoClipboard
	if ac.Available() {
		t.Fatal("nil auto clipboard should not be available")
	}

	got, err := ac.Read()
	if err != nil {
		t.Fatalf("nil Read should not error: %v", err)
	}
	if got != "" {
		t.Fatalf("nil Read = %q, want empty", got)
	}

	if err := ac.Write("test"); err != nil {
		t.Fatalf("nil Write should not error: %v", err)
	}

	if ac.Primary() != nil {
		t.Fatal("nil Primary should return nil")
	}
	if ac.Reader() != nil {
		t.Fatal("nil Reader should return nil")
	}
}

func TestAutoClipboardAlwaysAvailable(t *testing.T) {
	// Even without any real clipboard support, Available should be true
	// because the in-memory fallback is always there.
	ac := NewAutoClipboard(nil)
	if !ac.Available() {
		t.Fatal("auto clipboard should always report available")
	}
}

func TestAutoClipboardWriteAlsoWritesFallback(t *testing.T) {
	// When primary is not the fallback, writes should also go to fallback
	// so that reads work even without platform tools.
	ac := NewAutoClipboard(nil)
	if err := ac.Write("sync test"); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Read from the fallback (which is the reader if no platform tools).
	got, err := ac.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if got != "sync test" {
		t.Fatalf("Read = %q, want %q", got, "sync test")
	}
}

func TestAutoClipboardMultipleWrites(t *testing.T) {
	ac := NewAutoClipboard(nil)

	texts := []string{"first", "second", "third"}
	for _, text := range texts {
		if err := ac.Write(text); err != nil {
			t.Fatalf("Write %q failed: %v", text, err)
		}
	}

	// Should read the last written value.
	got, err := ac.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if got != "third" {
		t.Fatalf("Read = %q, want %q", got, "third")
	}
}
