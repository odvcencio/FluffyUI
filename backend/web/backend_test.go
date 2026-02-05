//go:build !js

package web

import (
	"context"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/terminal"
)

func TestNew(t *testing.T) {
	cfg := Config{Addr: ":0"}
	be := New(cfg)
	if be == nil {
		t.Fatal("New returned nil")
	}
	if be.config.Addr != ":0" {
		t.Errorf("expected Addr :0, got %s", be.config.Addr)
	}
}

func TestNewWithAddr(t *testing.T) {
	be := NewWithAddr(":9090")
	if be == nil {
		t.Fatal("NewWithAddr returned nil")
	}
	if be.config.Addr != ":9090" {
		t.Errorf("expected Addr :9090, got %s", be.config.Addr)
	}
}

func TestBackendInit(t *testing.T) {
	be := NewWithAddr(":0")
	if err := be.Init(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	defer be.Fini()

	// Check that we got a real address
	addr := be.Addr()
	if addr == nil {
		t.Fatal("Addr returned nil after Init")
	}
	if addr.String() == "" {
		t.Error("Addr is empty")
	}

	// Check URL is generated
	url := be.URL()
	if url == "" {
		t.Error("URL is empty")
	}
}

func TestBackendSize(t *testing.T) {
	be := NewWithAddr(":0")
	be.width = 100
	be.height = 50

	w, h := be.Size()
	if w != 100 || h != 50 {
		t.Errorf("Size() = (%d, %d), want (100, 50)", w, h)
	}
}

func TestSetContent(t *testing.T) {
	be := NewWithAddr(":0")
	be.width = 10
	be.height = 10
	be.cells = make([]backend.Cell, 100)

	style := backend.DefaultStyle().Foreground(backend.ColorRed)
	be.SetContent(5, 5, 'X', nil, style)

	idx := 5*10 + 5
	if be.cells[idx].Rune != 'X' {
		t.Errorf("cell at (5,5) has rune %q, want 'X'", be.cells[idx].Rune)
	}
}

func TestSetContentOutOfBounds(t *testing.T) {
	be := NewWithAddr(":0")
	be.width = 10
	be.height = 10
	be.cells = make([]backend.Cell, 100)

	// These should not panic
	be.SetContent(-1, 0, 'X', nil, backend.DefaultStyle())
	be.SetContent(0, -1, 'X', nil, backend.DefaultStyle())
	be.SetContent(10, 0, 'X', nil, backend.DefaultStyle())
	be.SetContent(0, 10, 'X', nil, backend.DefaultStyle())
}

func TestClear(t *testing.T) {
	be := NewWithAddr(":0")
	be.width = 10
	be.height = 10
	be.cells = make([]backend.Cell, 100)

	// Set some content
	be.SetContent(5, 5, 'X', nil, backend.DefaultStyle())

	// Clear
	be.Clear()

	// Check all cells are cleared
	for i, cell := range be.cells {
		if cell.Rune != ' ' && cell.Rune != 0 {
			t.Errorf("cell %d not cleared, has rune %q", i, cell.Rune)
		}
	}
}

func TestCursor(t *testing.T) {
	be := NewWithAddr(":0")
	be.width = 80
	be.height = 24
	be.cells = make([]backend.Cell, 80*24)

	// Test cursor positioning
	be.SetCursorPos(10, 5)
	if be.cursorX != 10 || be.cursorY != 5 {
		t.Errorf("cursor position = (%d, %d), want (10, 5)", be.cursorX, be.cursorY)
	}

	// Test show/hide cursor
	be.cursorOn = false
	be.ShowCursor()
	if !be.cursorOn {
		t.Error("cursor should be visible after ShowCursor")
	}

	be.HideCursor()
	if be.cursorOn {
		t.Error("cursor should be hidden after HideCursor")
	}
}

func TestPostEvent(t *testing.T) {
	be := NewWithAddr(":0")
	be.ctx, be.cancel = context.WithCancel(context.Background())
	t.Cleanup(be.cancel)

	// Post a key event
	ev := terminal.KeyEvent{Key: terminal.KeyEnter}
	if err := be.PostEvent(ev); err != nil {
		t.Errorf("PostEvent failed: %v", err)
	}

	// Verify we can poll it
	select {
	case polled := <-be.events:
		if polled != ev {
			t.Error("polled event doesn't match posted event")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for event")
	}
}

func TestRenderFullANSI_TracksBackgroundTransitions(t *testing.T) {
	be := NewWithAddr(":0")
	be.width = 2
	be.height = 1
	be.cells = make([]backend.Cell, 2)

	style1 := backend.DefaultStyle().
		Foreground(backend.ColorWhite).
		Background(backend.ColorBlue)
	style2 := backend.DefaultStyle().
		Foreground(backend.ColorWhite).
		Background(backend.ColorGreen)

	be.SetContent(0, 0, 'A', nil, style1)
	be.SetContent(1, 0, 'B', nil, style2)

	ansi := string(be.renderFullANSI())
	if !contains([]byte(ansi), []byte("48;5;")) {
		t.Fatalf("expected ANSI output to contain background color codes: %q", ansi)
	}
	// Each cell has a distinct background, so output should contain at least two
	// background selectors.
	if count := countSubstr(ansi, "48;5;"); count < 2 {
		t.Fatalf("expected at least 2 background transitions, got %d in %q", count, ansi)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Addr != ":8080" {
		t.Errorf("default Addr = %q, want :8080", cfg.Addr)
	}
	if cfg.Renderer != RendererANSI {
		t.Error("default Renderer should be RendererANSI")
	}
	if cfg.Path != "/" {
		t.Errorf("default Path = %q, want /", cfg.Path)
	}
	if cfg.WSPath != "/ws" {
		t.Errorf("default WSPath = %q, want /ws", cfg.WSPath)
	}
}

func TestFillDefaults(t *testing.T) {
	cfg := Config{Addr: ":3000"}
	fillDefaults(&cfg)

	if cfg.Addr != ":3000" {
		t.Error("existing Addr should not be changed")
	}
	if cfg.Path != "/" {
		t.Error("Path should be filled with default")
	}
	if cfg.WSPath != "/ws" {
		t.Error("WSPath should be filled with default")
	}
}

func TestBackendImplementsInterface(t *testing.T) {
	be := NewWithAddr(":0")

	// Verify backend.Backend interface
	var _ backend.Backend = be

	// Verify optional interfaces
	var _ backend.RowWriter = be
	var _ backend.RectWriter = be
}

func TestRenderANSI(t *testing.T) {
	be := NewWithAddr(":0")
	be.width = 3
	be.height = 2
	be.cells = make([]backend.Cell, 6)

	// Set some content
	be.SetContent(0, 0, 'A', nil, backend.DefaultStyle())
	be.SetContent(1, 0, 'B', nil, backend.DefaultStyle())
	be.SetContent(2, 0, 'C', nil, backend.DefaultStyle())
	be.SetContent(0, 1, 'D', nil, backend.DefaultStyle())
	be.SetContent(1, 1, 'E', nil, backend.DefaultStyle())
	be.SetContent(2, 1, 'F', nil, backend.DefaultStyle())

	// Render to ANSI
	ansi := be.renderANSI()
	if len(ansi) == 0 {
		t.Error("renderANSI returned empty data")
	}

	// Should contain clear screen sequence
	if !contains(ansi, []byte("\x1b[2J")) {
		t.Error("ANSI output missing clear screen sequence")
	}
}

func TestStyleToANSI(t *testing.T) {
	be := NewWithAddr(":0")

	tests := []struct {
		name     string
		fg       backend.Color
		bg       backend.Color
		attrs    backend.AttrMask
		contains []byte
	}{
		{
			name:     "default",
			fg:       backend.ColorDefault,
			bg:       backend.ColorDefault,
			attrs:    0,
			contains: []byte("39;49"),
		},
		{
			name:     "bold red on blue",
			fg:       backend.ColorRed,
			bg:       backend.ColorBlue,
			attrs:    backend.AttrBold,
			contains: []byte("1;"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ansi := be.styleToANSI(tt.fg, tt.bg, tt.attrs)
			if !contains(ansi, tt.contains) {
				t.Errorf("ANSI %q doesn't contain %q", ansi, tt.contains)
			}
		})
	}
}

// contains checks if haystack contains needle.
func contains(haystack, needle []byte) bool {
	return len(haystack) >= len(needle) && containsSubslice(haystack, needle)
}

func containsSubslice(haystack, needle []byte) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if string(haystack[i:i+len(needle)]) == string(needle) {
			return true
		}
	}
	return false
}

func countSubstr(s, needle string) int {
	count := 0
	for {
		idx := -1
		if len(s) >= len(needle) {
			for i := 0; i <= len(s)-len(needle); i++ {
				if s[i:i+len(needle)] == needle {
					idx = i
					break
				}
			}
		}
		if idx < 0 {
			return count
		}
		count++
		s = s[idx+len(needle):]
	}
}
