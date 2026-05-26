package html

import (
	"fmt"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/style"
)

// Options configures HTML page generation.
type Options struct {
	Title string
}

// Generate renders a widget tree as a complete HTML page.
func Generate(root runtime.Widget, sheet *style.Stylesheet, opts Options) ([]byte, error) {
	hr, ok := root.(runtime.HTMLRenderer)
	if !ok {
		return nil, fmt.Errorf("root widget %T does not implement HTMLRenderer", root)
	}

	ctx := runtime.HTMLContext{Depth: 0}
	bodyHTML := string(hr.RenderHTML(ctx))
	css := StylesheetToCSS(sheet)
	js := interactivityJS

	return composePage(opts.Title, css, bodyHTML, js), nil
}
