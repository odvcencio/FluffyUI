package forms

import (
	"fmt"
	"regexp"
	"strings"
)

// Validator validates a field value.
type Validator interface {
	Validate(value any) *ValidationError
}

type validatorFunc func(value any) *ValidationError

func (v validatorFunc) Validate(value any) *ValidationError {
	return v(value)
}

// Required ensures a value is present.
func Required(msg string) Validator {
	message := fallbackMessage(msg, "This field is required.")
	return validatorFunc(func(value any) *ValidationError {
		if value == nil {
			return &ValidationError{Message: message}
		}
		switch v := value.(type) {
		case string:
			if strings.TrimSpace(v) == "" {
				return &ValidationError{Message: message}
			}
		case []string:
			if len(v) == 0 {
				return &ValidationError{Message: message}
			}
		case []any:
			if len(v) == 0 {
				return &ValidationError{Message: message}
			}
		case map[string]any:
			if len(v) == 0 {
				return &ValidationError{Message: message}
			}
		}
		return nil
	})
}

// MinLength enforces a minimum length.
func MinLength(n int, msg string) Validator {
	message := fallbackMessage(msg, fmt.Sprintf("Minimum length is %d.", n))
	return validatorFunc(func(value any) *ValidationError {
		if n < 0 {
			return nil
		}
		if length, ok := lengthOf(value); ok && length < n {
			return &ValidationError{Message: message}
		}
		return nil
	})
}

// MaxLength enforces a maximum length.
func MaxLength(n int, msg string) Validator {
	message := fallbackMessage(msg, fmt.Sprintf("Maximum length is %d.", n))
	return validatorFunc(func(value any) *ValidationError {
		if n < 0 {
			return nil
		}
		if length, ok := lengthOf(value); ok && length > n {
			return &ValidationError{Message: message}
		}
		return nil
	})
}

// Pattern enforces a regex pattern match.
func Pattern(regex string, msg string) Validator {
	re, err := regexp.Compile(regex)
	if err != nil {
		message := fallbackMessage(msg, "Invalid pattern.")
		return validatorFunc(func(value any) *ValidationError {
			return &ValidationError{Message: message}
		})
	}
	message := fallbackMessage(msg, "Value does not match pattern.")
	return validatorFunc(func(value any) *ValidationError {
		if value == nil {
			return nil
		}
		text, ok := value.(string)
		if !ok {
			return nil
		}
		if !re.MatchString(text) {
			return &ValidationError{Message: message}
		}
		return nil
	})
}

// Min enforces a minimum numeric value.
func Min(n float64, msg string) Validator {
	message := fallbackMessage(msg, fmt.Sprintf("Minimum value is %.2f.", n))
	return validatorFunc(func(value any) *ValidationError {
		if num, ok := toFloat(value); ok && num < n {
			return &ValidationError{Message: message}
		}
		return nil
	})
}

// Max enforces a maximum numeric value.
func Max(n float64, msg string) Validator {
	message := fallbackMessage(msg, fmt.Sprintf("Maximum value is %.2f.", n))
	return validatorFunc(func(value any) *ValidationError {
		if num, ok := toFloat(value); ok && num > n {
			return &ValidationError{Message: message}
		}
		return nil
	})
}

// Email validates a simple email format.
func Email(msg string) Validator {
	re := regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	message := fallbackMessage(msg, "Invalid email address.")
	return validatorFunc(func(value any) *ValidationError {
		text, ok := value.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return nil
		}
		if !re.MatchString(text) {
			return &ValidationError{Message: message}
		}
		return nil
	})
}

// URL validates a basic URL format.
func URL(msg string) Validator {
	re := regexp.MustCompile(`^https?://[^\s]+$`)
	message := fallbackMessage(msg, "Invalid URL.")
	return validatorFunc(func(value any) *ValidationError {
		text, ok := value.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return nil
		}
		if !re.MatchString(text) {
			return &ValidationError{Message: message}
		}
		return nil
	})
}

// Numeric ensures the value contains only digits.
func Numeric(msg string) Validator {
	re := regexp.MustCompile(`^\d+$`)
	message := fallbackMessage(msg, "Value must be numeric.")
	return validatorFunc(func(value any) *ValidationError {
		text, ok := value.(string)
		if !ok || text == "" {
			return nil
		}
		if !re.MatchString(text) {
			return &ValidationError{Message: message}
		}
		return nil
	})
}

// AlphaNumeric ensures the value contains only letters and digits.
func AlphaNumeric(msg string) Validator {
	re := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	message := fallbackMessage(msg, "Value must be alphanumeric.")
	return validatorFunc(func(value any) *ValidationError {
		text, ok := value.(string)
		if !ok || text == "" {
			return nil
		}
		if !re.MatchString(text) {
			return &ValidationError{Message: message}
		}
		return nil
	})
}

// Range enforces a numeric value within min and max (inclusive).
func Range(min, max float64, msg string) Validator {
	message := fallbackMessage(msg, fmt.Sprintf("Value must be between %.2f and %.2f.", min, max))
	return validatorFunc(func(value any) *ValidationError {
		num, ok := toFloat(value)
		if !ok {
			return nil
		}
		if num < min || num > max {
			return &ValidationError{Message: message}
		}
		return nil
	})
}

// Length enforces an exact length.
func Length(n int, msg string) Validator {
	message := fallbackMessage(msg, fmt.Sprintf("Length must be exactly %d.", n))
	return validatorFunc(func(value any) *ValidationError {
		if n < 0 {
			return nil
		}
		if length, ok := lengthOf(value); ok && length != n {
			return &ValidationError{Message: message}
		}
		return nil
	})
}

// Custom wraps a validator function.
func Custom(fn func(any) *ValidationError) Validator {
	if fn == nil {
		return validatorFunc(func(any) *ValidationError { return nil })
	}
	return validatorFunc(fn)
}

// All requires all validators to pass.
func All(validators ...Validator) Validator {
	return validatorFunc(func(value any) *ValidationError {
		for _, v := range validators {
			if v == nil {
				continue
			}
			if err := v.Validate(value); err != nil {
				return err
			}
		}
		return nil
	})
}

// Any passes if any validator passes.
func Any(validators ...Validator) Validator {
	return validatorFunc(func(value any) *ValidationError {
		var firstErr *ValidationError
		for _, v := range validators {
			if v == nil {
				continue
			}
			if err := v.Validate(value); err == nil {
				return nil
			} else if firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	})
}

// When conditionally applies a validator.
func When(cond func() bool, v Validator) Validator {
	return validatorFunc(func(value any) *ValidationError {
		if cond == nil || !cond() || v == nil {
			return nil
		}
		return v.Validate(value)
	})
}

func fallbackMessage(msg string, fallback string) string {
	if strings.TrimSpace(msg) == "" {
		return fallback
	}
	return msg
}

func lengthOf(value any) (int, bool) {
	switch v := value.(type) {
	case string:
		return len(v), true
	case []string:
		return len(v), true
	case []any:
		return len(v), true
	case []int:
		return len(v), true
	case []float64:
		return len(v), true
	default:
		return 0, false
	}
}

func toFloat(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}
