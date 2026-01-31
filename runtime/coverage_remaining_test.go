package runtime

import (
	"math"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/animation"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/i18n"
	"github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/theme"
)

type renderChildWidget struct {
	bounds     Rect
	rendered   bool
	lastBounds Rect
}

func (w *renderChildWidget) Measure(Constraints) Size { return Size{} }
func (w *renderChildWidget) Layout(bounds Rect)       { w.bounds = bounds }
func (w *renderChildWidget) Render(ctx RenderContext) {
	w.rendered = true
	w.lastBounds = ctx.Bounds
}
func (w *renderChildWidget) HandleMessage(Message) HandleResult { return Unhandled() }
func (w *renderChildWidget) Bounds() Rect                       { return w.bounds }

type renderChildNoBounds struct {
	rendered   bool
	lastBounds Rect
}

func (w *renderChildNoBounds) Measure(Constraints) Size { return Size{} }
func (w *renderChildNoBounds) Layout(Rect)              {}
func (w *renderChildNoBounds) Render(ctx RenderContext) {
	w.rendered = true
	w.lastBounds = ctx.Bounds
}
func (w *renderChildNoBounds) HandleMessage(Message) HandleResult { return Unhandled() }

func TestRenderContextStyleHelpers(t *testing.T) {
	sheet := style.NewStylesheet().Add(style.Select("Root"), style.Style{})
	root := &styleTestWidget{typ: "Root"}
	resolver := newStyleResolver(sheet, []Widget{root}, style.MediaContext{})
	if resolver == nil {
		t.Fatalf("expected resolver")
	}

	ctx := RenderContext{
		Buffer:        NewBuffer(2, 1),
		Bounds:        Rect{X: 0, Y: 0, Width: 2, Height: 1},
		Focused:       true,
		styleResolver: resolver,
	}

	if got := ctx.ResolveStyle(root); !got.IsZero() {
		t.Fatalf("expected zero style, got %+v", got)
	}
	_ = ctx.ResolveBackendStyle(root)

	sub := ctx.WithBuffer(NewBuffer(1, 1), Rect{X: 0, Y: 0, Width: 1, Height: 1})
	if sub.Buffer == ctx.Buffer {
		t.Fatalf("expected new buffer")
	}
	if sub.Bounds.Width != 1 || sub.Bounds.Height != 1 {
		t.Fatalf("unexpected bounds: %+v", sub.Bounds)
	}
}

func TestRenderChild(t *testing.T) {
	ctx := RenderContext{Buffer: NewBuffer(6, 6), Bounds: Rect{X: 0, Y: 0, Width: 6, Height: 6}}
	child := &renderChildWidget{bounds: Rect{X: 1, Y: 1, Width: 2, Height: 2}}
	if !RenderChild(ctx, child) {
		t.Fatalf("expected child to render")
	}
	if !child.rendered {
		t.Fatalf("expected child render to be called")
	}
	if child.lastBounds != child.bounds {
		t.Fatalf("expected render bounds %+v, got %+v", child.bounds, child.lastBounds)
	}

	offscreen := &renderChildWidget{bounds: Rect{X: 10, Y: 10, Width: 1, Height: 1}}
	if RenderChild(ctx, offscreen) {
		t.Fatalf("expected offscreen child to skip rendering")
	}
	if offscreen.rendered {
		t.Fatalf("did not expect offscreen render")
	}

	plain := &renderChildNoBounds{}
	if !RenderChild(ctx, plain) {
		t.Fatalf("expected plain child to render")
	}
	if !plain.rendered {
		t.Fatalf("expected plain render to be called")
	}
	if plain.lastBounds != ctx.Bounds {
		t.Fatalf("expected context bounds for plain child")
	}
}

func TestRenderSamplerSummary(t *testing.T) {
	sampler := NewRenderSampler(0)
	if sampler.window != 120 {
		t.Fatalf("expected default window, got %d", sampler.window)
	}
	stats1 := RenderStats{
		TotalDuration:  10 * time.Millisecond,
		RenderDuration: 6 * time.Millisecond,
		FlushDuration:  4 * time.Millisecond,
		DirtyCells:     5,
		TotalCells:     10,
	}
	stats2 := RenderStats{
		TotalDuration:  20 * time.Millisecond,
		RenderDuration: 12 * time.Millisecond,
		FlushDuration:  8 * time.Millisecond,
		DirtyCells:     0,
		TotalCells:     0,
	}
	sampler.ObserveRender(stats1)
	sampler.ObserveRender(stats2)

	summary := sampler.Summary()
	if summary.Frames != 2 {
		t.Fatalf("expected 2 frames, got %d", summary.Frames)
	}
	if summary.Samples != 2 {
		t.Fatalf("expected 2 samples, got %d", summary.Samples)
	}
	if summary.Last.TotalDuration != stats2.TotalDuration {
		t.Fatalf("unexpected last sample")
	}
	if summary.MaxTotal != stats2.TotalDuration {
		t.Fatalf("unexpected max total duration")
	}
	if math.Abs(summary.AvgDirtyRatio-0.5) > 0.0001 {
		t.Fatalf("unexpected avg dirty ratio: %f", summary.AvgDirtyRatio)
	}
}

func TestRenderObserverFunc(t *testing.T) {
	var called bool
	var got RenderStats
	observer := RenderObserverFunc(func(stats RenderStats) {
		called = true
		got = stats
	})
	observer.ObserveRender(RenderStats{Frame: 7})
	if !called || got.Frame != 7 {
		t.Fatalf("expected observer to capture stats")
	}

	var nilObserver RenderObserverFunc
	nilObserver.ObserveRender(RenderStats{Frame: 1})
}

func TestSnapshotText(t *testing.T) {
	buf := NewBuffer(2, 2)
	buf.Set(0, 0, 'A', backend.DefaultStyle())
	buf.Set(0, 1, rune(0x80), backend.DefaultStyle())
	buf.Set(1, 1, 'B', backend.DefaultStyle())

	expected := "A \n" + string([]rune{rune(0x80), 'B'})
	if got := buf.SnapshotText(); got != expected {
		t.Fatalf("unexpected snapshot: %q", got)
	}

	app := NewApp(AppConfig{})
	app.screen = NewScreen(2, 1)
	app.screen.Buffer().SetString(0, 0, "Hi", backend.DefaultStyle())
	if got := app.SnapshotText(); got != "Hi" {
		t.Fatalf("unexpected app snapshot: %q", got)
	}
}

func TestSameRootsAndClear(t *testing.T) {
	root := &styleTestWidget{typ: "Root"}
	other := &styleTestWidget{typ: "Other"}
	if !sameRoots([]Widget{root}, []Widget{root}) {
		t.Fatalf("expected same roots")
	}
	if sameRoots([]Widget{root}, []Widget{other}) {
		t.Fatalf("expected different roots")
	}
	if sameRoots([]Widget{root}, []Widget{root, other}) {
		t.Fatalf("expected different root lengths")
	}

	buf := NewBuffer(2, 1)
	buf.Set(0, 0, 'X', backend.DefaultStyle())
	ctx := RenderContext{Buffer: buf, Bounds: Rect{X: 0, Y: 0, Width: 2, Height: 1}}
	ctx.Clear(backend.DefaultStyle())
	if got := buf.Get(0, 0).Rune; got != ' ' {
		t.Fatalf("expected cleared buffer, got %q", got)
	}
}

func TestServicesAccessorsAdditional(t *testing.T) {
	anim := animation.NewAnimator()
	localizer := i18n.MapLocalizer{LocaleCode: "en"}
	th := theme.DefaultTheme()

	app := NewApp(AppConfig{Theme: th, Localizer: localizer, Animator: anim})
	services := app.Services()
	if services.Theme() == nil {
		t.Fatalf("expected theme")
	}
	if services.Localizer() == nil {
		t.Fatalf("expected localizer")
	}
	if services.Animator() == nil {
		t.Fatalf("expected animator")
	}
	if services.Scheduler() == nil {
		t.Fatalf("expected scheduler")
	}
	if services.InvalidateScheduler() == nil {
		t.Fatalf("expected invalidation scheduler")
	}
}

func TestStyleNodeAccessors(t *testing.T) {
	node := styleNode{
		typ:     "panel",
		id:      "main",
		classes: []string{"primary", "rounded"},
		state:   style.WidgetState{Focused: true},
	}
	if node.StyleType() != "panel" {
		t.Fatalf("unexpected type")
	}
	if node.StyleID() != "main" {
		t.Fatalf("unexpected id")
	}
	if len(node.StyleClasses()) != 2 {
		t.Fatalf("unexpected classes")
	}
	if !node.StyleState().Focused {
		t.Fatalf("expected focused state")
	}
}

func TestSpacerAndPoolHelpers(t *testing.T) {
	spacer := NewSpacer()
	spacer.Layout(Rect{X: 1, Y: 2, Width: 3, Height: 4})
	if spacer.Bounds().Width != 3 {
		t.Fatalf("unexpected spacer bounds")
	}
	spacer.Render(RenderContext{})
	_ = Space()
	_ = FixedSpace(2)

	pool := NewWidgetPool(func() int { return 1 }, nil, 3)
	if pool.MaxSize() != 3 {
		t.Fatalf("unexpected pool max size")
	}
	var nilPool *WidgetPool[int]
	if nilPool.MaxSize() != 0 {
		t.Fatalf("expected nil pool max size to be 0")
	}
}

var _ Widget = (*renderChildWidget)(nil)
var _ BoundsProvider = (*renderChildWidget)(nil)
var _ Widget = (*renderChildNoBounds)(nil)
