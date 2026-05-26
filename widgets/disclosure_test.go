package widgets

import (
	"strings"
	"testing"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
	fluffytest "m31labs.dev/fluffyui/testing"
)

func TestDisclosure_New(t *testing.T) {
	d := NewDisclosure("Details", NewLabel("Content here"))
	if d == nil {
		t.Fatal("expected non-nil Disclosure")
	}
	if d.Expanded() {
		t.Fatal("expected collapsed by default")
	}
	if !d.CanFocus() {
		t.Fatal("expected disclosure to be focusable")
	}
}

func TestDisclosure_Role(t *testing.T) {
	d := NewDisclosure("Info", NewLabel("Details"))
	if d.AccessibleRole() != accessibility.RoleButton {
		t.Fatalf("expected role button, got %q", d.AccessibleRole())
	}
	st := d.AccessibleState()
	if st.Expanded == nil {
		t.Fatal("expected Expanded to be set")
	}
	if *st.Expanded {
		t.Fatal("expected Expanded to be false")
	}
}

func TestDisclosure_Toggle(t *testing.T) {
	var changed bool
	d := NewDisclosure("Title", NewLabel("Body"))
	d.SetOnChange(func(expanded bool) { changed = expanded })

	d.Toggle()
	if !d.Expanded() {
		t.Fatal("expected expanded after Toggle")
	}
	if !changed {
		t.Fatal("expected onChange called with true")
	}

	d.Toggle()
	if d.Expanded() {
		t.Fatal("expected collapsed after second Toggle")
	}
	if changed {
		t.Fatal("expected onChange called with false")
	}
}

func TestDisclosure_SetExpanded(t *testing.T) {
	d := NewDisclosure("Title", NewLabel("Body"))
	d.SetExpanded(true)
	if !d.Expanded() {
		t.Fatal("expected expanded")
	}
	// Setting same value should be a no-op
	d.SetExpanded(true)
	if !d.Expanded() {
		t.Fatal("expected still expanded")
	}
	d.SetExpanded(false)
	if d.Expanded() {
		t.Fatal("expected collapsed")
	}
}

func TestDisclosure_KeyboardEnter(t *testing.T) {
	d := NewDisclosure("Section", NewLabel("Content"))
	d.Focus()

	result := d.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !result.Handled {
		t.Fatal("expected Enter to be handled")
	}
	if !d.Expanded() {
		t.Fatal("expected expanded after Enter")
	}
}

func TestDisclosure_KeyboardSpace(t *testing.T) {
	d := NewDisclosure("Section", NewLabel("Content"))
	d.Focus()

	result := d.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ' '})
	if !result.Handled {
		t.Fatal("expected Space to be handled")
	}
	if !d.Expanded() {
		t.Fatal("expected expanded after Space")
	}
}

func TestDisclosure_UnfocusedNoHandle(t *testing.T) {
	d := NewDisclosure("Title", NewLabel("Body"))
	result := d.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if result.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestDisclosure_Measure(t *testing.T) {
	d := NewDisclosure("Details", NewLabel("Content text"))
	// Collapsed: just header
	size := d.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 10})
	if size.Height != 1 {
		t.Fatalf("expected height 1 when collapsed, got %d", size.Height)
	}

	// Expanded: header + content
	d.SetExpanded(true)
	size = d.Measure(runtime.Constraints{MaxWidth: 40, MaxHeight: 10})
	if size.Height < 2 {
		t.Fatalf("expected height >= 2 when expanded, got %d", size.Height)
	}
}

func TestDisclosure_RenderCollapsed(t *testing.T) {
	d := NewDisclosure("Details", NewLabel("Hidden content"))
	output := fluffytest.RenderToString(d, 30, 1)
	if !strings.Contains(output, "Details") {
		t.Fatalf("expected title in output, got %q", output)
	}
}

func TestDisclosure_RenderExpanded(t *testing.T) {
	d := NewDisclosure("Details", NewLabel("Visible content"))
	d.SetExpanded(true)
	output := fluffytest.RenderToString(d, 30, 3)
	if !strings.Contains(output, "Details") {
		t.Fatalf("expected title in output, got %q", output)
	}
	if !strings.Contains(output, "Visible content") {
		t.Fatalf("expected content in output, got %q", output)
	}
}

func TestDisclosure_ChildWidgets(t *testing.T) {
	content := NewLabel("Body")
	d := NewDisclosure("Title", content)

	// Collapsed: no children
	if len(d.ChildWidgets()) != 0 {
		t.Fatal("expected no children when collapsed")
	}

	// Expanded: content is a child
	d.SetExpanded(true)
	children := d.ChildWidgets()
	if len(children) != 1 {
		t.Fatalf("expected 1 child when expanded, got %d", len(children))
	}
}

func TestDisclosure_A11yExpandedState(t *testing.T) {
	d := NewDisclosure("Info", NewLabel("Detail"))
	st := d.AccessibleState()
	if st.Expanded == nil || *st.Expanded {
		t.Fatal("expected Expanded false when collapsed")
	}

	d.SetExpanded(true)
	st = d.AccessibleState()
	if st.Expanded == nil || !*st.Expanded {
		t.Fatal("expected Expanded true when expanded")
	}
}
