package widgets

import (
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/forms"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/testing/a11ytest"
)

func TestForm_Basic(t *testing.T) {
	nameInput := NewInput()
	emailInput := NewInput()

	form := NewForm(
		FormField("Name", nameInput),
		FormField("Email", emailInput),
	)

	if form.FieldCount() != 2 {
		t.Errorf("FieldCount() = %d, want 2", form.FieldCount())
	}

	if form.StyleType() != "Form" {
		t.Errorf("StyleType() = %q, want %q", form.StyleType(), "Form")
	}

	// Test SetLabel
	form.SetLabel("Login Form")
	if form.Base.Label != "Login Form" {
		t.Errorf("Label = %q, want %q", form.Base.Label, "Login Form")
	}

	// Test SetSubmitLabel
	form.SetSubmitLabel("Sign In")
	if form.submitLabel != "Sign In" {
		t.Errorf("submitLabel = %q, want %q", form.submitLabel, "Sign In")
	}

	// Test SetCancelLabel
	form.SetCancelLabel("Back")
	if form.cancelLabel != "Back" {
		t.Errorf("cancelLabel = %q, want %q", form.cancelLabel, "Back")
	}
}

func TestForm_Validation(t *testing.T) {
	nameInput := NewInput()
	nameInput.SetValidators(forms.Required("Name is required"))

	emailInput := NewInput()
	emailInput.SetValidators(forms.Required("Email is required"))

	form := NewForm(
		FormField("Name", nameInput),
		FormField("Email", emailInput),
	)

	// Both empty, should fail validation
	if form.Validate() {
		t.Error("Validate() returned true with empty required fields")
	}

	errs := form.Errors()
	if len(errs) != 2 {
		t.Errorf("Errors() returned %d errors, want 2", len(errs))
	}
	if errs[0] != "Name is required" {
		t.Errorf("Error[0] = %q, want %q", errs[0], "Name is required")
	}
	if errs[1] != "Email is required" {
		t.Errorf("Error[1] = %q, want %q", errs[1], "Email is required")
	}

	// Fill in name, email still empty
	nameInput.SetText("Alice")
	if form.Validate() {
		t.Error("Validate() returned true with one empty required field")
	}
	errs = form.Errors()
	if len(errs) != 1 {
		t.Errorf("Errors() returned %d errors, want 1", len(errs))
	}

	// Fill both
	emailInput.SetText("alice@example.com")
	if !form.Validate() {
		t.Error("Validate() returned false with all fields filled")
	}
	errs = form.Errors()
	if len(errs) != 0 {
		t.Errorf("Errors() returned %d errors, want 0", len(errs))
	}
}

func TestForm_Render(t *testing.T) {
	nameInput := NewInput()
	emailInput := NewInput()

	form := NewForm(
		FormField("Name", nameInput),
		FormField("Email", emailInput),
	)

	constraints := runtime.Constraints{
		MinWidth: 0, MaxWidth: 40,
		MinHeight: 0, MaxHeight: 20,
	}
	size := form.Measure(constraints)

	if size.Width <= 0 {
		t.Errorf("Measure width = %d, want > 0", size.Width)
	}
	if size.Height <= 0 {
		t.Errorf("Measure height = %d, want > 0", size.Height)
	}

	// Layout the form
	form.Layout(runtime.Rect{X: 0, Y: 0, Width: 40, Height: 20})

	// Verify labels exist as children
	children := form.ChildWidgets()
	if len(children) == 0 {
		t.Fatal("ChildWidgets() returned empty list")
	}

	// Check that at least first child is the Name label
	if label, ok := children[0].(*Label); ok {
		if label.text != "Name:" {
			t.Errorf("First label text = %q, want %q", label.text, "Name:")
		}
	} else {
		t.Errorf("First child is %T, want *Label", children[0])
	}
}

func TestForm_TabNavigation(t *testing.T) {
	nameInput := NewInput()
	emailInput := NewInput()

	form := NewForm(
		FormField("Name", nameInput),
		FormField("Email", emailInput),
	)

	// Initially focusIdx is 0
	if form.focusIdx != 0 {
		t.Errorf("initial focusIdx = %d, want 0", form.focusIdx)
	}

	// Tab should move to next field
	tabMsg := runtime.KeyMsg{Key: terminal.KeyTab}
	result := form.HandleMessage(tabMsg)
	if !result.Handled {
		t.Error("Tab was not handled")
	}
	if form.focusIdx != 1 {
		t.Errorf("after Tab focusIdx = %d, want 1", form.focusIdx)
	}
	if !emailInput.IsFocused() {
		t.Error("Email input should be focused after Tab")
	}

	// Tab again should wrap to first field
	form.HandleMessage(tabMsg)
	if form.focusIdx != 0 {
		t.Errorf("after second Tab focusIdx = %d, want 0", form.focusIdx)
	}
	if !nameInput.IsFocused() {
		t.Error("Name input should be focused after wrapping Tab")
	}

	// Shift+Tab should go backwards
	shiftTabMsg := runtime.KeyMsg{Key: terminal.KeyTab, Shift: true}
	form.HandleMessage(shiftTabMsg)
	if form.focusIdx != 1 {
		t.Errorf("after Shift+Tab focusIdx = %d, want 1", form.focusIdx)
	}
}

func TestForm_Accessibility(t *testing.T) {
	nameInput := NewInput()
	form := NewForm(
		FormField("Name", nameInput),
	)

	a11ytest.Assert(t, form,
		a11ytest.HasRole(accessibility.RoleForm),
		a11ytest.HasLabel("Form"),
	)

	// Custom label
	form.SetLabel("Registration")
	a11ytest.Assert(t, form,
		a11ytest.HasLabel("Registration"),
	)
}

func TestForm_Submit(t *testing.T) {
	nameInput := NewInput()
	nameInput.SetText("Alice")

	var submitted map[string]string
	form := NewForm(
		FormField("Name", nameInput),
	)
	form.SetOnSubmit(func(values map[string]string) {
		submitted = values
	})

	// Submit via Enter key
	enterMsg := runtime.KeyMsg{Key: terminal.KeyEnter}
	form.HandleMessage(enterMsg)

	if submitted == nil {
		t.Fatal("onSubmit was not called")
	}
	if submitted["Name"] != "Alice" {
		t.Errorf("submitted Name = %q, want %q", submitted["Name"], "Alice")
	}
}

func TestForm_Cancel(t *testing.T) {
	nameInput := NewInput()

	cancelled := false
	form := NewForm(
		FormField("Name", nameInput),
	)
	form.SetOnCancel(func() {
		cancelled = true
	})

	escMsg := runtime.KeyMsg{Key: terminal.KeyEscape}
	form.HandleMessage(escMsg)

	if !cancelled {
		t.Error("onCancel was not called")
	}
}

func TestForm_SubmitBlockedByValidation(t *testing.T) {
	nameInput := NewInput()
	nameInput.SetValidators(forms.Required("Name is required"))

	var submitted map[string]string
	form := NewForm(
		FormField("Name", nameInput),
	)
	form.SetOnSubmit(func(values map[string]string) {
		submitted = values
	})

	// Submit with empty required field should not call onSubmit
	enterMsg := runtime.KeyMsg{Key: terminal.KeyEnter}
	form.HandleMessage(enterMsg)

	if submitted != nil {
		t.Error("onSubmit should not have been called when validation fails")
	}
}

func TestForm_NilSafety(t *testing.T) {
	var form *Form

	// None of these should panic
	form.SetOnSubmit(nil)
	form.SetOnCancel(nil)
	form.SetSubmitLabel("Submit")
	form.SetCancelLabel("Cancel")
	form.SetLabel("Form")

	if form.FieldCount() != 0 {
		t.Errorf("nil Form.FieldCount() = %d, want 0", form.FieldCount())
	}
	if !form.Validate() {
		t.Error("nil Form.Validate() should return true")
	}
	if form.Errors() != nil {
		t.Error("nil Form.Errors() should return nil")
	}
}
