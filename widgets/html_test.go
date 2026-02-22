package widgets

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
)

func TestLabel_RenderHTML(t *testing.T) {
	l := NewLabel("Hello World")
	got := string(l.RenderHTML(runtime.HTMLContext{}))
	if !strings.Contains(got, `class="fluffy-Label"`) {
		t.Errorf("missing class: %s", got)
	}
	if !strings.Contains(got, "Hello World") {
		t.Errorf("missing text: %s", got)
	}
}

func TestLabel_RenderHTML_Escapes(t *testing.T) {
	l := NewLabel("<script>alert('xss')</script>")
	got := string(l.RenderHTML(runtime.HTMLContext{}))
	if strings.Contains(got, "<script>") {
		t.Errorf("XSS not escaped: %s", got)
	}
}

func TestLink_RenderHTML(t *testing.T) {
	l := NewLink("FluffyUI", "https://github.com/odvcencio/FluffyUI")
	got := string(l.RenderHTML(runtime.HTMLContext{}))
	if !strings.Contains(got, `class="fluffy-Link"`) {
		t.Errorf("missing class: %s", got)
	}
	if !strings.Contains(got, `href="https://github.com/odvcencio/FluffyUI"`) {
		t.Errorf("missing href: %s", got)
	}
	if !strings.Contains(got, `target="_blank"`) {
		t.Errorf("missing target: %s", got)
	}
	if !strings.Contains(got, "FluffyUI") {
		t.Errorf("missing label: %s", got)
	}
}

func TestSimpleWidget_RenderHTML_Nil(t *testing.T) {
	w := NewSimpleWidget()
	got := w.RenderHTML(runtime.HTMLContext{})
	if got != "" {
		t.Errorf("expected empty, got %s", got)
	}
}

func TestSimpleWidget_RenderHTML_Custom(t *testing.T) {
	w := NewSimpleWidget()
	w.HTMLRenderFunc = func(ctx runtime.HTMLContext) runtime.HTML {
		return `<hr class="fluffy-Divider">`
	}
	got := string(w.RenderHTML(runtime.HTMLContext{}))
	if got != `<hr class="fluffy-Divider">` {
		t.Errorf("got %s", got)
	}
}
