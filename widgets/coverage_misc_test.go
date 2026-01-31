package widgets

import (
	"errors"
	"image"
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/animation"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/graphics"
	"github.com/odvcencio/fluffyui/i18n"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	flufftest "github.com/odvcencio/fluffyui/testing"
)

func TestAccordionToggleAndNavigation(t *testing.T) {
	sec1 := NewAccordionSection("First", NewLabel("One"), WithSectionExpanded(true), WithSectionAnimation(0, nil))
	sec2 := NewAccordionSection("Second", NewLabel("Two"))
	sec1.SetTitle("First Updated")
	_ = sec1.Title()
	sec1.SetContent(NewLabel("Alt"))
	sec1.SetDisabled(false)
	_ = sec1.Disabled()
	acc := NewAccordion(sec1, sec2)
	acc.SetAllowMultiple(false)
	acc.SetStyles(backend.DefaultStyle(), backend.DefaultStyle(), backend.DefaultStyle(), backend.DefaultStyle())
	acc.SetLabel("Accordion")
	acc.SetSelected(0)
	_ = acc.StyleType()
	acc.Bind(runtime.Services{})
	acc.Layout(runtime.Rect{X: 0, Y: 0, Width: 20, Height: 6})

	if !sec1.Expanded() {
		t.Fatalf("expected first section expanded")
	}

	acc.Focus()
	acc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	acc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !sec2.Expanded() {
		t.Fatalf("expected second section expanded after enter")
	}
	if sec1.Expanded() {
		t.Fatalf("expected first section collapsed when allowMultiple=false")
	}

	acc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyLeft})
	if sec2.Expanded() {
		t.Fatalf("expected left key to collapse section")
	}

	acc.HandleMessage(runtime.MouseMsg{Button: runtime.MouseLeft, Action: runtime.MousePress, X: 1, Y: 0})
	if !sec1.Expanded() {
		t.Fatalf("expected mouse click to toggle first section")
	}

	acc.ToggleSection(0)
	acc.moveSelectionTo(1)
	_ = acc.ChildWidgets()

	_ = acc.PathSegment(sec1.Content())
	acc.AddSection(NewAccordionSection("Third", NewLabel("Three")))
	acc.SetSections(sec1, sec2)
	acc.Unbind()
}

func TestAsyncImagePlaceholderAndError(t *testing.T) {
	placeholder := NewLabel("Loading image")
	img := NewAsyncImageWithLoader(func() (image.Image, error) {
		return nil, errors.New("boom")
	}, WithAsyncImagePlaceholder(placeholder), WithAsyncImageBlitter(&graphics.BrailleBlitter{}), WithAsyncImageScaleMode(graphics.ScaleNearest), WithAsyncImageScaleToFit(true), WithAsyncImageCenter(true))

	img.Bind(runtime.Services{})
	out := flufftest.RenderToString(img, 20, 1)
	if !strings.Contains(out, "Loading image") {
		t.Fatalf("expected placeholder to render, got:\n%s", out)
	}
	img.Unbind()

	_ = NewAsyncImage("missing.png", WithAsyncImageBlitter(&graphics.QuadrantBlitter{}))

	errImg := NewAsyncImageWithLoader(func() (image.Image, error) {
		return nil, errors.New("boom")
	})
	errImg.setResult(nil, errors.New("boom"))
	out = flufftest.RenderToString(errImg, 20, 1)
	if !strings.Contains(out, "Image error") {
		t.Fatalf("expected error message to render, got:\n%s", out)
	}
	if err := errImg.Error(); err == nil {
		t.Fatalf("expected Error() to wrap loader error")
	}
}

func TestPerformanceDashboardSummary(t *testing.T) {
	sampler := runtime.NewRenderSampler(3)
	sampler.ObserveRender(runtime.RenderStats{TotalDuration: 16 * time.Millisecond, RenderDuration: 10 * time.Millisecond, FlushDuration: 4 * time.Millisecond, TotalCells: 100, DirtyCells: 10, LayerCount: 2})
	bundle := i18n.NewBundle("en")
	bundle.AddMessages("en", map[string]string{"perf": "Performance"})

	dash := NewPerformanceDashboard(sampler, WithPerformanceRefresh(0))
	app := runtime.NewApp(runtime.AppConfig{Localizer: bundle.Localizer("en")})
	dash.Bind(app.Services())

	out := flufftest.RenderToString(dash, 40, 12)
	if !strings.Contains(out, "Render Summary") {
		t.Fatalf("expected render summary to render, got:\n%s", out)
	}

	dash.SetSampler(nil)
	out = flufftest.RenderToString(dash, 40, 2)
	if !strings.Contains(out, "No render sampler") {
		t.Fatalf("expected no sampler message, got:\n%s", out)
	}
}

func TestLineChartSpinnerAndGauge(t *testing.T) {
	chart := NewLineChart()
	chart.AddSeries(ChartSeries{Data: []float64{1, 3, 2, 4}, Color: backend.ColorRGB(100, 200, 255), Smooth: true, Fill: true})
	chart.SetYAxis(0, 5)
	chart.AutoYAxis()
	_ = chart.StyleType()
	_ = flufftest.RenderToString(chart, 20, 6)

	spinner := NewSpinner()
	spinner.Advance()
	spinner.HandleMessage(runtime.TickMsg{})
	out := flufftest.RenderToString(spinner, 2, 1)
	if strings.TrimSpace(out) == "" {
		t.Fatalf("expected spinner to render a frame")
	}

	gauge := NewAnimatedGauge(0, 100)
	app := runtime.NewApp(runtime.AppConfig{Animator: animation.NewAnimator()})
	gauge.Bind(app.Services())
	gauge.SetValue(50)
	_ = gauge.StyleType()
	if gauge.spring != nil {
		gauge.spring.Value = 0.5
	}
	_ = flufftest.RenderToString(gauge, 12, 6)
	gauge.Unbind()
}

func TestBaseStyleClasses(t *testing.T) {
	base := &Base{}
	base.AddClass("primary")
	base.AddClasses("rounded", "padded")
	classes := base.StyleClasses()
	if len(classes) != 3 {
		t.Fatalf("expected 3 classes, got %d", len(classes))
	}
}
