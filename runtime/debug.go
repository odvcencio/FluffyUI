package runtime

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
)

// ErrorReporter formats widget errors with context.
type ErrorReporter struct {
	ShowStackTrace bool
	ShowWidgetTree bool
	Writer         io.Writer
	RootProvider   func() Widget
}

// ReportWidgetError outputs a formatted error report for a widget.
func (er *ErrorReporter) ReportWidgetError(widget Widget, err error, msg Message) {
	if er == nil {
		return
	}
	if err == nil {
		err = errors.New("unknown error")
	}
	report := er.format(widget, err, msg)
	out := er.Writer
	if out == nil {
		out = os.Stderr
	}
	fmt.Fprintln(out, report)
}

func (er *ErrorReporter) format(widget Widget, err error, msg Message) string {
	name := widgetDisplayName(widget)
	lines := []string{
		"Widget Error",
		fmt.Sprintf("In %s", name),
		"",
		err.Error(),
	}

	if root := er.rootForTree(); root != nil {
		if path := buildWidgetPath(root, widget); len(path) > 0 {
			lines = append(lines, "")
			lines = append(lines, "Widget Path: "+strings.Join(path, " > "))
		}
	}

	if er.ShowWidgetTree {
		if root := er.rootForTree(); root != nil {
			lines = append(lines, "")
			lines = append(lines, "Widget Tree:")
			lines = append(lines, buildWidgetTree(root, widget)...)
		}
	}

	if msg != nil {
		lines = append(lines, "")
		lines = append(lines, "Last Message: "+formatMessage(msg))
	}

	if er.ShowStackTrace {
		if pe, ok := err.(panicError); ok && len(pe.stack) > 0 {
			lines = append(lines, "")
			lines = append(lines, "Stack Trace:")
			stackLines := strings.Split(strings.TrimSpace(string(pe.stack)), "\n")
			lines = append(lines, stackLines...)
		}
	}

	return formatBox(lines)
}

func (er *ErrorReporter) rootForTree() Widget {
	if er.RootProvider == nil {
		return nil
	}
	return er.RootProvider()
}

type panicError struct {
	value any
	stack []byte
}

func (p panicError) Error() string {
	return fmt.Sprintf("panic: %v", p.value)
}

func newPanicError(value any) error {
	return panicError{
		value: value,
		stack: debug.Stack(),
	}
}

type widgetTreeRoot struct {
	roots []Widget
}

func (w *widgetTreeRoot) Measure(constraints Constraints) Size {
	return Size{}
}

func (w *widgetTreeRoot) Layout(bounds Rect) {}

func (w *widgetTreeRoot) Render(ctx RenderContext) {}

func (w *widgetTreeRoot) HandleMessage(msg Message) HandleResult {
	return Unhandled()
}

func (w *widgetTreeRoot) ChildWidgets() []Widget {
	if w == nil {
		return nil
	}
	return w.roots
}

func widgetDisplayName(widget Widget) string {
	if widget == nil {
		return "Unknown"
	}
	if _, ok := widget.(*widgetTreeRoot); ok {
		return "App"
	}
	name := typeName(widget)
	label := ""
	if accessible, ok := widget.(accessibility.Accessible); ok && accessible != nil {
		label = strings.TrimSpace(accessible.AccessibleLabel())
	}
	if label != "" {
		name += " (" + label + ")"
	}
	if ider, ok := widget.(interface{ ID() string }); ok {
		id := strings.TrimSpace(ider.ID())
		if id != "" {
			name += " [id:" + id + "]"
		}
	}
	return name
}

func typeName(value any) string {
	if value == nil {
		return "Unknown"
	}
	name := fmt.Sprintf("%T", value)
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	name = strings.TrimPrefix(name, "*")
	return name
}

func formatMessage(msg Message) string {
	if msg == nil {
		return "none"
	}
	return fmt.Sprintf("%s%+v", typeName(msg), msg)
}

func buildWidgetTree(root Widget, target Widget) []string {
	if root == nil {
		return nil
	}
	var lines []string
	var walk func(node Widget, prefix string, isLast bool)
	walk = func(node Widget, prefix string, isLast bool) {
		if node == nil {
			return
		}
		name := widgetDisplayName(node)
		if node == target {
			name += " [HERE]"
		}
		line := name
		if prefix != "" {
			branch := "|- "
			if isLast {
				branch = "`- "
			}
			line = prefix + branch + name
		}
		lines = append(lines, line)

		childrenProvider, ok := node.(ChildProvider)
		if !ok {
			return
		}
		var children []Widget
		for _, child := range childrenProvider.ChildWidgets() {
			if child != nil {
				children = append(children, child)
			}
		}
		for i, child := range children {
			nextPrefix := prefix
			if prefix != "" {
				if isLast {
					nextPrefix += "   "
				} else {
					nextPrefix += "|  "
				}
			} else {
				nextPrefix = ""
			}
			walk(child, nextPrefix, i == len(children)-1)
		}
	}
	walk(root, "", true)
	return lines
}

func buildWidgetPath(root Widget, target Widget) []string {
	if root == nil || target == nil {
		return nil
	}
	var walk func(node Widget) ([]string, bool)
	walk = func(node Widget) ([]string, bool) {
		if node == nil {
			return nil, false
		}
		if node == target {
			return []string{widgetDisplayName(node)}, true
		}
		container, ok := node.(ChildProvider)
		if !ok {
			return nil, false
		}
		for _, child := range container.ChildWidgets() {
			if child == nil {
				continue
			}
			if path, ok := walk(child); ok {
				segment := widgetDisplayName(node)
				if segmenter, ok := node.(PathSegmenter); ok {
					if seg := strings.TrimSpace(segmenter.PathSegment(child)); seg != "" {
						segment = seg
					}
				}
				return append([]string{segment}, path...), true
			}
		}
		return nil, false
	}
	path, _ := walk(root)
	return path
}

func formatBox(lines []string) string {
	width := 0
	for _, line := range lines {
		if len(line) > width {
			width = len(line)
		}
	}
	border := "+" + strings.Repeat("-", width+2) + "+"
	var sb strings.Builder
	sb.WriteString(border)
	sb.WriteString("\n")
	for _, line := range lines {
		sb.WriteString("| ")
		sb.WriteString(padRightASCII(line, width))
		sb.WriteString(" |")
		sb.WriteString("\n")
	}
	sb.WriteString(border)
	return sb.String()
}

func padRightASCII(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

var _ Widget = (*widgetTreeRoot)(nil)
var _ ChildProvider = (*widgetTreeRoot)(nil)
