package a11ytest_test

import (
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/testing/a11ytest"
	"github.com/odvcencio/fluffyui/widgets"
)

// ---------------------------------------------------------------------------
// Assert / AssertAll basics
// ---------------------------------------------------------------------------

func TestAssertStopsOnFirstError(t *testing.T) {
	btn := widgets.NewButton("OK")
	ft := &fakeTB{}
	// First check passes (role is button), second fails (label is not "Nope").
	a11ytest.Assert(ft, btn,
		a11ytest.HasRole(accessibility.RoleButton),
		a11ytest.HasLabel("Nope"),
		a11ytest.HasLabel("also bad"), // should not be reached
	)
	if ft.errorCount != 1 {
		t.Errorf("Assert should stop after first error, got %d errors", ft.errorCount)
	}
}

func TestAssertAllReportsAllErrors(t *testing.T) {
	btn := widgets.NewButton("OK")
	ft := &fakeTB{}
	a11ytest.AssertAll(ft, btn,
		a11ytest.HasLabel("wrong1"),
		a11ytest.HasLabel("wrong2"),
	)
	if ft.errorCount != 2 {
		t.Errorf("AssertAll should report all errors, got %d", ft.errorCount)
	}
}

func TestAssertNoErrorOnPass(t *testing.T) {
	btn := widgets.NewButton("Save")
	ft := &fakeTB{}
	a11ytest.Assert(ft, btn,
		a11ytest.HasRole(accessibility.RoleButton),
		a11ytest.HasLabel("Save"),
	)
	if ft.errorCount != 0 {
		t.Errorf("Assert should not error on passing checks, got %d errors", ft.errorCount)
	}
}

// ---------------------------------------------------------------------------
// Identity checks
// ---------------------------------------------------------------------------

func TestHasRole(t *testing.T) {
	btn := widgets.NewButton("Click")
	a11ytest.Assert(t, btn, a11ytest.HasRole(accessibility.RoleButton))

	cb := widgets.NewCheckbox("Toggle")
	a11ytest.Assert(t, cb, a11ytest.HasRole(accessibility.RoleCheckbox))

	txt := widgets.NewText("hello")
	a11ytest.Assert(t, txt, a11ytest.HasRole(accessibility.RoleText))
}

func TestHasRoleMismatch(t *testing.T) {
	btn := widgets.NewButton("Click")
	err := a11ytest.HasRole(accessibility.RoleCheckbox)(btn)
	if err == nil {
		t.Fatal("expected error for role mismatch")
	}
}

func TestHasLabel(t *testing.T) {
	btn := widgets.NewButton("Submit")
	a11ytest.Assert(t, btn, a11ytest.HasLabel("Submit"))
}

func TestHasLabelMismatch(t *testing.T) {
	btn := widgets.NewButton("Submit")
	err := a11ytest.HasLabel("Cancel")(btn)
	if err == nil {
		t.Fatal("expected error for label mismatch")
	}
}

func TestHasLabelContaining(t *testing.T) {
	btn := widgets.NewButton("Save Document")
	a11ytest.Assert(t, btn, a11ytest.HasLabelContaining("Document"))
	a11ytest.Assert(t, btn, a11ytest.HasLabelContaining("Save"))
}

func TestHasLabelContainingMismatch(t *testing.T) {
	btn := widgets.NewButton("Save")
	err := a11ytest.HasLabelContaining("Delete")(btn)
	if err == nil {
		t.Fatal("expected error for label not containing substring")
	}
}

func TestHasDescription(t *testing.T) {
	btn := widgets.NewButton("OK")
	// Button description is "" by default.
	a11ytest.Assert(t, btn, a11ytest.HasDescription(""))
}

func TestHasDescriptionMismatch(t *testing.T) {
	btn := widgets.NewButton("OK")
	err := a11ytest.HasDescription("some description")(btn)
	if err == nil {
		t.Fatal("expected error for description mismatch")
	}
}

func TestIsFocusable(t *testing.T) {
	btn := widgets.NewButton("OK")
	a11ytest.Assert(t, btn, a11ytest.IsFocusable())

	cb := widgets.NewCheckbox("Accept")
	a11ytest.Assert(t, cb, a11ytest.IsFocusable())

	input := widgets.NewInput()
	a11ytest.Assert(t, input, a11ytest.IsFocusable())
}

func TestIsFocusableFailsForNonFocusable(t *testing.T) {
	txt := widgets.NewText("hello")
	err := a11ytest.IsFocusable()(txt)
	if err == nil {
		t.Fatal("expected error: Text widget is not focusable")
	}
}

// ---------------------------------------------------------------------------
// State checks
// ---------------------------------------------------------------------------

func TestIsEnabled(t *testing.T) {
	btn := widgets.NewButton("OK")
	a11ytest.Assert(t, btn, a11ytest.IsEnabled())
}

func TestIsDisabledMismatch(t *testing.T) {
	btn := widgets.NewButton("OK")
	err := a11ytest.IsDisabled()(btn)
	if err == nil {
		t.Fatal("expected error: button is not disabled")
	}
}

func TestIsChecked(t *testing.T) {
	cb := widgets.NewCheckbox("Accept")
	checked := true
	cb.SetChecked(&checked)
	a11ytest.Assert(t, cb, a11ytest.IsChecked())
}

func TestIsCheckedMismatchUnchecked(t *testing.T) {
	cb := widgets.NewCheckbox("Accept")
	unchecked := false
	cb.SetChecked(&unchecked)
	err := a11ytest.IsChecked()(cb)
	if err == nil {
		t.Fatal("expected error: checkbox is unchecked")
	}
}

func TestIsExpanded(t *testing.T) {
	w := &mockExpandable{expanded: true}
	a11ytest.Assert(t, w, a11ytest.IsExpanded())
}

func TestIsCollapsed(t *testing.T) {
	w := &mockExpandable{expanded: false}
	a11ytest.Assert(t, w, a11ytest.IsCollapsed())
}

func TestIsExpandedMismatch(t *testing.T) {
	w := &mockExpandable{expanded: false}
	err := a11ytest.IsExpanded()(w)
	if err == nil {
		t.Fatal("expected error: widget is not expanded")
	}
}

func TestIsCollapsedMismatch(t *testing.T) {
	w := &mockExpandable{expanded: true}
	err := a11ytest.IsCollapsed()(w)
	if err == nil {
		t.Fatal("expected error: widget is expanded, not collapsed")
	}
}

func TestIsSelected(t *testing.T) {
	w := &mockSelectable{selected: true}
	a11ytest.Assert(t, w, a11ytest.IsSelected())
}

func TestIsSelectedMismatch(t *testing.T) {
	w := &mockSelectable{selected: false}
	err := a11ytest.IsSelected()(w)
	if err == nil {
		t.Fatal("expected error: widget is not selected")
	}
}

func TestIsRequired(t *testing.T) {
	w := &mockStateful{required: true}
	a11ytest.Assert(t, w, a11ytest.IsRequired())
}

func TestIsRequiredMismatch(t *testing.T) {
	w := &mockStateful{}
	err := a11ytest.IsRequired()(w)
	if err == nil {
		t.Fatal("expected error: widget is not required")
	}
}

func TestIsInvalid(t *testing.T) {
	w := &mockStateful{invalid: true}
	a11ytest.Assert(t, w, a11ytest.IsInvalid())
}

func TestIsInvalidMismatch(t *testing.T) {
	w := &mockStateful{}
	err := a11ytest.IsInvalid()(w)
	if err == nil {
		t.Fatal("expected error: widget is not invalid")
	}
}

func TestHasValue(t *testing.T) {
	w := &mockWithValue{text: "hello"}
	a11ytest.Assert(t, w, a11ytest.HasValue("hello"))
}

func TestHasValueMismatch(t *testing.T) {
	w := &mockWithValue{text: "hello"}
	err := a11ytest.HasValue("world")(w)
	if err == nil {
		t.Fatal("expected error for value mismatch")
	}
}

func TestHasValueNilValue(t *testing.T) {
	w := &mockStateful{}
	err := a11ytest.HasValue("any")(w)
	if err == nil {
		t.Fatal("expected error when value is nil")
	}
}

// ---------------------------------------------------------------------------
// Keyboard checks
// ---------------------------------------------------------------------------

func TestHandlesKey(t *testing.T) {
	btn := widgets.NewButton("OK")
	btn.Focus()
	a11ytest.Assert(t, btn, a11ytest.HandlesKey(terminal.KeyEnter))
}

func TestHandlesKeyMismatch(t *testing.T) {
	btn := widgets.NewButton("OK")
	btn.Focus()
	// Button does not handle Escape.
	err := a11ytest.HandlesKey(terminal.KeyEscape)(btn)
	if err == nil {
		t.Fatal("expected error: button does not handle Escape")
	}
}

func TestHandlesKeyRune(t *testing.T) {
	btn := widgets.NewButton("OK")
	btn.Focus()
	a11ytest.Assert(t, btn, a11ytest.HandlesKeyRune(' '))
}

func TestActivatableByKeyboard(t *testing.T) {
	btn := widgets.NewButton("Go")
	btn.Focus()
	a11ytest.Assert(t, btn, a11ytest.ActivatableByKeyboard())
}

func TestActivatableByKeyboardFails(t *testing.T) {
	txt := widgets.NewText("static")
	err := a11ytest.ActivatableByKeyboard()(txt)
	if err == nil {
		t.Fatal("expected error: text widget is not activatable by keyboard")
	}
}

// ---------------------------------------------------------------------------
// Tree checks
// ---------------------------------------------------------------------------

func TestAllFocusableHaveLabels(t *testing.T) {
	root := &mockContainer{
		children: []runtime.Widget{
			widgets.NewButton("Save"),
			widgets.NewCheckbox("Accept"),
		},
	}
	a11ytest.AssertTree(t, root, a11ytest.AllFocusableHaveLabels())
}

func TestAllFocusableHaveRoles(t *testing.T) {
	root := &mockContainer{
		children: []runtime.Widget{
			widgets.NewButton("Save"),
			widgets.NewCheckbox("Accept"),
		},
	}
	a11ytest.AssertTree(t, root, a11ytest.AllFocusableHaveRoles())
}

func TestNoDuplicateIDs(t *testing.T) {
	btn1 := widgets.NewButton("A")
	btn1.SetID("btn-1")
	btn2 := widgets.NewButton("B")
	btn2.SetID("btn-2")

	root := &mockContainer{
		children: []runtime.Widget{btn1, btn2},
	}
	a11ytest.AssertTree(t, root, a11ytest.NoDuplicateIDs())
}

func TestNoDuplicateIDsFailsOnDuplicate(t *testing.T) {
	btn1 := widgets.NewButton("A")
	btn1.SetID("same-id")
	btn2 := widgets.NewButton("B")
	btn2.SetID("same-id")

	root := &mockContainer{
		children: []runtime.Widget{btn1, btn2},
	}
	err := a11ytest.NoDuplicateIDs()(root)
	if err == nil {
		t.Fatal("expected error for duplicate IDs")
	}
}

func TestAllFocusableHaveLabelsFailsForEmptyLabel(t *testing.T) {
	btn := widgets.NewButton("")
	root := &mockContainer{
		children: []runtime.Widget{btn},
	}
	err := a11ytest.AllFocusableHaveLabels()(root)
	if err == nil {
		t.Fatal("expected error: button with empty label")
	}
}

func TestAssertTreeNilRoot(t *testing.T) {
	// Should not panic.
	a11ytest.AssertTree(t, nil, a11ytest.AllFocusableHaveLabels())
}

// ---------------------------------------------------------------------------
// Combined usage: multiple checks on real widgets
// ---------------------------------------------------------------------------

func TestButtonFullAssert(t *testing.T) {
	btn := widgets.NewButton("Delete", widgets.WithVariant(widgets.VariantDanger))
	btn.Focus()

	a11ytest.AssertAll(t, btn,
		a11ytest.HasRole(accessibility.RoleButton),
		a11ytest.HasLabel("Delete"),
		a11ytest.HasDescription(""),
		a11ytest.IsFocusable(),
		a11ytest.IsEnabled(),
		a11ytest.ActivatableByKeyboard(),
	)
}

func TestCheckboxFullAssert(t *testing.T) {
	cb := widgets.NewCheckbox("I agree")
	checked := true
	cb.SetChecked(&checked)

	a11ytest.AssertAll(t, cb,
		a11ytest.HasRole(accessibility.RoleCheckbox),
		a11ytest.HasLabel("I agree"),
		a11ytest.IsFocusable(),
		a11ytest.IsChecked(),
		a11ytest.IsEnabled(),
	)
}

func TestInputFullAssert(t *testing.T) {
	input := widgets.NewInput()
	input.SetLabel("Username")

	a11ytest.AssertAll(t, input,
		a11ytest.HasRole(accessibility.RoleTextbox),
		a11ytest.HasLabel("Username"),
		a11ytest.IsFocusable(),
		a11ytest.IsEnabled(),
	)
}

func TestTextFullAssert(t *testing.T) {
	txt := widgets.NewText("Hello world")

	a11ytest.AssertAll(t, txt,
		a11ytest.HasRole(accessibility.RoleText),
		a11ytest.HasLabel("Hello world"),
	)
}

// ---------------------------------------------------------------------------
// Helpers: fakeTB for counting errors without failing the real test
// ---------------------------------------------------------------------------

type fakeTB struct {
	testing.TB
	errorCount int
	logs       []string
}

func (f *fakeTB) Helper()                       {}
func (f *fakeTB) Errorf(format string, a ...any) { f.errorCount++ }
func (f *fakeTB) Logf(format string, a ...any)   { f.logs = append(f.logs, format) }
func (f *fakeTB) Fatalf(format string, a ...any) { f.errorCount++ }

// ---------------------------------------------------------------------------
// Mock widgets for state-based tests
// ---------------------------------------------------------------------------

// mockExpandable implements Widget + Accessible with expandable state.
type mockExpandable struct {
	accessibility.Base
	expanded bool
}

func (m *mockExpandable) Measure(c runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 1, Height: 1}
}
func (m *mockExpandable) Layout(b runtime.Rect)                          {}
func (m *mockExpandable) Render(ctx runtime.RenderContext)               {}
func (m *mockExpandable) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}
func (m *mockExpandable) AccessibleState() accessibility.StateSet {
	return accessibility.StateSet{Expanded: &m.expanded}
}

// mockSelectable implements Widget + Accessible with selected state.
type mockSelectable struct {
	accessibility.Base
	selected bool
}

func (m *mockSelectable) Measure(c runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 1, Height: 1}
}
func (m *mockSelectable) Layout(b runtime.Rect)                          {}
func (m *mockSelectable) Render(ctx runtime.RenderContext)               {}
func (m *mockSelectable) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}
func (m *mockSelectable) AccessibleState() accessibility.StateSet {
	return accessibility.StateSet{Selected: m.selected}
}

// mockStateful implements Widget + Accessible with required/invalid states.
type mockStateful struct {
	accessibility.Base
	required bool
	invalid  bool
}

func (m *mockStateful) Measure(c runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 1, Height: 1}
}
func (m *mockStateful) Layout(b runtime.Rect)                          {}
func (m *mockStateful) Render(ctx runtime.RenderContext)               {}
func (m *mockStateful) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}
func (m *mockStateful) AccessibleState() accessibility.StateSet {
	return accessibility.StateSet{Required: m.required, Invalid: m.invalid}
}

// mockWithValue implements Widget + Accessible with a value.
type mockWithValue struct {
	accessibility.Base
	text string
}

func (m *mockWithValue) Measure(c runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 1, Height: 1}
}
func (m *mockWithValue) Layout(b runtime.Rect)                          {}
func (m *mockWithValue) Render(ctx runtime.RenderContext)               {}
func (m *mockWithValue) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}
func (m *mockWithValue) AccessibleValue() *accessibility.ValueInfo {
	return &accessibility.ValueInfo{Text: m.text}
}

// mockContainer implements Widget + ChildProvider for tree tests.
type mockContainer struct {
	accessibility.Base
	children []runtime.Widget
}

func (m *mockContainer) Measure(c runtime.Constraints) runtime.Size {
	return runtime.Size{Width: 10, Height: 10}
}
func (m *mockContainer) Layout(b runtime.Rect) {}
func (m *mockContainer) Render(ctx runtime.RenderContext) {}
func (m *mockContainer) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}
func (m *mockContainer) ChildWidgets() []runtime.Widget {
	return m.children
}
func (m *mockContainer) AccessibleRole() accessibility.Role {
	return accessibility.RoleGroup
}
func (m *mockContainer) AccessibleLabel() string {
	return "container"
}
