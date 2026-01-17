package forms

import "strings"

// FormValidator validates multiple fields at once.
type FormValidator interface {
	Validate(values Values) []ValidationError
}

type formValidatorFunc func(values Values) []ValidationError

func (f formValidatorFunc) Validate(values Values) []ValidationError {
	return f(values)
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
