package forms

import (
	"context"
	"testing"
	"time"
)

func TestSimpleFieldAsyncValidation(t *testing.T) {
	field := NewField("status", "", Required("required"))
	field.SetAsyncValidators(Async(func(ctx context.Context, value any) *ValidationError {
		time.Sleep(10 * time.Millisecond)
		if value == "ok" {
			return nil
		}
		return &ValidationError{Message: "async fail"}
	}))

	field.SetValue("bad")
	ch := field.ValidateAsync(context.Background())
	errs := <-ch
	if len(errs) == 0 {
		t.Fatalf("expected async validation errors")
	}
	if field.Validating() {
		t.Fatalf("expected validating to be false after completion")
	}
	if len(field.Errors()) == 0 {
		t.Fatalf("expected field errors populated")
	}
}

func TestSimpleFieldAsyncValidationStale(t *testing.T) {
	field := NewField("status", "")
	field.SetAsyncValidators(Async(func(ctx context.Context, value any) *ValidationError {
		time.Sleep(30 * time.Millisecond)
		if value == "bad" {
			return &ValidationError{Message: "async fail"}
		}
		return nil
	}))

	field.SetValue("bad")
	ch := field.ValidateAsync(context.Background())
	field.SetValue("ok")

	<-ch
	if len(field.Errors()) != 0 {
		t.Fatalf("expected stale async errors to be ignored")
	}
}

func TestFormDependenciesAndAsyncValidation(t *testing.T) {
	fieldA := NewField("a", "one")
	fieldB := NewField("b", "one", validatorFunc(func(value any) *ValidationError {
		if fieldA.Value() != value {
			return &ValidationError{Field: "b", Message: "mismatch"}
		}
		return nil
	}))

	form := NewForm(fieldA, fieldB)
	form.SetDependencies("b", "a")

	form.Set("a", "two")
	if len(fieldB.Errors()) == 0 {
		t.Fatalf("expected dependent field to revalidate")
	}

	form.AddAsyncValidator(AsyncForm(func(ctx context.Context, values Values) []ValidationError {
		if values["a"] == "two" {
			return []ValidationError{{Field: "a", Message: "async fail"}}
		}
		return nil
	}))

	errs := <-form.ValidateAsync(context.Background())
	if len(errs) == 0 {
		t.Fatalf("expected async form validation errors")
	}
}
