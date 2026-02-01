package forms

import "context"

// AsyncValidator validates a field value asynchronously.
type AsyncValidator interface {
	ValidateAsync(ctx context.Context, value any) *ValidationError
}

type asyncValidatorFunc func(ctx context.Context, value any) *ValidationError

func (v asyncValidatorFunc) ValidateAsync(ctx context.Context, value any) *ValidationError {
	return v(ctx, value)
}

// Async wraps a function into an AsyncValidator.
func Async(fn func(ctx context.Context, value any) *ValidationError) AsyncValidator {
	if fn == nil {
		return nil
	}
	return asyncValidatorFunc(fn)
}
