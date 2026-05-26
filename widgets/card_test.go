package widgets

import (
	"strings"
	"testing"

	"m31labs.dev/fluffyui/runtime"
	fluffytest "m31labs.dev/fluffyui/testing"
)

func TestCard_NewCard(t *testing.T) {
	body := NewLabel("Hello")
	card := NewCard("My Card", body)
	if card == nil {
		t.Fatal("expected non-nil card")
	}
	if card.Title() != "My Card" {
		t.Fatalf("expected title 'My Card', got %q", card.Title())
	}
	if card.AccessibleRole() != "group" {
		t.Fatalf("expected role group, got %q", card.AccessibleRole())
	}
	if card.AccessibleLabel() != "My Card" {
		t.Fatalf("expected label 'My Card', got %q", card.AccessibleLabel())
	}
}

func TestCard_Measure(t *testing.T) {
	body := NewLabel("Content")
	card := NewCard("Title", body)
	size := card.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 20})
	if size.Width <= 0 {
		t.Fatalf("expected positive width, got %d", size.Width)
	}
	if size.Height <= 0 {
		t.Fatalf("expected positive height, got %d", size.Height)
	}
}

func TestCard_ChildWidgets(t *testing.T) {
	header := NewLabel("Header")
	body := NewLabel("Body")
	footer := NewLabel("Footer")
	card := NewCard("Test", body)
	card.SetHeader(header)
	card.SetFooter(footer)

	children := card.ChildWidgets()
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}

	// Only body.
	card2 := NewCard("Test2", body)
	children2 := card2.ChildWidgets()
	if len(children2) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children2))
	}
}

func TestCard_SetTitle(t *testing.T) {
	card := NewCard("Initial", NewLabel("body"))
	card.SetTitle("Updated")
	if card.Title() != "Updated" {
		t.Fatalf("expected title 'Updated', got %q", card.Title())
	}
	if card.AccessibleLabel() != "Updated" {
		t.Fatalf("expected label 'Updated', got %q", card.AccessibleLabel())
	}
}

func TestCard_Render(t *testing.T) {
	body := NewLabel("Content")
	card := NewCard("My Card", body)
	output := fluffytest.RenderToString(card, 20, 5)
	if !strings.Contains(output, "My Card") {
		t.Fatalf("expected output to contain title, got %q", output)
	}
}

func TestCard_NoBorder(t *testing.T) {
	body := NewLabel("X")
	card := NewCard("Title", body)
	card.SetBorder(false)
	size := card.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 20})
	// Without border the card should be smaller.
	if size.Width <= 0 || size.Height <= 0 {
		t.Fatalf("expected positive dimensions without border, got %dx%d", size.Width, size.Height)
	}
}

func TestCard_HandleMessage(t *testing.T) {
	card := NewCard("Title", NewLabel("body"))
	result := card.HandleMessage(runtime.KeyMsg{})
	if result.Handled {
		t.Fatal("expected card to return unhandled")
	}
}
