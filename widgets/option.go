package widgets

// Option defines a configurable option for a widget.
type Option[T any] func(*T)
