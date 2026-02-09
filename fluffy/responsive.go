package fluffy

import (
	"sort"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/widgets"
)

// Breakpoint defines a width threshold and its associated layout.
type Breakpoint struct {
	MinWidth int            // minimum terminal width for this breakpoint
	Widget   runtime.Widget // the widget to show at this breakpoint
}

// Responsive returns a widget that switches layout based on terminal width.
// Breakpoints are evaluated from largest MinWidth to smallest.
// The first breakpoint whose MinWidth <= current terminal width wins.
// If no breakpoint matches, the last one (smallest MinWidth) is used as fallback.
//
// Example:
//
//	fluffy.Responsive(
//	    fluffy.Breakpoint{MinWidth: 120, Widget: threeColumnLayout},
//	    fluffy.Breakpoint{MinWidth: 80, Widget: twoColumnLayout},
//	    fluffy.Breakpoint{MinWidth: 0, Widget: singleColumnLayout},
//	)
func Responsive(breakpoints ...Breakpoint) runtime.Widget {
	if len(breakpoints) == 0 {
		return widgets.NewSimpleWidget()
	}

	// Sort breakpoints from largest MinWidth to smallest so we can iterate
	// and pick the first one whose MinWidth <= available width.
	sorted := make([]Breakpoint, len(breakpoints))
	copy(sorted, breakpoints)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].MinWidth > sorted[j].MinWidth
	})

	w := widgets.NewSimpleWidget()

	// Cache the index of the last-selected breakpoint to avoid redundant
	// measure/layout/render of the wrong child.
	activeIndex := -1

	pickBreakpoint := func(width int) int {
		for i, bp := range sorted {
			if width >= bp.MinWidth {
				return i
			}
		}
		// Fallback: use the smallest breakpoint (last in sorted order).
		return len(sorted) - 1
	}

	w.MeasureFunc = func(c runtime.Constraints) runtime.Size {
		idx := pickBreakpoint(c.MaxWidth)
		activeIndex = idx
		child := sorted[idx].Widget
		if child == nil {
			return c.MinSize()
		}
		return child.Measure(c)
	}

	w.LayoutFunc = func(bounds runtime.Rect) {
		if activeIndex < 0 || activeIndex >= len(sorted) {
			return
		}
		child := sorted[activeIndex].Widget
		if child == nil {
			return
		}
		child.Layout(bounds)
	}

	w.RenderFunc = func(ctx runtime.RenderContext) {
		if activeIndex < 0 || activeIndex >= len(sorted) {
			return
		}
		child := sorted[activeIndex].Widget
		if child == nil {
			return
		}
		child.Render(ctx)
	}

	w.HandleMessageFunc = func(msg runtime.Message) runtime.HandleResult {
		if activeIndex < 0 || activeIndex >= len(sorted) {
			return runtime.Unhandled()
		}
		child := sorted[activeIndex].Widget
		if child == nil {
			return runtime.Unhandled()
		}
		return child.HandleMessage(msg)
	}

	return w
}

// IfWide returns wide when terminal width >= threshold, otherwise narrow.
// This is a convenience wrapper around Responsive for the common two-layout case.
//
// Example:
//
//	fluffy.IfWide(80,
//	    fluffy.HStack(sidebar, content),
//	    fluffy.VStack(sidebar, content),
//	)
func IfWide(threshold int, wide, narrow runtime.Widget) runtime.Widget {
	return Responsive(
		Breakpoint{MinWidth: threshold, Widget: wide},
		Breakpoint{MinWidth: 0, Widget: narrow},
	)
}

// AdaptiveStack returns an HStack when terminal width >= threshold, otherwise
// a VStack. This is the most common responsive pattern -- horizontal layout
// on wide screens, vertical on narrow ones.
//
// Example:
//
//	fluffy.AdaptiveStack(80, sidebar, content, aside)
func AdaptiveStack(threshold int, children ...runtime.Widget) runtime.Widget {
	return IfWide(threshold, HStack(children...), VStack(children...))
}
