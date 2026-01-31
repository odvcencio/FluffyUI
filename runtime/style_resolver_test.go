package runtime

import (
	"testing"

	"github.com/odvcencio/fluffyui/style"
)

type styleTestWidget struct {
	typ      string
	id       string
	classes  []string
	state    style.WidgetState
	children []Widget
	bounds   Rect
	focused  bool
}

func (w *styleTestWidget) Measure(constraints Constraints) Size {
	return Size{Width: 1, Height: 1}
}

func (w *styleTestWidget) Layout(bounds Rect) {
	w.bounds = bounds
	for _, child := range w.children {
		if child != nil {
			child.Layout(bounds)
		}
	}
}

func (w *styleTestWidget) Render(ctx RenderContext) {}

func (w *styleTestWidget) HandleMessage(msg Message) HandleResult { return Unhandled() }

func (w *styleTestWidget) StyleType() string { return w.typ }

func (w *styleTestWidget) StyleID() string { return w.id }

func (w *styleTestWidget) StyleClasses() []string { return w.classes }

func (w *styleTestWidget) StyleState() style.WidgetState { return w.state }

func (w *styleTestWidget) ChildWidgets() []Widget { return w.children }

func (w *styleTestWidget) Bounds() Rect { return w.bounds }

func (w *styleTestWidget) CanFocus() bool { return true }

func (w *styleTestWidget) Focus() { w.focused = true }

func (w *styleTestWidget) Blur() { w.focused = false }

func (w *styleTestWidget) IsFocused() bool { return w.focused }

func TestStyleResolverResolve(t *testing.T) {
	sheet := style.NewStylesheet().Add(style.Select("Root"), style.Style{})
	root := &styleTestWidget{typ: "Root", id: "root", classes: []string{"a"}, state: style.WidgetState{Disabled: true}}
	child := &styleTestWidget{typ: "Child", id: "child", classes: []string{"b"}}
	root.children = []Widget{child}

	resolver := newStyleResolver(sheet, []Widget{root}, style.MediaContext{Width: 10, Height: 5})
	if resolver == nil {
		t.Fatalf("expected resolver")
	}
	_ = resolver.Resolve(root, true)
	_ = resolver.Resolve(child, false)
	resolver.ResetCache()

	_ = buildParentMap([]Widget{root})
	_ = normalizeClasses([]string{" a ", "", "b"})
}

var _ Widget = (*styleTestWidget)(nil)
var _ ChildProvider = (*styleTestWidget)(nil)
var _ BoundsProvider = (*styleTestWidget)(nil)
var _ Focusable = (*styleTestWidget)(nil)
var _ StyleTypeProvider = (*styleTestWidget)(nil)
var _ StyleIDProvider = (*styleTestWidget)(nil)
var _ StyleClassProvider = (*styleTestWidget)(nil)
var _ StyleStateProvider = (*styleTestWidget)(nil)
