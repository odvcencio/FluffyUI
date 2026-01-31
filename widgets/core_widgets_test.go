package widgets

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/terminal"
	fluffytest "github.com/odvcencio/fluffyui/testing"
)

func TestLabelRender(t *testing.T) {
	label := NewLabel("Hello")
	output := fluffytest.RenderToString(label, 10, 1)
	if !strings.Contains(output, "Hello") {
		t.Fatalf("expected output to contain label text, got %q", output)
	}
}

func TestButtonClick(t *testing.T) {
	clicked := false
	btn := NewButton("OK", WithOnClick(func() {
		clicked = true
	}))
	btn.Focus()
	result := btn.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Fatalf("expected handled result")
	}
	if !clicked {
		t.Fatalf("expected click handler to run")
	}
}

func TestButtonDisabledNoClick(t *testing.T) {
	disabled := state.NewSignal(true)
	clicked := false
	btn := NewButton("OK", WithDisabled(disabled), WithOnClick(func() {
		clicked = true
	}))
	btn.Focus()
	result := btn.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if result.Handled {
		t.Fatalf("expected disabled button to ignore input")
	}
	if clicked {
		t.Fatalf("expected disabled button to not invoke handler")
	}
}

func TestInputTypingAndSubmit(t *testing.T) {
	input := NewInput()
	input.Focus()
	input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'a'})
	input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'b'})
	input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'c'})

	if got := input.Text(); got != "abc" {
		t.Fatalf("expected text 'abc', got %q", got)
	}

	var submitted string
	input.SetOnSubmit(func(text string) {
		submitted = text
	})
	input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if submitted != "abc" {
		t.Fatalf("expected submit 'abc', got %q", submitted)
	}
}

func TestInputBackspace(t *testing.T) {
	input := NewInput()
	input.Focus()
	input.SetText("ab")
	input.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	if got := input.Text(); got != "a" {
		t.Fatalf("expected text 'a', got %q", got)
	}
}

func TestTabsRender(t *testing.T) {
	tabs := NewTabs(
		Tab{Title: "One", Content: NewLabel("First")},
		Tab{Title: "Two", Content: NewLabel("Second")},
	)
	output := fluffytest.RenderToString(tabs, 20, 3)
	if !strings.Contains(output, "One") {
		t.Fatalf("expected tab title in output, got %q", output)
	}
	if !strings.Contains(output, "First") {
		t.Fatalf("expected selected tab content, got %q", output)
	}
}

func TestSearchWidgetQuery(t *testing.T) {
	search := NewSearchWidget()
	search.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'f'})
	search.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'o'})
	search.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'o'})
	if search.Query() != "foo" {
		t.Fatalf("expected query 'foo', got %q", search.Query())
	}
	search.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	if search.Query() != "fo" {
		t.Fatalf("expected query 'fo', got %q", search.Query())
	}
}

func TestGaugeString(t *testing.T) {
	got := DrawGaugeString(10, 0.3, GaugeStyle{})
	want := "███░░░░░░░"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestAspectRatioLayout(t *testing.T) {
	child := NewLabel("X")
	container := NewAspectRatio(child, 2.0)
	container.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 10})
	bounds := child.Bounds()
	if bounds.Width != 10 || bounds.Height != 5 {
		t.Fatalf("expected child size 10x5, got %dx%d", bounds.Width, bounds.Height)
	}
	if bounds.X != 0 || bounds.Y != 2 {
		t.Fatalf("expected child positioned at (0,2), got (%d,%d)", bounds.X, bounds.Y)
	}
}

func TestAlertRender(t *testing.T) {
	alert := NewAlert("Boom", AlertError)
	output := fluffytest.RenderToString(alert, 10, 1)
	if !strings.Contains(output, "Boom") {
		t.Fatalf("expected alert text in output, got %q", output)
	}
}

func TestSelectNavigation(t *testing.T) {
	selectWidget := NewSelect(
		SelectOption{Label: "One", Value: 1},
		SelectOption{Label: "Two", Value: 2},
	)
	selectWidget.Focus()
	selectWidget.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if got := selectWidget.Selected(); got != 1 {
		t.Fatalf("expected selected index 1, got %d", got)
	}
}

func TestSelectDropdownOpens(t *testing.T) {
	selectWidget := NewSelect(
		SelectOption{Label: "One", Value: 1},
	).Apply(WithDropdownMode())
	selectWidget.Focus()
	result := selectWidget.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if len(result.Commands) != 1 {
		t.Fatalf("expected push overlay command")
	}
	if _, ok := result.Commands[0].(runtime.PushOverlay); !ok {
		t.Fatalf("expected PushOverlay command, got %T", result.Commands[0])
	}
}

func TestTooltipClickOpens(t *testing.T) {
	target := NewLabel("Target")
	content := NewLabel("Tip")
	tooltip := NewTooltip(target, content, WithTooltipTrigger(TooltipClick))
	tooltip.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 1})
	result := tooltip.HandleMessage(runtime.MouseMsg{X: 1, Y: 0, Button: runtime.MouseLeft, Action: runtime.MousePress})
	if len(result.Commands) == 0 {
		t.Fatalf("expected tooltip to push overlay")
	}
	if _, ok := result.Commands[0].(runtime.PushOverlay); !ok {
		t.Fatalf("expected PushOverlay command, got %T", result.Commands[0])
	}
}
