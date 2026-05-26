package fluffy

import (
	"fmt"
	"image"

	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/state"
	"m31labs.dev/fluffyui/style"
	"m31labs.dev/fluffyui/theme"
	"m31labs.dev/fluffyui/widgets"
)

// =============================================================================
// RUNTIME RE-EXPORTS
// =============================================================================

type (
	Widget          = runtime.Widget
	Message         = runtime.Message
	Constraints     = runtime.Constraints
	Size            = runtime.Size
	Rect            = runtime.Rect
	RenderContext   = runtime.RenderContext
	HandleResult    = runtime.HandleResult
	Command         = runtime.Command
	App             = runtime.App
	Services        = runtime.Services
	Persistable     = runtime.Persistable
	PersistSnapshot = runtime.PersistSnapshot
	KeyMsg          = runtime.KeyMsg
	MouseMsg        = runtime.MouseMsg
	ResizeMsg       = runtime.ResizeMsg
	PasteMsg        = runtime.PasteMsg
	TickMsg         = runtime.TickMsg
	CustomMsg       = runtime.CustomMsg

	FocusRegistrationMode = runtime.FocusRegistrationMode
)

var (
	Handled      = runtime.Handled
	Unhandled    = runtime.Unhandled
	WithCommand  = runtime.WithCommand
	WithCommands = runtime.WithCommands
	CaptureState = runtime.CaptureState
	ApplyState   = runtime.ApplyState
	SaveSnapshot = runtime.SaveSnapshot
	LoadSnapshot = runtime.LoadSnapshot
)

// =============================================================================
// STATE RE-EXPORTS (Simplified)
// =============================================================================

type (
	Subscribable     = state.Subscribable
	EqualFunc[T any] = state.EqualFunc[T]
)

// Signal creates a new reactive signal with an initial value.
// This is the simplest way to create reactive state:
//
//	count := fluffy.Signal(0)
//	count.Set(count.Get() + 1)
func Signal[T any](initial T) *state.Signal[T] {
	return state.NewSignal(initial)
}

// Computed creates a derived value that automatically recalculates
// when any dependency changes. If no deps are provided, dependencies
// are detected automatically by tracking signal reads.
//
//	count := fluffy.Signal(0)
//	doubled := fluffy.Computed(func() int { return count.Get() * 2 }, count)
func Computed[T any](fn func() T, deps ...state.Subscribable) *state.Computed[T] {
	return state.NewComputed(fn, deps...)
}

// NewSignal creates a signal with smart defaults for equality checking.
// Prefer Signal() for new code.
func NewSignal[T any](initial T) *state.Signal[T] {
	return state.NewSignal(initial)
}

// NewComputed creates a computed value from dependencies.
// Prefer Computed() for new code.
func NewComputed[T any](compute func() T, deps ...state.Subscribable) *state.Computed[T] {
	return state.NewComputed(compute, deps...)
}

// =============================================================================
// STYLE HELPERS
// =============================================================================

// DefaultStyle returns the base terminal style (no colors, no attributes).
func DefaultStyle() backend.Style {
	return backend.DefaultStyle()
}

// DefaultStylesheet returns the built-in stylesheet with standard widget styles.
func DefaultStylesheet() *style.Stylesheet {
	return theme.DefaultStylesheet()
}

// =============================================================================
// SIMPLIFIED WIDGET CONSTRUCTORS
// =============================================================================

// Label creates a text label.
func NewLabel(text string) *widgets.Label {
	return widgets.NewLabel(text)
}

// Label creates a text label.
func Label(text string, opts ...widgets.LabelOption) *widgets.Label {
	return widgets.NewLabel(text, opts...)
}

// Labelf creates a formatted text label.
func Labelf(format string, args ...any) *widgets.Label {
	return widgets.NewLabel(fmt.Sprintf(format, args...))
}

// Text creates a styled text widget.
func NewText(text string) *widgets.Text {
	return widgets.NewText(text)
}

// Text creates a styled text widget.
func Text(text string, opts ...widgets.TextOption) *widgets.Text {
	return widgets.NewText(text, opts...)
}

// Button creates a button with optional configuration.
func NewButton(label string, opts ...widgets.ButtonOption) *widgets.Button {
	return widgets.NewButton(label, opts...)
}

// PrimaryButton creates a primary-styled button.
func PrimaryButton(label string, onClick func()) *widgets.Button {
	return widgets.NewButton(label, WithVariant(VariantPrimary), WithOnClick(onClick))
}

// SecondaryButton creates a secondary-styled button.
func SecondaryButton(label string, onClick func()) *widgets.Button {
	return widgets.NewButton(label, WithVariant(VariantSecondary), WithOnClick(onClick))
}

// DangerButton creates a danger-styled button.
func DangerButton(label string, onClick func()) *widgets.Button {
	return widgets.NewButton(label, WithVariant(VariantDanger), WithOnClick(onClick))
}

// Input creates a text input.
func NewInput() *widgets.Input {
	return widgets.NewInput()
}

// TextArea creates a multi-line text area.
func NewTextArea() *widgets.TextArea {
	return widgets.NewTextArea()
}

// Checkbox creates a checkbox.
func NewCheckbox(label string) *widgets.Checkbox {
	return widgets.NewCheckbox(label)
}

// =============================================================================
// BASE TYPES (Progressive Capability)
// =============================================================================

// Base is the minimal widget base (no auto-wiring).
type Base = widgets.Base

// Component is the reactive base with manual subscriptions.
type Component = widgets.Component

// =============================================================================
// CONTAINER WIDGETS
// =============================================================================

// NewPanel wraps a child widget in a bordered panel with optional title.
func NewPanel(child runtime.Widget) *widgets.Panel {
	return widgets.NewPanel(child)
}

// NewGrid creates a rows-by-cols grid layout. Add children with Grid.Add.
func NewGrid(rows, cols int) *widgets.Grid {
	return widgets.NewGrid(rows, cols)
}

// NewStack layers children on top of each other (last child drawn on top).
func NewStack(children ...runtime.Widget) *widgets.Stack {
	return widgets.NewStack(children...)
}

// NewScrollView wraps content in a scrollable viewport with scrollbars.
func NewScrollView(content runtime.Widget) *widgets.ScrollView {
	return widgets.NewScrollView(content)
}

// NewAspectRatio constrains child to the given width:height ratio (e.g. 16.0/9.0).
func NewAspectRatio(child runtime.Widget, ratio float64) *widgets.AspectRatio {
	return widgets.NewAspectRatio(child, ratio)
}

// =============================================================================
// BUTTON OPTIONS
// =============================================================================

type ButtonOption = widgets.ButtonOption
type Variant = widgets.Variant

const (
	VariantPrimary   = widgets.VariantPrimary
	VariantSecondary = widgets.VariantSecondary
	VariantDanger    = widgets.VariantDanger
)

// WithVariant sets the button's visual variant (Primary, Secondary, or Danger).
func WithVariant(variant Variant) ButtonOption {
	return widgets.WithVariant(variant)
}

// WithDisabled binds a signal that controls whether the button is disabled.
func WithDisabled(disabled *state.Signal[bool]) ButtonOption {
	return widgets.WithDisabled(disabled)
}

// WithLoading binds a signal that shows a spinner and disables the button.
func WithLoading(loading *state.Signal[bool]) ButtonOption {
	return widgets.WithLoading(loading)
}

// WithOnClick sets the callback invoked when the button is activated.
func WithOnClick(fn func()) ButtonOption {
	return widgets.WithOnClick(fn)
}

// =============================================================================
// ADVANCED WIDGETS
// =============================================================================

type (
	Table        = widgets.Table
	Tree         = widgets.Tree
	Dialog       = widgets.Dialog
	Tabs         = widgets.Tabs
	Select       = widgets.Select
	Slider       = widgets.Slider
	Spinner      = widgets.Spinner
	DebugOverlay = widgets.DebugOverlay
	AsyncImage   = widgets.AsyncImage
)

type (
	SelectOption       = widgets.SelectOption
	SliderOption       = widgets.SliderOption
	DebugOverlayOption = widgets.DebugOverlayOption
	AsyncImageOption   = widgets.AsyncImageOption
)

// =============================================================================
// IMAGE WIDGETS
// =============================================================================

// NewAsyncImage loads an image from path in a background goroutine and renders
// it using Unicode half-block characters. Supports PNG, JPEG, and GIF.
func NewAsyncImage(path string, opts ...widgets.AsyncImageOption) *widgets.AsyncImage {
	return widgets.NewAsyncImage(path, opts...)
}

// NewAsyncImageWithLoader creates an image widget that uses a custom loader
// function instead of reading from a file path.
func NewAsyncImageWithLoader(loader func() (image.Image, error), opts ...widgets.AsyncImageOption) *widgets.AsyncImage {
	return widgets.NewAsyncImageWithLoader(loader, opts...)
}

// =============================================================================
// DEBUG WIDGETS
// =============================================================================

// NewDebugOverlay wraps root with a toggleable overlay showing widget bounds,
// focus state, and render statistics. Press F12 to toggle.
func NewDebugOverlay(root runtime.Widget, opts ...widgets.DebugOverlayOption) *widgets.DebugOverlay {
	return widgets.NewDebugOverlay(root, opts...)
}
