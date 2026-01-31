package widgets

import (
	"strings"

	"github.com/odvcencio/fluffyui/forms"
)

func validateValue(value any, validators []forms.Validator) ([]forms.ValidationError, []string) {
	if len(validators) == 0 {
		return nil, nil
	}
	var errs []forms.ValidationError
	for _, validator := range validators {
		if validator == nil {
			continue
		}
		if err := validator.Validate(value); err != nil {
			errs = append(errs, *err)
		}
	}
	if len(errs) == 0 {
		return nil, nil
	}
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		msg := strings.TrimSpace(err.Message)
		if msg == "" {
			continue
		}
		messages = append(messages, msg)
	}
	return errs, messages
}
