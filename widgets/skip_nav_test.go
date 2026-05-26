package widgets

import (
	"strings"
	"testing"

	"m31labs.dev/fluffyui/runtime"
	fluffytest "m31labs.dev/fluffyui/testing"
	"m31labs.dev/fluffyui/terminal"
)

func TestSkipNav_NewSkipNav(t *testing.T) {
	target := newFTTestFocusable("main")
	sn := NewSkipNav("Skip to main content", target)

	if sn == nil {
		t.Fatal("expected non-nil SkipNav")
	}
	if sn.Label() != "Skip to main content" {
		t.Errorf("expected label 'Skip to main content', got %q", sn.Label())
	}
	if sn.Target() != target {
		t.Error("expected target to match")
	}
	if !sn.CanFocus() {
		t.Error("SkipNav should be focusable")
	}
}

func TestSkipNav_AccessibilityRole(t *testing.T) {
	sn := NewSkipNav("Skip", newFTTestFocusable("main"))
	if sn.AccessibleRole() != "link" {
		t.Errorf("expected role 'link', got %q", sn.AccessibleRole())
	}
	if sn.AccessibleLabel() != "Skip" {
		t.Errorf("expected accessible label 'Skip', got %q", sn.AccessibleLabel())
	}
}

func TestSkipNav_HiddenWhenUnfocused(t *testing.T) {
	sn := NewSkipNav("Skip to content", newFTTestFocusable("main"))
	// Not focused - should measure to zero height
	size := sn.Measure(runtime.Loose(80, 24))
	if size.Height != 0 {
		t.Errorf("expected height 0 when unfocused, got %d", size.Height)
	}
}

func TestSkipNav_VisibleWhenFocused(t *testing.T) {
	sn := NewSkipNav("Skip to content", newFTTestFocusable("main"))
	sn.Focus()

	size := sn.Measure(runtime.Loose(80, 24))
	if size.Height != 1 {
		t.Errorf("expected height 1 when focused, got %d", size.Height)
	}
	if size.Width < 1 {
		t.Errorf("expected positive width when focused, got %d", size.Width)
	}
}

func TestSkipNav_RendersLabelWhenFocused(t *testing.T) {
	sn := NewSkipNav("Skip to main", newFTTestFocusable("main"))
	sn.Focus()

	output := fluffytest.RenderToString(sn, 30, 1)
	if !strings.Contains(output, "Skip to main") {
		t.Errorf("expected output to contain 'Skip to main', got %q", output)
	}
}

func TestSkipNav_DoesNotRenderWhenUnfocused(t *testing.T) {
	sn := NewSkipNav("Skip to main", newFTTestFocusable("main"))
	// Not focused - should not render any text
	output := fluffytest.RenderToString(sn, 30, 1)
	if strings.Contains(output, "Skip to main") {
		t.Errorf("expected output to NOT contain 'Skip to main' when unfocused, got %q", output)
	}
}

func TestSkipNav_EnterActivatesFocusTransfer(t *testing.T) {
	target := newFTTestFocusable("main")
	sn := NewSkipNav("Skip to main content", target)
	sn.Focus()

	// Press Enter
	result := sn.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Error("expected Enter to be handled")
	}

	// SkipNav should be blurred
	if sn.IsFocused() {
		t.Error("expected SkipNav to be blurred after activation")
	}

	// Target should be focused
	if !target.IsFocused() {
		t.Error("expected target to be focused after activation")
	}
}

func TestSkipNav_OtherKeysNotHandled(t *testing.T) {
	sn := NewSkipNav("Skip", newFTTestFocusable("main"))
	sn.Focus()

	result := sn.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if result.Handled {
		t.Error("expected Escape to not be handled")
	}

	result = sn.HandleMessage(runtime.KeyMsg{Key: terminal.KeyTab})
	if result.Handled {
		t.Error("expected Tab to not be handled")
	}
}

func TestSkipNav_UnfocusedIgnoresKeys(t *testing.T) {
	sn := NewSkipNav("Skip", newFTTestFocusable("main"))
	// Not focused

	result := sn.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if result.Handled {
		t.Error("expected Enter to be ignored when unfocused")
	}
}

func TestSkipNav_NonKeyMessageIgnored(t *testing.T) {
	sn := NewSkipNav("Skip", newFTTestFocusable("main"))
	sn.Focus()

	result := sn.HandleMessage(runtime.MouseMsg{X: 0, Y: 0})
	if result.Handled {
		t.Error("expected mouse message to not be handled")
	}
}

func TestSkipNav_SetLabel(t *testing.T) {
	sn := NewSkipNav("Skip", newFTTestFocusable("main"))
	sn.SetLabel("Skip to footer")
	if sn.Label() != "Skip to footer" {
		t.Errorf("expected label 'Skip to footer', got %q", sn.Label())
	}
	if sn.Base.Label != "Skip to footer" {
		t.Errorf("expected accessible label 'Skip to footer', got %q", sn.Base.Label)
	}
}

func TestSkipNav_SetTarget(t *testing.T) {
	target1 := newFTTestFocusable("main")
	target2 := newFTTestFocusable("footer")
	sn := NewSkipNav("Skip", target1)

	sn.SetTarget(target2)
	if sn.Target() != target2 {
		t.Error("expected target to be updated")
	}
}

func TestSkipNav_StyleType(t *testing.T) {
	sn := NewSkipNav("Skip", newFTTestFocusable("main"))
	if sn.StyleType() != "SkipNav" {
		t.Errorf("expected StyleType 'SkipNav', got %q", sn.StyleType())
	}
}

func TestSkipNav_FocusAffectsLayout(t *testing.T) {
	sn := NewSkipNav("Skip", newFTTestFocusable("main"))
	if !sn.FocusAffectsLayout() {
		t.Error("expected FocusAffectsLayout() to return true")
	}
}

func TestSkipNav_BindUnbind(t *testing.T) {
	sn := NewSkipNav("Skip", newFTTestFocusable("main"))
	// Should not panic
	sn.Bind(runtime.Services{})
	sn.Unbind()
}

func TestSkipNav_NilReceiver(t *testing.T) {
	var sn *SkipNav

	// All nil receiver methods should not panic
	sn.SetLabel("test")
	sn.SetTarget(nil)
	sn.Bind(runtime.Services{})
	sn.Unbind()

	if sn.Label() != "" {
		t.Error("nil SkipNav should return empty label")
	}
	if sn.Target() != nil {
		t.Error("nil SkipNav should return nil target")
	}
}

func TestSkipNav_NilTarget(t *testing.T) {
	sn := NewSkipNav("Skip", nil)
	sn.Focus()

	// Should not panic when activating with nil target
	result := sn.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Error("expected Enter to be handled even with nil target")
	}
}

func TestSkipNav_MeasureWidth(t *testing.T) {
	sn := NewSkipNav("ABC", newFTTestFocusable("main"))
	sn.Focus()

	size := sn.Measure(runtime.Loose(80, 24))
	if size.Width != 3 {
		t.Errorf("expected width 3 for label 'ABC', got %d", size.Width)
	}
}
