package widgets

import (
	"strings"
	"testing"

	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/state"
	"m31labs.dev/fluffyui/terminal"
	fluffytest "m31labs.dev/fluffyui/testing"
)

func TestToggle_New(t *testing.T) {
	sig := state.NewSignal(false)
	tog := NewToggle(sig)
	if tog == nil {
		t.Fatal("expected non-nil Toggle")
	}
	if tog.On() {
		t.Fatal("expected initial state false")
	}
	if !tog.CanFocus() {
		t.Fatal("expected toggle to be focusable")
	}
}

func TestToggle_Role(t *testing.T) {
	sig := state.NewSignal(false)
	tog := NewToggle(sig)
	if tog.AccessibleRole() != "switch" {
		t.Fatalf("expected role switch, got %q", tog.AccessibleRole())
	}
	if tog.AccessibleRoleDescription() != "toggle switch" {
		t.Fatalf("expected role description 'toggle switch', got %q", tog.AccessibleRoleDescription())
	}
}

func TestToggle_WithLabel(t *testing.T) {
	sig := state.NewSignal(true)
	label := state.NewSignal("Dark mode")
	tog := NewToggle(sig, WithToggleLabel(label))
	if tog.Label() != "Dark mode" {
		t.Fatalf("expected label 'Dark mode', got %q", tog.Label())
	}
	if tog.AccessibleLabel() != "Dark mode" {
		t.Fatalf("expected accessible label 'Dark mode', got %q", tog.AccessibleLabel())
	}
}

func TestToggle_SetOn(t *testing.T) {
	sig := state.NewSignal(false)
	var changed bool
	tog := NewToggle(sig, WithToggleOnChange(func(v bool) {
		changed = v
	}))

	tog.SetOn(true)
	if !tog.On() {
		t.Fatal("expected on after SetOn(true)")
	}
	if !changed {
		t.Fatal("expected onChange called with true")
	}

	tog.SetOn(false)
	if tog.On() {
		t.Fatal("expected off after SetOn(false)")
	}
	if changed {
		t.Fatal("expected onChange called with false")
	}
}

func TestToggle_KeyToggle(t *testing.T) {
	sig := state.NewSignal(false)
	tog := NewToggle(sig)
	tog.Focus()

	// Space toggles on.
	result := tog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ' '})
	if !result.Handled {
		t.Fatal("expected space to be handled")
	}
	if !tog.On() {
		t.Fatal("expected on after space")
	}

	// Enter toggles off.
	result = tog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Fatal("expected enter to be handled")
	}
	if tog.On() {
		t.Fatal("expected off after enter")
	}
}

func TestToggle_UnfocusedNoHandle(t *testing.T) {
	sig := state.NewSignal(false)
	tog := NewToggle(sig)
	result := tog.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestToggle_Measure(t *testing.T) {
	sig := state.NewSignal(false)
	tog := NewToggle(sig)
	size := tog.Measure(runtime.Constraints{MaxWidth: 80, MaxHeight: 10})
	if size.Width < 6 {
		t.Fatalf("expected width >= 6, got %d", size.Width)
	}
	if size.Height != 1 {
		t.Fatalf("expected height 1, got %d", size.Height)
	}
}

func TestToggle_Render(t *testing.T) {
	sig := state.NewSignal(true)
	tog := NewToggle(sig)
	output := fluffytest.RenderToString(tog, 20, 1)
	if !strings.Contains(output, "On") {
		t.Fatalf("expected output to contain 'On', got %q", output)
	}
}

func TestToggle_A11yState(t *testing.T) {
	sig := state.NewSignal(true)
	tog := NewToggle(sig)
	st := tog.AccessibleState()
	if st.Checked == nil {
		t.Fatal("expected Checked to be set")
	}
	if !*st.Checked {
		t.Fatal("expected Checked to be true when on")
	}
}
