package widgets

import (
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/animation"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/terminal"
	flufftest "github.com/odvcencio/fluffyui/testing"
)

func TestAutoCompleteFlow(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetOptions([]string{"Alpha", "Beta", "Gamma"})
	ac.SetMaxSuggestions(2)
	ac.SetLabel("Search")
	ac.SetQuery("a")

	selected := ""
	ac.SetOnSelect(func(value string) {
		selected = value
	})

	ac.Focus()
	ac.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 4})
	ac.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	ac.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if selected == "" {
		t.Fatalf("expected selection callback to fire")
	}

	ac.SetProvider(func(query string) []string {
		return []string{"Zeta"}
	})
	ac.SetQuery("z")
	if got := ac.Query(); got == "" {
		t.Fatalf("expected query to return text")
	}

	out := flufftest.RenderToString(ac, 20, 4)
	if !strings.Contains(out, "Zeta") {
		t.Fatalf("expected provider suggestion to render, got:\n%s", out)
	}

	ac.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	ac.Blur()

	if ac.Input() == nil {
		t.Fatalf("expected input child")
	}
	_ = ac.ChildWidgets()
}

func TestButtonOptionsAndInteraction(t *testing.T) {
	loading := state.NewSignal(false)
	disabled := state.NewSignal(false)
	clicked := 0

	btn := NewButton("Save",
		WithVariant(VariantPrimary),
		WithClass("cta"),
		WithClasses("rounded", "shadow"),
		WithLoading(loading),
	)
	btn.SetVariant(VariantSecondary)
	btn.Primary().Secondary().Danger()
	btn.Disabled(disabled)
	btn.Loading(loading)
	btn.SetOnClick(func() { clicked++ })
	btn.OnClick(func() { clicked++ })
	btn.Class("primary").Classes("a", "b")
	btn.SetLabel("Commit")
	btn.SetStyle(backend.DefaultStyle().Bold(true))
	btn.SetFocusStyle(backend.DefaultStyle().Underline(true))
	btn.SetDisabledStyle(backend.DefaultStyle().Dim(true))
	_ = btn.StyleType()
	_ = btn.StyleClasses()

	btn.Bind(runtime.Services{})
	btn.Focus()
	btn.Layout(runtime.Rect{X: 0, Y: 0, Width: 12, Height: 1})
	btn.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	btn.Unbind()

	out := flufftest.RenderToString(btn, 12, 1)
	if strings.TrimSpace(out) == "" {
		t.Fatalf("expected button to render label")
	}
	if clicked == 0 {
		t.Fatalf("expected click handler to run")
	}
}

func TestAnimatedWidgetHelpers(t *testing.T) {
	aw := NewAnimatedWidget()
	app := runtime.NewApp(runtime.AppConfig{Animator: animation.NewAnimator()})
	aw.Bind(app.Services())

	aw.Animate("Opacity", animation.Float64(0.5), animation.TweenConfig{Duration: time.Millisecond})
	aw.FadeIn(10 * time.Millisecond)
	aw.FadeOut(10*time.Millisecond, func() {})
	aw.SlideIn(DirectionLeft, 4, 10*time.Millisecond)
	aw.SlideIn(DirectionRight, 4, 10*time.Millisecond)
	aw.SlideIn(DirectionUp, 4, 10*time.Millisecond)
	aw.SlideIn(DirectionDown, 4, 10*time.Millisecond)

	aw.Unbind()
}

func TestAspectRatioAndAlert(t *testing.T) {
	child := NewLabel("Child")
	ar := NewAspectRatio(child, 1.5)
	ar.SetRatio(2)
	ar.SetLabel("Aspect")
	ar.Layout(runtime.Rect{X: 0, Y: 0, Width: 10, Height: 4})
	_ = flufftest.RenderToString(ar, 10, 4)
	_ = ar.ChildWidgets()
	ar.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'x'})

	alert := NewAlert("Warning", AlertWarning)
	alert.SetStyle(backend.DefaultStyle().Bold(true))
	alert.SetClasses("notice")
	_ = alert.StyleType()
	_ = alert.StyleClasses()
	alert.ApplyStyle(style.Style{})
	_ = flufftest.RenderToString(alert, 20, 1)
	alert.HandleMessage(runtime.TickMsg{})
}

func TestBaseMetadataAndHelpers(t *testing.T) {
	base := &Base{}
	base.SetID(" widget ")
	base.SetKey(" key ")
	if base.ID() == "" || base.Key() == "" {
		t.Fatalf("expected id/key to be set")
	}
	base.SetClasses("a", "b", "a")
	base.AddClass("c")
	base.ApplyStyle(style.Style{})
	_ = base.LayoutStyle()
	_ = base.StyleState()
	_ = base.StyleID()

	classes := normalizeClasses([]string{" a ", "", "b", "a"})
	if len(classes) == 0 {
		t.Fatalf("expected normalized classes")
	}
	if got := clipStringRight("hello", 3); got == "" {
		t.Fatalf("expected clipped string")
	}
	if got := padRight("hi", 4); got == "" {
		t.Fatalf("expected padded string")
	}
}
