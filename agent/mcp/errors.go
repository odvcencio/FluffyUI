package mcp

import "errors"

var (
	// Connection errors.
	ErrNotConnected     = errors.New("not connected to app")
	ErrInvalidTransport = errors.New("invalid transport")
	ErrAppCrashed       = errors.New("app process terminated")

	// Authentication errors.
	ErrAuthRequired   = errors.New("authentication required")
	ErrAuthFailed     = errors.New("authentication failed")
	ErrSessionExpired = errors.New("session expired")

	// Limit errors.
	ErrTooManySessions = errors.New("too many sessions")
	ErrRateLimited     = errors.New("rate limit exceeded")

	// Access errors.
	ErrTextDenied      = errors.New("text access denied")
	ErrClipboardDenied = errors.New("clipboard access denied")

	// Widget errors.
	ErrWidgetNotFound = errors.New("widget not found")
	ErrWidgetDisabled = errors.New("widget is disabled")
	ErrNotFocusable   = errors.New("widget cannot be focused")
	ErrAmbiguousLabel = errors.New("multiple widgets match label")

	// Operation errors.
	ErrTimeout = errors.New("operation timed out")
)
