package a11ytest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
)

// TreeCheck is an accessibility assertion that operates on a widget tree
// (the root and all its descendants).
type TreeCheck func(root runtime.Widget) error

// AssertTree runs tree-level checks against a root widget, failing the test
// for each error found.
func AssertTree(t testing.TB, root runtime.Widget, checks ...TreeCheck) {
	t.Helper()
	for _, check := range checks {
		if err := check(root); err != nil {
			t.Errorf("a11ytest tree: %s", err)
		}
	}
}

// AllFocusableHaveLabels asserts that every widget implementing
// runtime.Focusable (with CanFocus() == true) also has a non-empty
// accessibility label.
func AllFocusableHaveLabels() TreeCheck {
	return func(root runtime.Widget) error {
		var errs []string
		walkWidgetTree(root, func(w runtime.Widget, path string) {
			f, ok := w.(runtime.Focusable)
			if !ok || !f.CanFocus() {
				return
			}
			acc, ok := w.(accessibility.Accessible)
			if !ok {
				errs = append(errs, fmt.Sprintf("%s: focusable widget does not implement Accessible", path))
				return
			}
			if strings.TrimSpace(acc.AccessibleLabel()) == "" {
				errs = append(errs, fmt.Sprintf("%s: focusable widget has empty label", path))
			}
		})
		if len(errs) > 0 {
			return fmt.Errorf("AllFocusableHaveLabels:\n  %s", strings.Join(errs, "\n  "))
		}
		return nil
	}
}

// AllFocusableHaveRoles asserts that every widget implementing
// runtime.Focusable (with CanFocus() == true) has a non-empty accessibility
// role (i.e., something more specific than "").
func AllFocusableHaveRoles() TreeCheck {
	return func(root runtime.Widget) error {
		var errs []string
		walkWidgetTree(root, func(w runtime.Widget, path string) {
			f, ok := w.(runtime.Focusable)
			if !ok || !f.CanFocus() {
				return
			}
			acc, ok := w.(accessibility.Accessible)
			if !ok {
				errs = append(errs, fmt.Sprintf("%s: focusable widget does not implement Accessible", path))
				return
			}
			role := strings.TrimSpace(string(acc.AccessibleRole()))
			if role == "" {
				errs = append(errs, fmt.Sprintf("%s: focusable widget has empty role", path))
			}
		})
		if len(errs) > 0 {
			return fmt.Errorf("AllFocusableHaveRoles:\n  %s", strings.Join(errs, "\n  "))
		}
		return nil
	}
}

// NoDuplicateIDs asserts that no two widgets in the tree share the same
// non-empty accessibility ID. Widget IDs are obtained via the ID() method
// if the widget implements it.
func NoDuplicateIDs() TreeCheck {
	return func(root runtime.Widget) error {
		type idInfo struct {
			id   string
			path string
		}
		seen := map[string]string{} // id -> first path
		var errs []string
		walkWidgetTree(root, func(w runtime.Widget, path string) {
			idProvider, ok := w.(interface{ ID() string })
			if !ok {
				return
			}
			id := strings.TrimSpace(idProvider.ID())
			if id == "" {
				return
			}
			if firstPath, exists := seen[id]; exists {
				errs = append(errs, fmt.Sprintf("duplicate ID %q: first at %s, also at %s", id, firstPath, path))
			} else {
				seen[id] = path
			}
		})
		if len(errs) > 0 {
			return fmt.Errorf("NoDuplicateIDs:\n  %s", strings.Join(errs, "\n  "))
		}
		return nil
	}
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// walkWidgetTree traverses the widget tree depth-first, calling fn for each
// non-nil widget with a human-readable path string.
func walkWidgetTree(root runtime.Widget, fn func(w runtime.Widget, path string)) {
	if root == nil {
		return
	}
	var walk func(node runtime.Widget, path []string)
	walk = func(node runtime.Widget, path []string) {
		if node == nil {
			return
		}
		name := widgetTypeName(node)
		currentPath := append(path, name)
		pathStr := strings.Join(currentPath, "/")

		fn(node, pathStr)

		provider, ok := node.(runtime.ChildProvider)
		if !ok {
			return
		}
		for _, child := range provider.ChildWidgets() {
			if child == nil {
				continue
			}
			walk(child, currentPath)
		}
	}
	walk(root, nil)
}

// widgetTypeName returns a short type name for a widget.
func widgetTypeName(w runtime.Widget) string {
	if w == nil {
		return ""
	}
	name := fmt.Sprintf("%T", w)
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	return strings.TrimPrefix(name, "*")
}
