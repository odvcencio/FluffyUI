package widgets

import (
	"strings"
	"testing"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

func TestPagination_NewPagination(t *testing.T) {
	p := NewPagination(100, 10)
	if p == nil {
		t.Fatal("expected non-nil pagination")
	}
	if p.Page() != 1 {
		t.Fatalf("expected page 1, got %d", p.Page())
	}
	if p.TotalPages() != 10 {
		t.Fatalf("expected 10 total pages, got %d", p.TotalPages())
	}
	if p.AccessibleRole() != "navigation" {
		t.Fatalf("expected role navigation, got %q", p.AccessibleRole())
	}
}

func TestPagination_Navigate(t *testing.T) {
	p := NewPagination(50, 10)
	p.Focus()

	// Navigate right.
	result := p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if !result.Handled {
		t.Fatal("expected right arrow to be handled")
	}
	if p.Page() != 2 {
		t.Fatalf("expected page 2, got %d", p.Page())
	}

	// Navigate right again.
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if p.Page() != 3 {
		t.Fatalf("expected page 3, got %d", p.Page())
	}

	// Navigate left.
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if p.Page() != 2 {
		t.Fatalf("expected page 2 after left, got %d", p.Page())
	}

	// Navigate to end.
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnd})
	if p.Page() != 5 {
		t.Fatalf("expected page 5 at end, got %d", p.Page())
	}

	// Navigate to home.
	p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyHome})
	if p.Page() != 1 {
		t.Fatalf("expected page 1 at home, got %d", p.Page())
	}

	// Cannot navigate past first page.
	result = p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if result.Handled {
		t.Fatal("expected left arrow at page 1 to be unhandled")
	}
}

func TestPagination_Measure(t *testing.T) {
	p := NewPagination(100, 10)
	size := p.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	if size.Width <= 0 {
		t.Fatalf("expected positive width, got %d", size.Width)
	}
	if size.Height != 1 {
		t.Fatalf("expected height 1, got %d", size.Height)
	}
}

func TestPagination_SetPage(t *testing.T) {
	p := NewPagination(100, 10)

	var changedTo int
	p.SetOnChange(func(page int) {
		changedTo = page
	})

	p.SetPage(5)
	if p.Page() != 5 {
		t.Fatalf("expected page 5, got %d", p.Page())
	}
	if changedTo != 5 {
		t.Fatalf("expected onChange called with 5, got %d", changedTo)
	}

	// Setting to same page should not trigger onChange.
	changedTo = 0
	p.SetPage(5)
	if changedTo != 0 {
		t.Fatal("expected no onChange when setting same page")
	}

	// Clamp to bounds.
	p.SetPage(0)
	if p.Page() != 1 {
		t.Fatalf("expected page clamped to 1, got %d", p.Page())
	}

	p.SetPage(999)
	if p.Page() != 10 {
		t.Fatalf("expected page clamped to 10, got %d", p.Page())
	}
}

func TestPagination_RenderText(t *testing.T) {
	p := NewPagination(100, 10)
	text := p.renderText()
	if !strings.Contains(text, "[1]") {
		t.Fatalf("expected rendered text to contain [1], got %q", text)
	}
	if !strings.Contains(text, "<") || !strings.Contains(text, ">") {
		t.Fatalf("expected rendered text to contain < and >, got %q", text)
	}
}

func TestPagination_UnfocusedNoHandle(t *testing.T) {
	p := NewPagination(100, 10)
	// Do not focus.
	result := p.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRight})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}
