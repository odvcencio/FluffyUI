package widgets

import (
	"fmt"
	"strings"
	"testing"

	"m31labs.dev/fluffyui/runtime"
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

func TestChip_RenderHTML(t *testing.T) {
	c := NewChip("golang")
	got := string(c.RenderHTML(runtime.HTMLContext{}))
	if !strings.Contains(got, `class="fluffy-Chip"`) {
		t.Errorf("missing class: %s", got)
	}
	if !strings.Contains(got, "golang") {
		t.Errorf("missing label: %s", got)
	}
	if !strings.HasPrefix(got, "<button") {
		t.Errorf("expected button element: %s", got)
	}
}

func TestSearchWidget_RenderHTML(t *testing.T) {
	s := NewSearchWidget()
	got := string(s.RenderHTML(runtime.HTMLContext{}))
	if !strings.Contains(got, `type="search"`) {
		t.Errorf("missing type=search: %s", got)
	}
	if !strings.Contains(got, `class="fluffy-Search"`) {
		t.Errorf("missing class: %s", got)
	}
}

func TestMarkdownViewer_RenderHTML(t *testing.T) {
	m := NewMarkdownViewer("# Hello\n\nSome **bold** text.")
	got := string(m.RenderHTML(runtime.HTMLContext{}))
	if !strings.Contains(got, `class="fluffy-MarkdownViewer"`) {
		t.Errorf("missing class: %s", got)
	}
	if !strings.Contains(got, "<h1>Hello</h1>") {
		t.Errorf("missing h1: %s", got)
	}
	if !strings.Contains(got, "<strong>bold</strong>") {
		t.Errorf("missing bold: %s", got)
	}
}

func TestList_RenderHTML_Default(t *testing.T) {
	adapter := NewSliceAdapter([]string{"Alpha", "Beta"}, nil)
	l := NewList[string](adapter)
	got := string(l.RenderHTML(runtime.HTMLContext{}))
	if !strings.Contains(got, `class="fluffy-List"`) {
		t.Errorf("missing class: %s", got)
	}
	if !strings.Contains(got, "<li>Alpha</li>") {
		t.Errorf("missing Alpha: %s", got)
	}
	if !strings.Contains(got, "<li>Beta</li>") {
		t.Errorf("missing Beta: %s", got)
	}
}

func TestList_RenderHTML_CustomRenderer(t *testing.T) {
	adapter := NewSliceAdapter([]string{"one", "two"}, nil)
	l := NewList[string](adapter)
	l.SetHTMLItemRenderer(func(item string, index int) runtime.HTML {
		return runtime.HTML(fmt.Sprintf(`<li data-idx="%d">%s</li>`, index, item))
	})
	got := string(l.RenderHTML(runtime.HTMLContext{}))
	if !strings.Contains(got, `data-idx="0"`) {
		t.Errorf("missing custom attr: %s", got)
	}
}

func TestSplitter_RenderHTML(t *testing.T) {
	first := NewLabel("sidebar")
	second := NewLabel("content")
	s := NewSplitter(first, second)
	s.Orientation = SplitHorizontal
	s.Ratio = 0.25
	got := string(s.RenderHTML(runtime.HTMLContext{}))
	if !strings.Contains(got, `class="fluffy-Splitter"`) {
		t.Errorf("missing class: %s", got)
	}
	if !strings.Contains(got, "flex-direction:row") {
		t.Errorf("missing row direction: %s", got)
	}
	if !strings.Contains(got, "25%") {
		t.Errorf("missing ratio: %s", got)
	}
	if !strings.Contains(got, "sidebar") {
		t.Errorf("missing first pane: %s", got)
	}
	if !strings.Contains(got, "content") {
		t.Errorf("missing second pane: %s", got)
	}
}
