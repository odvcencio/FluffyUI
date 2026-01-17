package forms

import "testing"

func TestFieldValidation(t *testing.T) {
	field := NewField("name", "", Required("required"))
	errs := field.Validate()
	if len(errs) == 0 {
		t.Fatalf("expected validation error")
	}
	field.SetValue("ok")
	errs = field.Validate()
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d", len(errs))
	}
}

func TestFormSubmit(t *testing.T) {
	field := NewField("email", "", All(Required("required"), Email("invalid")))
	form := NewForm(field)
	called := false
	form.OnSubmit(func(values Values) {
		called = true
	})
	form.Submit()
	if called {
		t.Fatalf("expected submit to be blocked by validation")
	}
	field.SetValue("test@example.com")
	form.Submit()
	if !called {
		t.Fatalf("expected submit to be called")
	}
}

func TestFieldsMatch(t *testing.T) {
	form := NewForm(
		NewField("password", "abc"),
		NewField("confirmPassword", "def"),
	)
	form.AddValidator(FieldsMatch("password", "confirmPassword", "mismatch"))
	errs := form.Validate()
	if len(errs) == 0 {
		t.Fatalf("expected mismatch error")
	}
	form.Set("confirmPassword", "abc")
	errs = form.Validate()
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got %d", len(errs))
	}
}
