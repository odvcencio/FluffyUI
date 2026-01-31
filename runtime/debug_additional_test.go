package runtime

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
)

type debugWidget struct {
	accessibility.Base
	id       string
	children []Widget
}

func (d *debugWidget) Measure(constraints Constraints) Size { return Size{} }

func (d *debugWidget) Layout(bounds Rect) {}

func (d *debugWidget) Render(ctx RenderContext) {}

func (d *debugWidget) HandleMessage(msg Message) HandleResult { return Unhandled() }

func (d *debugWidget) ChildWidgets() []Widget { return d.children }

func (d *debugWidget) ID() string { return d.id }

func (d *debugWidget) PathSegment(child Widget) string { return "Segment" }

func TestErrorReporterFormats(t *testing.T) {
	child := &debugWidget{Base: accessibility.Base{Label: "Child"}, id: "child"}
	root := &debugWidget{Base: accessibility.Base{Label: "Root"}, id: "root", children: []Widget{child}}
	buf := &bytes.Buffer{}
	reporter := &ErrorReporter{
		ShowStackTrace: true,
		ShowWidgetTree: true,
		Writer:         buf,
		RootProvider: func() Widget {
			return root
		},
	}

	reporter.ReportWidgetError(child, errors.New("boom"), KeyMsg{Key: 0})
	out := buf.String()
	if !strings.Contains(out, "Widget Error") {
		t.Fatalf("expected formatted error output, got: %s", out)
	}
	if !strings.Contains(out, "Widget Tree") {
		t.Fatalf("expected widget tree output")
	}
}

var _ Widget = (*debugWidget)(nil)
var _ ChildProvider = (*debugWidget)(nil)
var _ PathSegmenter = (*debugWidget)(nil)
