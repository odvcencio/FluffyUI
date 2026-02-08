package widgets

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
	fluffytest "github.com/odvcencio/fluffyui/testing"
)

func TestBadge_NewBadge(t *testing.T) {
	b := NewBadge("NEW")
	if b == nil {
		t.Fatal("expected non-nil badge")
	}
	if b.Text() != "NEW" {
		t.Fatalf("expected text 'NEW', got %q", b.Text())
	}
	if b.AccessibleRole() != "status" {
		t.Fatalf("expected role status, got %q", b.AccessibleRole())
	}
	if b.AccessibleLive() != "polite" {
		t.Fatalf("expected live polite, got %q", b.AccessibleLive())
	}
	if b.CanFocus() {
		t.Fatal("expected badge to not be focusable")
	}
}

func TestBadge_SetText(t *testing.T) {
	b := NewBadge("OLD")
	b.SetText("UPDATED")
	if b.Text() != "UPDATED" {
		t.Fatalf("expected text 'UPDATED', got %q", b.Text())
	}
	if b.AccessibleLabel() != "UPDATED" {
		t.Fatalf("expected label 'UPDATED', got %q", b.AccessibleLabel())
	}
}

func TestBadge_Measure(t *testing.T) {
	b := NewBadge("INFO")
	size := b.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	if size.Width != 4 {
		t.Fatalf("expected width 4 for 'INFO', got %d", size.Width)
	}
	if size.Height != 1 {
		t.Fatalf("expected height 1, got %d", size.Height)
	}
}

func TestBadge_Render(t *testing.T) {
	b := NewBadge("OK")
	output := fluffytest.RenderToString(b, 10, 1)
	if !strings.Contains(output, "OK") {
		t.Fatalf("expected output to contain 'OK', got %q", output)
	}
}

func TestBadge_HandleMessage(t *testing.T) {
	b := NewBadge("test")
	result := b.HandleMessage(runtime.KeyMsg{})
	if result.Handled {
		t.Fatal("expected badge to return unhandled")
	}
}
