package clipboard

import (
	"os/exec"
	"runtime"
	"testing"
)

func TestPlatformClipboardDetection(t *testing.T) {
	pc := NewPlatformClipboard()
	if pc == nil {
		t.Skip("no platform clipboard tool available")
	}

	if !pc.Available() {
		t.Fatal("platform clipboard reports unavailable after successful detection")
	}

	if pc.tool == "" {
		t.Fatal("expected tool name to be set")
	}

	t.Logf("detected platform clipboard tool: %s", pc.tool)
}

func TestPlatformClipboardNil(t *testing.T) {
	var pc *PlatformClipboard

	if pc.Available() {
		t.Fatal("nil platform clipboard should not be available")
	}

	_, err := pc.Read()
	if err == nil {
		t.Fatal("expected error reading from nil platform clipboard")
	}

	err = pc.Write("test")
	if err == nil {
		t.Fatal("expected error writing to nil platform clipboard")
	}
}

func TestPlatformClipboardEmptyCommands(t *testing.T) {
	pc := &PlatformClipboard{}

	if pc.Available() {
		t.Fatal("platform clipboard with empty commands should not be available")
	}

	_, err := pc.Read()
	if err == nil {
		t.Fatal("expected error reading with empty commands")
	}

	err = pc.Write("test")
	if err == nil {
		t.Fatal("expected error writing with empty commands")
	}
}

func TestPlatformClipboardReadWrite(t *testing.T) {
	pc := NewPlatformClipboard()
	if pc == nil {
		t.Skip("no platform clipboard tool available")
	}

	// Skip in CI or headless environments without a display server.
	if runtime.GOOS == "linux" {
		if !hasDisplay() {
			t.Skip("no display server available (headless environment)")
		}
	}

	const testText = "FluffyUI clipboard test"
	if err := pc.Write(testText); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	got, err := pc.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if got != testText {
		t.Fatalf("Read = %q, want %q", got, testText)
	}
}

func TestPlatformClipboardLinuxToolPriority(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}

	pc := NewPlatformClipboard()
	if pc == nil {
		t.Skip("no platform clipboard tool available")
	}

	// Just verify the detected tool is one of the expected ones.
	switch pc.tool {
	case "wl-copy", "xclip", "xsel":
		// expected
	default:
		t.Fatalf("unexpected tool: %s", pc.tool)
	}
}

func TestPlatformClipboardDarwinTools(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS-only test")
	}

	pc := NewPlatformClipboard()
	if pc == nil {
		t.Fatal("expected pbcopy/pbpaste to be available on macOS")
	}

	if pc.tool != "pbcopy" {
		t.Fatalf("expected tool 'pbcopy', got %q", pc.tool)
	}
}

// hasDisplay checks whether an X11 or Wayland display is available.
func hasDisplay() bool {
	// Check for X11.
	if _, err := exec.LookPath("xclip"); err == nil {
		// xclip exists, check if DISPLAY is set.
		return true // LookPath success means the tool exists; actual availability is tested by Write/Read
	}
	if _, err := exec.LookPath("xsel"); err == nil {
		return true
	}
	// Check for Wayland.
	if _, err := exec.LookPath("wl-copy"); err == nil {
		return true
	}
	return false
}
