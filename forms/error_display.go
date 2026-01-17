package forms

import "github.com/odvcencio/fluffy-ui/backend"

// ErrorDisplay describes how to render an error for a field.
type ErrorDisplay struct {
	Field    Field
	Position ErrorPosition
	Style    backend.Style
}

// ErrorPosition defines error placement.
type ErrorPosition int

const (
	ErrorBelow ErrorPosition = iota
	ErrorInline
	ErrorSummary
)
