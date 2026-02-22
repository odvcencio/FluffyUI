package html

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/style"
)

type testWidget struct {
	html string
}

func (w *testWidget) Measure(c runtime.Constraints) runtime.Size          { return runtime.Size{} }
func (w *testWidget) Layout(b runtime.Rect)                               {}
func (w *testWidget) Render(ctx runtime.RenderContext)                     {}
func (w *testWidget) HandleMessage(m runtime.Message) runtime.HandleResult { return runtime.HandleResult{} }
func (w *testWidget) RenderHTML(ctx runtime.HTMLContext) runtime.HTML      { return runtime.HTML(w.html) }

func TestGenerate_Basic(t *testing.T) {
	root := &testWidget{html: `<div class="fluffy-Label">Hello</div>`}
	sheet := style.NewStylesheet()

	page, err := Generate(root, sheet, Options{Title: "Test Page"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	html := string(page)
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("missing doctype")
	}
	if !strings.Contains(html, "<title>Test Page</title>") {
		t.Error("missing title")
	}
	if !strings.Contains(html, `class="fluffy-Label"`) {
		t.Error("missing widget HTML")
	}
	if !strings.Contains(html, "<style>") {
		t.Error("missing style tag")
	}
	if !strings.Contains(html, "<script>") {
		t.Error("missing script tag")
	}
}

func TestGenerate_NilSheet(t *testing.T) {
	root := &testWidget{html: "<span>hi</span>"}
	page, err := Generate(root, nil, Options{Title: "No Sheet"})
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if !strings.Contains(string(page), "<span>hi</span>") {
		t.Error("missing content")
	}
}

type nonHTMLWidget struct{}

func (w *nonHTMLWidget) Measure(c runtime.Constraints) runtime.Size          { return runtime.Size{} }
func (w *nonHTMLWidget) Layout(b runtime.Rect)                               {}
func (w *nonHTMLWidget) Render(ctx runtime.RenderContext)                     {}
func (w *nonHTMLWidget) HandleMessage(m runtime.Message) runtime.HandleResult { return runtime.HandleResult{} }

func TestGenerate_NonHTMLRenderer(t *testing.T) {
	root := &nonHTMLWidget{}
	_, err := Generate(root, nil, Options{Title: "Test"})
	if err == nil {
		t.Error("expected error for non-HTMLRenderer root")
	}
}
