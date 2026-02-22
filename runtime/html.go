// runtime/html.go
package runtime

import "html/template"

// HTML is an alias for template.HTML so widget packages need not import html/template.
type HTML = template.HTML

// HTMLContext carries state for HTML rendering.
type HTMLContext struct {
	Depth int // nesting depth for optional indentation
}

// Child returns a context for rendering child widgets.
func (ctx HTMLContext) Child() HTMLContext {
	return HTMLContext{Depth: ctx.Depth + 1}
}

// HTMLRenderer is implemented by widgets that can render to static HTML.
type HTMLRenderer interface {
	RenderHTML(ctx HTMLContext) HTML
}
