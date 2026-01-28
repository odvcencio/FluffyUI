package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
)

type mockFocusable struct {
	Base
}

func (m *mockFocusable) Measure(constraints runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 4, Height: 1}
}

func (m *mockFocusable) Layout(bounds runtime.Rect) {
	m.Base.Layout(bounds)
}

func (m *mockFocusable) Render(ctx runtime.RenderContext) {}

func (m *mockFocusable) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

func (m *mockFocusable) CanFocus() bool { return true }

func TestTooltipHoverOpens(t *testing.T) {
	target := NewLabel("Target")
	content := NewPanel(NewLabel("Tip")).WithBorder(backend.DefaultStyle())
	tt := NewTooltip(target, content, WithTooltipTrigger(TooltipHover))
	tt.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 1})

	result := tt.HandleMessage(runtime.MouseMsg{X: 1, Y: 0, Button: runtime.MouseNone, Action: runtime.MouseMove})
	if len(result.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(result.Commands))
	}
	if _, ok := result.Commands[0].(runtime.PushOverlay); !ok {
		t.Fatalf("expected PushOverlay, got %T", result.Commands[0])
	}
}

func TestTooltipClickToggles(t *testing.T) {
	target := NewLabel("Target")
	content := NewLabel("Tip")
	tt := NewTooltip(target, content, WithTooltipTrigger(TooltipClick))
	tt.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 1})

	open := tt.HandleMessage(runtime.MouseMsg{X: 1, Y: 0, Button: runtime.MouseLeft, Action: runtime.MousePress})
	if len(open.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(open.Commands))
	}
	if _, ok := open.Commands[0].(runtime.PushOverlay); !ok {
		t.Fatalf("expected PushOverlay, got %T", open.Commands[0])
	}

	close := tt.HandleMessage(runtime.MouseMsg{X: 1, Y: 0, Button: runtime.MouseLeft, Action: runtime.MousePress})
	if len(close.Commands) != 1 {
		t.Fatalf("expected 1 command on close, got %d", len(close.Commands))
	}
	if _, ok := close.Commands[0].(runtime.PopOverlay); !ok {
		t.Fatalf("expected PopOverlay, got %T", close.Commands[0])
	}
}

func TestTooltipFocusOpens(t *testing.T) {
	target := &mockFocusable{}
	content := NewLabel("Tip")
	tt := NewTooltip(target, content, WithTooltipTrigger(TooltipFocus))
	tt.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 1})

	open := tt.HandleMessage(runtime.FocusChangedMsg{Prev: nil, Next: target})
	if len(open.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(open.Commands))
	}
	if _, ok := open.Commands[0].(runtime.PushOverlay); !ok {
		t.Fatalf("expected PushOverlay, got %T", open.Commands[0])
	}

	close := tt.HandleMessage(runtime.FocusChangedMsg{Prev: target, Next: nil})
	if len(close.Commands) != 1 {
		t.Fatalf("expected 1 command on close, got %d", len(close.Commands))
	}
	if _, ok := close.Commands[0].(runtime.PopOverlay); !ok {
		t.Fatalf("expected PopOverlay, got %T", close.Commands[0])
	}
}
