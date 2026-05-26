// Package a11ytest provides ergonomic accessibility assertion functions for
// widget testing. It allows widget authors to write a11y tests as naturally
// as unit tests using composable Check functions.
//
// Example:
//
//	btn := widgets.NewButton("Save")
//	a11ytest.Assert(t, btn,
//	    a11ytest.HasRole(accessibility.RoleButton),
//	    a11ytest.HasLabel("Save"),
//	    a11ytest.IsFocusable(),
//	)
package a11ytest

import (
	"fmt"
	"strings"
	"testing"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

// Check is an accessibility assertion that returns nil on success or an error
// describing what failed.
type Check func(w runtime.Widget) error

// Assert runs checks against a widget, failing the test on the first error.
func Assert(t testing.TB, w runtime.Widget, checks ...Check) {
	t.Helper()
	for _, check := range checks {
		if err := check(w); err != nil {
			t.Errorf("a11ytest: %s", err)
			return
		}
	}
}

// AssertAll runs all checks and reports all failures rather than stopping at
// the first one.
func AssertAll(t testing.TB, w runtime.Widget, checks ...Check) {
	t.Helper()
	for _, check := range checks {
		if err := check(w); err != nil {
			t.Errorf("a11ytest: %s", err)
		}
	}
}

// ---------------------------------------------------------------------------
// Identity checks
// ---------------------------------------------------------------------------

// HasRole asserts that the widget has the given accessibility role.
func HasRole(role accessibility.Role) Check {
	return func(w runtime.Widget) error {
		acc, ok := w.(accessibility.Accessible)
		if !ok {
			return fmt.Errorf("HasRole: widget %T does not implement accessibility.Accessible", w)
		}
		got := acc.AccessibleRole()
		if got != role {
			return fmt.Errorf("HasRole: got %q, want %q", got, role)
		}
		return nil
	}
}

// HasLabel asserts that the widget's accessibility label exactly matches the
// given string.
func HasLabel(label string) Check {
	return func(w runtime.Widget) error {
		acc, ok := w.(accessibility.Accessible)
		if !ok {
			return fmt.Errorf("HasLabel: widget %T does not implement accessibility.Accessible", w)
		}
		got := acc.AccessibleLabel()
		if got != label {
			return fmt.Errorf("HasLabel: got %q, want %q", got, label)
		}
		return nil
	}
}

// HasLabelContaining asserts that the widget's accessibility label contains
// the given substring.
func HasLabelContaining(substr string) Check {
	return func(w runtime.Widget) error {
		acc, ok := w.(accessibility.Accessible)
		if !ok {
			return fmt.Errorf("HasLabelContaining: widget %T does not implement accessibility.Accessible", w)
		}
		got := acc.AccessibleLabel()
		if !strings.Contains(got, substr) {
			return fmt.Errorf("HasLabelContaining: label %q does not contain %q", got, substr)
		}
		return nil
	}
}

// HasDescription asserts that the widget's accessibility description exactly
// matches the given string.
func HasDescription(desc string) Check {
	return func(w runtime.Widget) error {
		acc, ok := w.(accessibility.Accessible)
		if !ok {
			return fmt.Errorf("HasDescription: widget %T does not implement accessibility.Accessible", w)
		}
		got := acc.AccessibleDescription()
		if got != desc {
			return fmt.Errorf("HasDescription: got %q, want %q", got, desc)
		}
		return nil
	}
}

// IsFocusable asserts that the widget implements runtime.Focusable and reports
// CanFocus() == true.
func IsFocusable() Check {
	return func(w runtime.Widget) error {
		f, ok := w.(runtime.Focusable)
		if !ok {
			return fmt.Errorf("IsFocusable: widget %T does not implement runtime.Focusable", w)
		}
		if !f.CanFocus() {
			return fmt.Errorf("IsFocusable: widget %T reports CanFocus() == false", w)
		}
		return nil
	}
}

// ---------------------------------------------------------------------------
// State checks
// ---------------------------------------------------------------------------

// IsDisabled asserts that the widget's accessible state has Disabled == true.
func IsDisabled() Check {
	return func(w runtime.Widget) error {
		acc, ok := w.(accessibility.Accessible)
		if !ok {
			return fmt.Errorf("IsDisabled: widget %T does not implement accessibility.Accessible", w)
		}
		state := acc.AccessibleState()
		if !state.Disabled {
			return fmt.Errorf("IsDisabled: widget is not disabled")
		}
		return nil
	}
}

// IsEnabled asserts that the widget's accessible state has Disabled == false.
func IsEnabled() Check {
	return func(w runtime.Widget) error {
		acc, ok := w.(accessibility.Accessible)
		if !ok {
			return fmt.Errorf("IsEnabled: widget %T does not implement accessibility.Accessible", w)
		}
		state := acc.AccessibleState()
		if state.Disabled {
			return fmt.Errorf("IsEnabled: widget is disabled")
		}
		return nil
	}
}

// IsChecked asserts that the widget's Checked state is non-nil and true.
func IsChecked() Check {
	return func(w runtime.Widget) error {
		acc, ok := w.(accessibility.Accessible)
		if !ok {
			return fmt.Errorf("IsChecked: widget %T does not implement accessibility.Accessible", w)
		}
		state := acc.AccessibleState()
		if state.Checked == nil {
			return fmt.Errorf("IsChecked: Checked state is nil (not applicable)")
		}
		if !*state.Checked {
			return fmt.Errorf("IsChecked: widget is not checked")
		}
		return nil
	}
}

// IsExpanded asserts that the widget's Expanded state is non-nil and true.
func IsExpanded() Check {
	return func(w runtime.Widget) error {
		acc, ok := w.(accessibility.Accessible)
		if !ok {
			return fmt.Errorf("IsExpanded: widget %T does not implement accessibility.Accessible", w)
		}
		state := acc.AccessibleState()
		if state.Expanded == nil {
			return fmt.Errorf("IsExpanded: Expanded state is nil (not applicable)")
		}
		if !*state.Expanded {
			return fmt.Errorf("IsExpanded: widget is not expanded")
		}
		return nil
	}
}

// IsCollapsed asserts that the widget's Expanded state is non-nil and false.
func IsCollapsed() Check {
	return func(w runtime.Widget) error {
		acc, ok := w.(accessibility.Accessible)
		if !ok {
			return fmt.Errorf("IsCollapsed: widget %T does not implement accessibility.Accessible", w)
		}
		state := acc.AccessibleState()
		if state.Expanded == nil {
			return fmt.Errorf("IsCollapsed: Expanded state is nil (not applicable)")
		}
		if *state.Expanded {
			return fmt.Errorf("IsCollapsed: widget is expanded, not collapsed")
		}
		return nil
	}
}

// IsSelected asserts that the widget's Selected state is true.
func IsSelected() Check {
	return func(w runtime.Widget) error {
		acc, ok := w.(accessibility.Accessible)
		if !ok {
			return fmt.Errorf("IsSelected: widget %T does not implement accessibility.Accessible", w)
		}
		state := acc.AccessibleState()
		if !state.Selected {
			return fmt.Errorf("IsSelected: widget is not selected")
		}
		return nil
	}
}

// IsRequired asserts that the widget's Required state is true.
func IsRequired() Check {
	return func(w runtime.Widget) error {
		acc, ok := w.(accessibility.Accessible)
		if !ok {
			return fmt.Errorf("IsRequired: widget %T does not implement accessibility.Accessible", w)
		}
		state := acc.AccessibleState()
		if !state.Required {
			return fmt.Errorf("IsRequired: widget is not required")
		}
		return nil
	}
}

// IsInvalid asserts that the widget's Invalid state is true.
func IsInvalid() Check {
	return func(w runtime.Widget) error {
		acc, ok := w.(accessibility.Accessible)
		if !ok {
			return fmt.Errorf("IsInvalid: widget %T does not implement accessibility.Accessible", w)
		}
		state := acc.AccessibleState()
		if !state.Invalid {
			return fmt.Errorf("IsInvalid: widget is not invalid")
		}
		return nil
	}
}

// HasValue asserts that the widget's accessible value has matching Text.
func HasValue(text string) Check {
	return func(w runtime.Widget) error {
		acc, ok := w.(accessibility.Accessible)
		if !ok {
			return fmt.Errorf("HasValue: widget %T does not implement accessibility.Accessible", w)
		}
		v := acc.AccessibleValue()
		if v == nil {
			return fmt.Errorf("HasValue: widget has no value (AccessibleValue is nil)")
		}
		if v.Text != text {
			return fmt.Errorf("HasValue: got %q, want %q", v.Text, text)
		}
		return nil
	}
}

// ---------------------------------------------------------------------------
// Keyboard checks
// ---------------------------------------------------------------------------

// HandlesKey sends a KeyMsg with the given key to the widget and asserts that
// the result is Handled. The widget must be focused first if it checks focus
// state internally.
func HandlesKey(key terminal.Key) Check {
	return func(w runtime.Widget) error {
		msg := runtime.KeyMsg{Key: key}
		result := w.HandleMessage(msg)
		if !result.Handled {
			return fmt.Errorf("HandlesKey: widget did not handle key %d", key)
		}
		return nil
	}
}

// HandlesKeyRune sends a KeyMsg with KeyRune and the given rune, and asserts
// that the result is Handled.
func HandlesKeyRune(r rune) Check {
	return func(w runtime.Widget) error {
		msg := runtime.KeyMsg{Key: terminal.KeyRune, Rune: r}
		result := w.HandleMessage(msg)
		if !result.Handled {
			return fmt.Errorf("HandlesKeyRune: widget did not handle rune %q", r)
		}
		return nil
	}
}

// ActivatableByKeyboard asserts that the widget handles either Enter or Space.
// The widget should be focused before running this check.
func ActivatableByKeyboard() Check {
	return func(w runtime.Widget) error {
		enterMsg := runtime.KeyMsg{Key: terminal.KeyEnter}
		spaceMsg := runtime.KeyMsg{Key: terminal.KeyRune, Rune: ' '}
		enterResult := w.HandleMessage(enterMsg)
		spaceResult := w.HandleMessage(spaceMsg)
		if !enterResult.Handled && !spaceResult.Handled {
			return fmt.Errorf("ActivatableByKeyboard: widget handles neither Enter nor Space")
		}
		return nil
	}
}
