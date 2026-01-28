package fluffy

import (
	"fmt"
	"image"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/theme"
	"github.com/odvcencio/fluffyui/widgets"
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
	Signal[T any]    = state.Signal[T]
	Computed[T any]  = state.Computed[T]
	Subscribable     = state.Subscribable
	EqualFunc[T any] = state.EqualFunc[T]
)

// NewSignal creates a signal with smart defaults for equality checking.
func NewSignal[T any](initial T) *state.Signal[T] {
	return state.NewSignal(initial)
}

// NewComputed creates a computed value from dependencies.
func NewComputed[T any](compute func() T, deps ...state.Subscribable) *state.Computed[T] {
	return state.NewComputed(compute, deps...)
}

// =============================================================================
// STYLE HELPERS
// =============================================================================

func DefaultStyle() backend.Style {
	return backend.DefaultStyle()
}

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

// Labelf creates a formatted text label.
func Labelf(format string, args ...any) *widgets.Label {
	return widgets.NewLabel(fmt.Sprintf(format, args...))
}

// Text creates a styled text widget.
func NewText(text string) *widgets.Text {
	return widgets.NewText(text)
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

func NewPanel(child runtime.Widget) *widgets.Panel {
	return widgets.NewPanel(child)
}

func NewGrid(rows, cols int) *widgets.Grid {
	return widgets.NewGrid(rows, cols)
}

func NewStack(children ...runtime.Widget) *widgets.Stack {
	return widgets.NewStack(children...)
}

func NewScrollView(content runtime.Widget) *widgets.ScrollView {
	return widgets.NewScrollView(content)
}

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

func WithVariant(variant Variant) ButtonOption {
	return widgets.WithVariant(variant)
}

func WithDisabled(disabled *state.Signal[bool]) ButtonOption {
	return widgets.WithDisabled(disabled)
}

func WithLoading(loading *state.Signal[bool]) ButtonOption {
	return widgets.WithLoading(loading)
}

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

func NewAsyncImage(path string, opts ...widgets.AsyncImageOption) *widgets.AsyncImage {
	return widgets.NewAsyncImage(path, opts...)
}

func NewAsyncImageWithLoader(loader func() (image.Image, error), opts ...widgets.AsyncImageOption) *widgets.AsyncImage {
	return widgets.NewAsyncImageWithLoader(loader, opts...)
}

// =============================================================================
// DEBUG WIDGETS
// =============================================================================

func NewDebugOverlay(root runtime.Widget, opts ...widgets.DebugOverlayOption) *widgets.DebugOverlay {
	return widgets.NewDebugOverlay(root, opts...)
}
