// runtime/html_test.go
package runtime

import (
	"html/template"
	"strings"
	"testing"
)

type mockHTMLWidget struct{ html string }

func (m *mockHTMLWidget) RenderHTML(ctx HTMLContext) HTML {
	return HTML(m.html)
}
func (m *mockHTMLWidget) Measure(c Constraints) Size            { return Size{} }
func (m *mockHTMLWidget) Layout(b Rect)                         {}
func (m *mockHTMLWidget) Render(ctx RenderContext)               {}
func (m *mockHTMLWidget) HandleMessage(msg Message) HandleResult { return HandleResult{} }

func TestHTMLContext_Child(t *testing.T) {
	ctx := HTMLContext{Depth: 0}
	child := ctx.Child()
	if child.Depth != 1 {
		t.Errorf("Child().Depth = %d, want 1", child.Depth)
	}
}

func TestHTMLRenderer_Interface(t *testing.T) {
	w := &mockHTMLWidget{html: "<span>hello</span>"}
	var r HTMLRenderer = w
	got := r.RenderHTML(HTMLContext{})
	if got != template.HTML("<span>hello</span>") {
		t.Errorf("RenderHTML = %q, want <span>hello</span>", got)
	}
}

func TestFlex_RenderHTML_Column(t *testing.T) {
	child := &mockHTMLWidget{html: "<span>child</span>"}
	f := VBox(Fixed(child))
	got := string(f.RenderHTML(HTMLContext{}))
	if !strings.Contains(got, "flex-direction:column") {
		t.Errorf("missing column direction: %s", got)
	}
	if !strings.Contains(got, "<span>child</span>") {
		t.Errorf("missing child content: %s", got)
	}
}

func TestFlex_RenderHTML_Row(t *testing.T) {
	child := &mockHTMLWidget{html: "<span>item</span>"}
	f := HBox(Fixed(child))
	got := string(f.RenderHTML(HTMLContext{}))
	if !strings.Contains(got, "flex-direction:row") {
		t.Errorf("missing row direction: %s", got)
	}
}
