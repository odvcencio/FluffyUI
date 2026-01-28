package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

func TestSelect_DropdownModeOpensOverlay(t *testing.T) {
	selecter := NewSelect(
		SelectOption{Label: "One"},
		SelectOption{Label: "Two"},
	).Apply(WithDropdownMode())
	selecter.Focus()
	selecter.Layout(runtime.Rect{X: 0, Y: 0, Width: 12, Height: 1})

	result := selecter.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Fatal("expected Enter to be handled in dropdown mode")
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	cmd, ok := result.Commands[0].(runtime.PushOverlay)
	if !ok {
		t.Fatalf("expected PushOverlay command, got %T", result.Commands[0])
	}
	if cmd.Widget == nil {
		t.Fatal("expected PushOverlay.Widget to be set")
	}
}

func TestSelect_DropdownModeMouseOpensOverlay(t *testing.T) {
	selecter := NewSelect(
		SelectOption{Label: "One"},
		SelectOption{Label: "Two"},
	).Apply(WithDropdownMode())
	selecter.Layout(runtime.Rect{X: 0, Y: 0, Width: 12, Height: 1})

	result := selecter.HandleMessage(runtime.MouseMsg{X: 1, Y: 0, Button: runtime.MouseLeft, Action: runtime.MousePress})
	if !result.Handled {
		t.Fatal("expected mouse press to be handled in dropdown mode")
	}
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	if _, ok := result.Commands[0].(runtime.PushOverlay); !ok {
		t.Fatalf("expected PushOverlay command, got %T", result.Commands[0])
	}
}
