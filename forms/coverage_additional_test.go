package forms

import "testing"

func TestFieldBaseStateHelpers(t *testing.T) {
	base := NewFieldBase("name", "", Required("required"))
	if base.Name() != "name" {
		t.Fatalf("expected name")
	}
	base.MarkTouched()
	if !base.Touched() {
		t.Fatalf("expected touched")
	}
	base.UpdateDirty("value")
	if !base.Dirty() {
		t.Fatalf("expected dirty")
	}
	base.SetErrors([]string{"err"})
	if len(base.Errors()) != 1 {
		t.Fatalf("expected error")
	}
	base.ResetState()
	if base.Dirty() || base.Touched() || len(base.Errors()) != 0 {
		t.Fatalf("expected state reset")
	}
}

func TestFormSignalsAndCancel(t *testing.T) {
	field := NewField("name", "")
	form := NewForm(field)
	if form.DirtySignal().Get() {
		t.Fatalf("expected clean form")
	}
	form.Set("name", "value")
	if !form.DirtySignal().Get() {
		t.Fatalf("expected dirty form")
	}
	form.Reset()
	if form.DirtySignal().Get() {
		t.Fatalf("expected reset to clean")
	}
	called := false
	form.OnCancel(func() { called = true })
	form.Cancel()
	if !called {
		t.Fatalf("expected cancel callback")
	}
}

func TestFormValidators(t *testing.T) {
	form := NewForm(NewField("a", "1"), NewField("b", "2"))
	form.AddValidator(formValidatorFunc(func(values Values) []ValidationError {
		if values == nil || values["b"] == "" {
			return []ValidationError{{Field: "b", Message: "required"}}
		}
		return nil
	}))
	if len(form.Validate()) != 0 {
		t.Fatalf("expected no errors")
	}
	form.Set("b", "")
	errs := form.Validate()
	if len(errs) == 0 {
		t.Fatalf("expected validation errors")
	}
	if !form.ValidSignal().Get() {
		// valid signal should already be updated when errors exist
	} else {
		t.Fatalf("expected invalid signal")
	}
}
