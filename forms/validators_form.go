package forms

import (
	"context"
	"strings"
)

// FormValidator validates multiple fields at once.
type FormValidator interface {
	Validate(values Values) []ValidationError
}

type formValidatorFunc func(values Values) []ValidationError

func (f formValidatorFunc) Validate(values Values) []ValidationError {
	return f(values)
}

// AsyncFormValidator validates multiple fields asynchronously.
type AsyncFormValidator interface {
	ValidateAsync(ctx context.Context, values Values) []ValidationError
}

type asyncFormValidatorFunc func(ctx context.Context, values Values) []ValidationError

func (f asyncFormValidatorFunc) ValidateAsync(ctx context.Context, values Values) []ValidationError {
	return f(ctx, values)
}

// AsyncForm wraps a function into an AsyncFormValidator.
func AsyncForm(fn func(ctx context.Context, values Values) []ValidationError) AsyncFormValidator {
	if fn == nil {
		return nil
	}
	return asyncFormValidatorFunc(fn)
}

// FieldsMatch validates that two fields are equal.
func FieldsMatch(a, b, msg string) FormValidator {
	fieldA := strings.TrimSpace(a)
	fieldB := strings.TrimSpace(b)
	message := msg
	if strings.TrimSpace(message) == "" {
		message = "Fields do not match."
	}
	return formValidatorFunc(func(values Values) []ValidationError {
		if values == nil {
			return nil
		}
		va, okA := values[fieldA]
		vb, okB := values[fieldB]
		if !okA || !okB {
			return nil
		}
		if va != vb {
			return []ValidationError{
				{Field: fieldB, Message: message},
			}
		}
		return nil
	})
}
