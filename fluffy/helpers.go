package fluffy

import (
	"strings"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/widgets"
)

// Spacer returns a flexible spacer widget that fills available space.
// Use this inside VStack/HStack to push siblings apart.
//
// Example:
//
//	fluffy.VStack(header, fluffy.Spacer(), footer)
func Spacer() runtime.Widget {
	w := widgets.NewSimpleWidget()
	w.MeasureFunc = func(c runtime.Constraints) runtime.Size {
		return c.MaxSize()
	}
	return w
}

// FixedSpacer returns a fixed-size spacer widget.
// The spacer occupies the given number of cells in both width and height.
//
// Example:
//
//	fluffy.VStack(title, fluffy.FixedSpacer(2), body)
func FixedSpacer(size int) runtime.Widget {
	w := widgets.NewSimpleWidget()
	w.MeasureFunc = func(c runtime.Constraints) runtime.Size {
		return c.Constrain(runtime.Size{Width: size, Height: size})
	}
	return w
}

// Divider returns a horizontal divider line that fills available width.
//
// Example:
//
//	fluffy.VStack(section1, fluffy.Divider(), section2)
func Divider() runtime.Widget {
	w := widgets.NewSimpleWidget()
	w.MeasureFunc = func(c runtime.Constraints) runtime.Size {
		return c.Constrain(runtime.Size{Width: c.MaxWidth, Height: 1})
	}
	w.RenderFunc = func(ctx runtime.RenderContext) {
		b := w.ContentBounds()
		if b.Width <= 0 || b.Height <= 0 {
			return
		}
		line := strings.Repeat("─", b.Width)
		ctx.Buffer.SetString(b.X, b.Y, line, backend.DefaultStyle().Dim(true))
	}
	w.HTMLRenderFunc = func(ctx runtime.HTMLContext) runtime.HTML {
		return `<hr class="fluffy-Divider">`
	}
	return w
}

// VDivider returns a vertical divider line that fills available height.
//
// Example:
//
//	fluffy.HStack(left, fluffy.VDivider(), right)
func VDivider() runtime.Widget {
	w := widgets.NewSimpleWidget()
	w.MeasureFunc = func(c runtime.Constraints) runtime.Size {
		return c.Constrain(runtime.Size{Width: 1, Height: c.MaxHeight})
	}
	w.RenderFunc = func(ctx runtime.RenderContext) {
		b := w.ContentBounds()
		if b.Width <= 0 || b.Height <= 0 {
			return
		}
		s := backend.DefaultStyle().Dim(true)
		for y := 0; y < b.Height; y++ {
			ctx.Buffer.SetString(b.X, b.Y+y, "│", s)
		}
	}
	w.HTMLRenderFunc = func(ctx runtime.HTMLContext) runtime.HTML {
		return `<div class="fluffy-VDivider"></div>`
	}
	return w
}

// Padding wraps a child widget with uniform padding on all sides.
//
// Example:
//
//	fluffy.Padding(2, fluffy.Label("Hello"))
func Padding(padding int, child runtime.Widget) runtime.Widget {
	box := widgets.NewBox(child)
	box.ApplyStyle(style.Style{Padding: style.Pad(padding)})
	return box
}

// PaddingXY wraps a child widget with horizontal (px) and vertical (py) padding.
//
// Example:
//
//	fluffy.PaddingXY(4, 1, fluffy.Label("Centered"))
func PaddingXY(px, py int, child runtime.Widget) runtime.Widget {
	box := widgets.NewBox(child)
	box.ApplyStyle(style.Style{Padding: style.PadXY(px, py)})
	return box
}

// Center wraps a child widget centered in the available space.
// It uses a flex layout with spacers on both axes.
//
// Example:
//
//	fluffy.Center(fluffy.Label("Centered"))
func Center(child runtime.Widget) runtime.Widget {
	return runtime.VBox(
		runtime.Space(),
		runtime.Fixed(runtime.HBox(
			runtime.Space(),
			runtime.Fixed(child),
			runtime.Space(),
		)),
		runtime.Space(),
	)
}

// Bold creates a bold text widget.
//
// Example:
//
//	fluffy.Bold("Important")
func Bold(text string) runtime.Widget {
	return widgets.NewText(text, widgets.WithTextStyle(backend.DefaultStyle().Bold(true)))
}

// Italic creates an italic text widget.
//
// Example:
//
//	fluffy.Italic("Emphasis")
func Italic(text string) runtime.Widget {
	return widgets.NewText(text, widgets.WithTextStyle(backend.DefaultStyle().Italic(true)))
}

// DangerText creates a red-colored text widget.
//
// Example:
//
//	fluffy.DangerText("Error: something went wrong")
func DangerText(text string) runtime.Widget {
	return widgets.NewText(text, widgets.WithTextStyle(backend.DefaultStyle().Foreground(backend.ColorRed)))
}

// SuccessText creates a green-colored text widget.
//
// Example:
//
//	fluffy.SuccessText("Operation completed")
func SuccessText(text string) runtime.Widget {
	return widgets.NewText(text, widgets.WithTextStyle(backend.DefaultStyle().Foreground(backend.ColorGreen)))
}

// WarningText creates a yellow-colored text widget.
//
// Example:
//
//	fluffy.WarningText("Disk usage above 90%")
func WarningText(text string) runtime.Widget {
	return widgets.NewText(text, widgets.WithTextStyle(backend.DefaultStyle().Foreground(backend.ColorYellow)))
}

// Group wraps children in a bordered panel with a title.
// This is a shorthand for creating a Panel with a VStack of children.
//
// Example:
//
//	fluffy.Group("Settings",
//	    fluffy.Label("Theme: dark"),
//	    fluffy.Label("Font size: 14"),
//	)
func Group(title string, children ...runtime.Widget) runtime.Widget {
	panel := widgets.NewPanel(VStack(children...), widgets.WithPanelTitle(title))
	panel.SetBorder(true)
	return panel
}

// Conditional returns the widget if show is true, or nil otherwise.
// Nil children are safely skipped by VStack, HStack, and other containers.
//
// Example:
//
//	fluffy.VStack(
//	    fluffy.Label("Always visible"),
//	    fluffy.Conditional(showDetails, fluffy.Label("Details...")),
//	)
func Conditional(show bool, widget runtime.Widget) runtime.Widget {
	if show {
		return widget
	}
	return nil
}

// onMounter is implemented by widgets that support an OnMount lifecycle hook.
type onMounter interface {
	OnMount(fn func())
}

// onUnmounter is implemented by widgets that support an OnUnmount lifecycle hook.
type onUnmounter interface {
	OnUnmount(fn func())
}

// WithMount sets an onMount callback on a widget and returns it for chaining.
// The callback fires when the widget is bound (mounted into a screen).
// If the widget does not support OnMount, this is a no-op.
//
// Example:
//
//	label := fluffy.WithMount(fluffy.Label("Hello"), func() {
//	    log.Println("label mounted")
//	})
func WithMount(w runtime.Widget, fn func()) runtime.Widget {
	if m, ok := w.(onMounter); ok {
		m.OnMount(fn)
	}
	return w
}

// WithUnmount sets an onUnmount callback on a widget and returns it for chaining.
// The callback fires when the widget is unbound (removed from a screen).
// If the widget does not support OnUnmount, this is a no-op.
//
// Example:
//
//	label := fluffy.WithUnmount(fluffy.Label("Hello"), func() {
//	    log.Println("label unmounted")
//	})
func WithUnmount(w runtime.Widget, fn func()) runtime.Widget {
	if m, ok := w.(onUnmounter); ok {
		m.OnUnmount(fn)
	}
	return w
}
