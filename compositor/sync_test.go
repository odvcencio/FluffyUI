package compositor

import (
	"strings"
	"testing"
)

func TestSyncBegin(t *testing.T) {
	got := SyncBegin()
	want := "\x1b[?2026h"
	if got != want {
		t.Errorf("SyncBegin() = %q, want %q", got, want)
	}
}

func TestSyncEnd(t *testing.T) {
	got := SyncEnd()
	want := "\x1b[?2026l"
	if got != want {
		t.Errorf("SyncEnd() = %q, want %q", got, want)
	}
}

func TestSyncBeginMatchesConstant(t *testing.T) {
	if SyncBegin() != ANSISyncStart {
		t.Error("SyncBegin() should match ANSISyncStart constant")
	}
}

func TestSyncEndMatchesConstant(t *testing.T) {
	if SyncEnd() != ANSISyncEnd {
		t.Error("SyncEnd() should match ANSISyncEnd constant")
	}
}

func TestWrapSync_Enabled(t *testing.T) {
	content := "frame data"
	got := WrapSync(content, true)
	want := SyncBegin() + content + SyncEnd()
	if got != want {
		t.Errorf("WrapSync(content, true) = %q, want %q", got, want)
	}
}

func TestWrapSync_Disabled(t *testing.T) {
	content := "frame data"
	got := WrapSync(content, false)
	if got != content {
		t.Errorf("WrapSync(content, false) = %q, want %q (unchanged)", got, content)
	}
}

func TestWrapSync_EmptyContent(t *testing.T) {
	got := WrapSync("", true)
	if got != "" {
		t.Errorf("WrapSync(\"\", true) = %q, want empty string", got)
	}
}

func TestRenderer_SyncOutput_Default(t *testing.T) {
	screen := NewScreen(10, 5)
	r := NewRenderer(screen)
	if !r.SyncOutput() {
		t.Error("sync output should be enabled by default")
	}
}

func TestRenderer_SetSyncOutput(t *testing.T) {
	screen := NewScreen(10, 5)
	r := NewRenderer(screen)

	r.SetSyncOutput(false)
	if r.SyncOutput() {
		t.Error("sync output should be disabled after SetSyncOutput(false)")
	}

	r.SetSyncOutput(true)
	if !r.SyncOutput() {
		t.Error("sync output should be enabled after SetSyncOutput(true)")
	}
}

func TestRender_WithSyncEnabled(t *testing.T) {
	screen := NewScreen(10, 2)
	screen.Set(0, 0, 'X', DefaultStyle())
	r := NewRenderer(screen)
	r.SetSyncOutput(true)

	output := r.Render()

	if !strings.Contains(output, ANSISyncStart) {
		t.Error("render output should contain sync start when sync enabled")
	}
	if !strings.Contains(output, ANSISyncEnd) {
		t.Error("render output should contain sync end when sync enabled")
	}

	// Verify ordering: sync start comes before sync end
	startIdx := strings.Index(output, ANSISyncStart)
	endIdx := strings.Index(output, ANSISyncEnd)
	if startIdx >= endIdx {
		t.Error("sync start should come before sync end")
	}
}

func TestRender_WithSyncDisabled(t *testing.T) {
	screen := NewScreen(10, 2)
	screen.Set(0, 0, 'X', DefaultStyle())
	r := NewRenderer(screen)
	r.SetSyncOutput(false)

	output := r.Render()

	if strings.Contains(output, ANSISyncStart) {
		t.Error("render output should NOT contain sync start when sync disabled")
	}
	if strings.Contains(output, ANSISyncEnd) {
		t.Error("render output should NOT contain sync end when sync disabled")
	}
}

func TestRenderFull_WithSyncEnabled(t *testing.T) {
	screen := NewScreen(10, 2)
	screen.Set(0, 0, 'A', DefaultStyle())
	r := NewRenderer(screen)
	r.SetSyncOutput(true)

	output := r.RenderFull()

	if !strings.Contains(output, ANSISyncStart) {
		t.Error("full render output should contain sync start when sync enabled")
	}
	if !strings.Contains(output, ANSISyncEnd) {
		t.Error("full render output should contain sync end when sync enabled")
	}

	// Content should still be present
	if !strings.Contains(output, "A") {
		t.Error("full render should contain the character data")
	}
}

func TestRenderFull_WithSyncDisabled(t *testing.T) {
	screen := NewScreen(10, 2)
	screen.Set(0, 0, 'A', DefaultStyle())
	r := NewRenderer(screen)
	r.SetSyncOutput(false)

	output := r.RenderFull()

	if strings.Contains(output, ANSISyncStart) {
		t.Error("full render output should NOT contain sync start when sync disabled")
	}
	if strings.Contains(output, ANSISyncEnd) {
		t.Error("full render output should NOT contain sync end when sync disabled")
	}

	// Content should still be present even without sync
	if !strings.Contains(output, "A") {
		t.Error("full render should contain character data regardless of sync setting")
	}
}

func TestCompositor_SetSyncOutput(t *testing.T) {
	c := NewCompositor(10, 5)
	c.Screen().Set(0, 0, 'Z', DefaultStyle())

	// Default: sync enabled
	output := c.Render()
	if !strings.Contains(output, ANSISyncStart) {
		t.Error("compositor render should include sync by default")
	}

	// Disable sync
	c.SetSyncOutput(false)
	c.Screen().Set(0, 0, 'Y', DefaultStyle())
	output = c.Render()
	if strings.Contains(output, ANSISyncStart) {
		t.Error("compositor render should NOT include sync after disabling")
	}

	// Re-enable sync
	c.SetSyncOutput(true)
	c.Screen().Set(0, 0, 'W', DefaultStyle())
	output = c.Render()
	if !strings.Contains(output, ANSISyncStart) {
		t.Error("compositor render should include sync after re-enabling")
	}
}

func TestRender_SyncWrapsEntireFrame(t *testing.T) {
	screen := NewScreen(10, 2)
	screen.Set(3, 0, 'M', DefaultStyle())
	r := NewRenderer(screen)
	r.SetSyncOutput(true)

	output := r.Render()

	// The sync start should be at the very beginning
	if !strings.HasPrefix(output, ANSISyncStart) {
		t.Error("sync start should be at the beginning of frame output")
	}

	// The sync end should be at the very end
	if !strings.HasSuffix(output, ANSISyncEnd) {
		t.Error("sync end should be at the end of frame output")
	}
}

func TestRenderFull_SyncWrapsEntireFrame(t *testing.T) {
	screen := NewScreen(10, 2)
	screen.Set(0, 0, 'N', DefaultStyle())
	r := NewRenderer(screen)
	r.SetSyncOutput(true)

	output := r.RenderFull()

	if !strings.HasPrefix(output, ANSISyncStart) {
		t.Error("sync start should be at the beginning of full frame output")
	}
	if !strings.HasSuffix(output, ANSISyncEnd) {
		t.Error("sync end should be at the end of full frame output")
	}
}
