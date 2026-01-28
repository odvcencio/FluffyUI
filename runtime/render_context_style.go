package runtime

import (
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/style"
)

// ResolveStyle returns the resolved stylesheet style for the widget.
func (ctx RenderContext) ResolveStyle(widget Widget) style.Style {
	if ctx.styleResolver == nil || widget == nil {
		return style.Style{}
	}
	return ctx.styleResolver.Resolve(widget, ctx.Focused)
}

// ResolveBackendStyle returns a backend style for the widget.
func (ctx RenderContext) ResolveBackendStyle(widget Widget) backend.Style {
	return ctx.ResolveStyle(widget).ToBackend()
}

// WithBuffer returns a new context that renders into the provided buffer.
func (ctx RenderContext) WithBuffer(buffer *Buffer, bounds Rect) RenderContext {
	return RenderContext{
		Buffer:        buffer,
		Focused:       ctx.Focused,
		Bounds:        bounds,
		styleResolver: ctx.styleResolver,
	}
}
