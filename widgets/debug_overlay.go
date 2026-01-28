package widgets

import (
	"fmt"
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
)

// DebugOverlay draws widget bounds and labels for layout debugging.
type DebugOverlay struct {
	Base
	root       runtime.Widget
	style      backend.Style
	labelStyle backend.Style
	showLabels bool
	maxDepth   int
}

// DebugOverlayOption configures a debug overlay.
type DebugOverlayOption func(*DebugOverlay)

// NewDebugOverlay creates a debug overlay for the provided root widget.
func NewDebugOverlay(root runtime.Widget, opts ...DebugOverlayOption) *DebugOverlay {
	overlay := &DebugOverlay{
		root:       root,
		style:      backend.DefaultStyle().Dim(true),
		labelStyle: backend.DefaultStyle().Dim(true),
		showLabels: true,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(overlay)
		}
	}
	return overlay
}

// WithDebugStyle sets the box style.
func WithDebugStyle(style backend.Style) DebugOverlayOption {
	return func(d *DebugOverlay) {
		if d == nil {
			return
		}
		d.style = style
	}
}

// WithDebugLabelStyle sets the label style.
func WithDebugLabelStyle(style backend.Style) DebugOverlayOption {
	return func(d *DebugOverlay) {
		if d == nil {
			return
		}
		d.labelStyle = style
	}
}

// WithDebugLabels toggles label rendering.
func WithDebugLabels(enabled bool) DebugOverlayOption {
	return func(d *DebugOverlay) {
		if d == nil {
			return
		}
		d.showLabels = enabled
	}
}

// WithDebugMaxDepth limits traversal depth (0 = unlimited).
func WithDebugMaxDepth(depth int) DebugOverlayOption {
	return func(d *DebugOverlay) {
		if d == nil {
			return
		}
		d.maxDepth = depth
	}
}

// Measure takes all available space.
func (d *DebugOverlay) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.MaxSize()
}

// Layout stores bounds.
func (d *DebugOverlay) Layout(bounds runtime.Rect) {
	d.Base.Layout(bounds)
}

// Render draws bounding boxes for the widget tree.
func (d *DebugOverlay) Render(ctx runtime.RenderContext) {
	if d == nil || d.root == nil {
		return
	}
	visited := map[runtime.Widget]struct{}{}
	var walk func(node runtime.Widget, depth int)
	walk = func(node runtime.Widget, depth int) {
		if node == nil {
			return
		}
		if _, ok := visited[node]; ok {
			return
		}
		visited[node] = struct{}{}
		if d.maxDepth > 0 && depth > d.maxDepth {
			return
		}
		if node != d {
			if bp, ok := node.(runtime.BoundsProvider); ok {
				bounds := bp.Bounds()
				if bounds.Width > 0 && bounds.Height > 0 {
					if bounds.Width >= 2 && bounds.Height >= 2 {
						ctx.Buffer.DrawBox(bounds, d.style)
					} else {
						ctx.Buffer.Set(bounds.X, bounds.Y, '·', d.style)
					}
					if d.showLabels {
						label := debugLabel(node)
						if label != "" && bounds.Width > 2 {
							ctx.Buffer.SetString(bounds.X+1, bounds.Y, clipString(label, bounds.Width-2), d.labelStyle)
						}
					}
				}
			}
		}
		if container, ok := node.(runtime.ChildProvider); ok {
			for _, child := range container.ChildWidgets() {
				walk(child, depth+1)
			}
		}
	}
	walk(d.root, 0)
}

func debugLabel(widget runtime.Widget) string {
	name := fmt.Sprintf("%T", widget)
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	name = strings.TrimPrefix(name, "*")
	if accessible, ok := widget.(accessibility.Accessible); ok && accessible != nil {
		label := strings.TrimSpace(accessible.AccessibleLabel())
		if label != "" {
			return name + " (" + label + ")"
		}
	}
	if ider, ok := widget.(interface{ ID() string }); ok {
		id := strings.TrimSpace(ider.ID())
		if id != "" {
			return name + " #" + id
		}
	}
	return name
}

var _ runtime.Widget = (*DebugOverlay)(nil)
