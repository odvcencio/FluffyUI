package fluffy

import (
	"image"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/state"
	"github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/theme"
	"github.com/odvcencio/fluffyui/widgets"
)

// Common runtime re-exports.
type Widget = runtime.Widget
type Message = runtime.Message
type Constraints = runtime.Constraints
type Size = runtime.Size
type Rect = runtime.Rect
type RenderContext = runtime.RenderContext
type HandleResult = runtime.HandleResult
type Command = runtime.Command
type App = runtime.App
type Services = runtime.Services
type Persistable = runtime.Persistable
type PersistSnapshot = runtime.PersistSnapshot

type KeyMsg = runtime.KeyMsg
type MouseMsg = runtime.MouseMsg

type FocusRegistrationMode = runtime.FocusRegistrationMode

var (
	Handled   = runtime.Handled
	Unhandled = runtime.Unhandled
	WithCommand  = runtime.WithCommand
	WithCommands = runtime.WithCommands
	CaptureState = runtime.CaptureState
	ApplyState   = runtime.ApplyState
	SaveSnapshot = runtime.SaveSnapshot
	LoadSnapshot = runtime.LoadSnapshot
)

// State re-exports.
type Signal[T any] = state.Signal[T]
type Computed[T any] = state.Computed[T]

type Subscribable = state.Subscribable

type EqualFunc[T any] = state.EqualFunc[T]

func NewSignal[T any](initial T) *state.Signal[T] {
	return state.NewSignal(initial)
}

func NewComputed[T any](compute func() T, deps ...state.Subscribable) *state.Computed[T] {
	return state.NewComputed(compute, deps...)
}

// Styles.
func DefaultStyle() backend.Style {
	return backend.DefaultStyle()
}

func DefaultStylesheet() *style.Stylesheet {
	return theme.DefaultStylesheet()
}

// Widget constructors.
func NewLabel(text string) *widgets.Label {
	return widgets.NewLabel(text)
}

// Shorthand constructors for declarative layouts.
func Label(text string) *widgets.Label {
	return widgets.NewLabel(text)
}

func NewText(text string) *widgets.Text {
	return widgets.NewText(text)
}

func Text(text string) *widgets.Text {
	return widgets.NewText(text)
}

func NewButton(label string, opts ...widgets.ButtonOption) *widgets.Button {
	return widgets.NewButton(label, opts...)
}

func Button(label string, opts ...widgets.ButtonOption) *widgets.Button {
	return widgets.NewButton(label, opts...)
}

func NewInput() *widgets.Input {
	return widgets.NewInput()
}

func Input() *widgets.Input {
	return widgets.NewInput()
}

func NewTextArea() *widgets.TextArea {
	return widgets.NewTextArea()
}

func TextArea() *widgets.TextArea {
	return widgets.NewTextArea()
}

func NewGrid(rows, cols int) *widgets.Grid {
	return widgets.NewGrid(rows, cols)
}

func Grid(rows, cols int) *widgets.Grid {
	return widgets.NewGrid(rows, cols)
}

func NewPanel(child runtime.Widget) *widgets.Panel {
	return widgets.NewPanel(child)
}

func Panel(child runtime.Widget) *widgets.Panel {
	return widgets.NewPanel(child)
}

func NewStack(children ...runtime.Widget) *widgets.Stack {
	return widgets.NewStack(children...)
}

func Overlay(children ...runtime.Widget) *widgets.Stack {
	return widgets.NewStack(children...)
}

func NewScrollView(content runtime.Widget) *widgets.ScrollView {
	return widgets.NewScrollView(content)
}

func ScrollView(content runtime.Widget) *widgets.ScrollView {
	return widgets.NewScrollView(content)
}

func NewAspectRatio(child runtime.Widget, ratio float64) *widgets.AspectRatio {
	return widgets.NewAspectRatio(child, ratio)
}

type AspectRatio = widgets.AspectRatio

func NewDebugOverlay(root runtime.Widget, opts ...widgets.DebugOverlayOption) *widgets.DebugOverlay {
	return widgets.NewDebugOverlay(root, opts...)
}

func NewAsyncImage(path string, opts ...widgets.AsyncImageOption) *widgets.AsyncImage {
	return widgets.NewAsyncImage(path, opts...)
}

func NewAsyncImageWithLoader(loader func() (image.Image, error), opts ...widgets.AsyncImageOption) *widgets.AsyncImage {
	return widgets.NewAsyncImageWithLoader(loader, opts...)
}

type AsyncImageOption = widgets.AsyncImageOption
type DebugOverlayOption = widgets.DebugOverlayOption

// Button option re-exports.
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

// Layout helpers.
func Stack(children ...runtime.Widget) *widgets.Stack {
	return widgets.NewStack(children...)
}
