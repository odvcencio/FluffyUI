package testing

import (
	"fmt"
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
)

// A11ySeverity describes audit severity.
type A11ySeverity string

const (
	A11yError   A11ySeverity = "error"
	A11yWarning A11ySeverity = "warning"
)

// A11yIssue describes an accessibility audit issue.
type A11yIssue struct {
	Path     string
	Message  string
	Severity A11ySeverity
}

// AuditAccessibility walks the widget tree and reports missing accessibility metadata.
func AuditAccessibility(root runtime.Widget) []A11yIssue {
	if root == nil {
		return nil
	}
	var issues []A11yIssue
	var walk func(node runtime.Widget, path []string)
	walk = func(node runtime.Widget, path []string) {
		if node == nil {
			return
		}
		name := widgetName(node)
		currentPath := append(path, name)

		if accessible, ok := node.(accessibility.Accessible); ok {
			role := strings.TrimSpace(string(accessible.AccessibleRole()))
			label := strings.TrimSpace(accessible.AccessibleLabel())
			if role == "" {
				issues = append(issues, A11yIssue{
					Path:     strings.Join(currentPath, "/"),
					Message:  "missing accessible role",
					Severity: A11yError,
				})
			}
			if label == "" {
				severity := A11yWarning
				if focusable, ok := node.(runtime.Focusable); ok && focusable.CanFocus() {
					severity = A11yError
				}
				issues = append(issues, A11yIssue{
					Path:     strings.Join(currentPath, "/"),
					Message:  "missing accessible label",
					Severity: severity,
				})
			}
		}

		provider, ok := node.(runtime.ChildProvider)
		if !ok {
			return
		}
		for _, child := range provider.ChildWidgets() {
			if child == nil {
				continue
			}
			segment := widgetName(child)
			if segmenter, ok := node.(runtime.PathSegmenter); ok {
				if seg := strings.TrimSpace(segmenter.PathSegment(child)); seg != "" {
					segment = seg
				}
			}
			walk(child, append(path, segment))
		}
	}

	walk(root, nil)
	return issues
}

// AssertAccessible fails the test if accessibility errors are found.
func AssertAccessible(t *testing.T, root runtime.Widget) {
	t.Helper()
	issues := AuditAccessibility(root)
	for _, issue := range issues {
		if issue.Severity == A11yWarning {
			t.Logf("a11y warning: %s: %s", issue.Path, issue.Message)
			continue
		}
		t.Errorf("a11y error: %s: %s", issue.Path, issue.Message)
	}
}

func widgetName(w runtime.Widget) string {
	if w == nil {
		return ""
	}
	name := fmt.Sprintf("%T", w)
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return strings.TrimPrefix(name, "*")
}

