package widgets

import "github.com/odvcencio/fluffyui/runtime"

// Flex re-exports runtime flex layout types for widget-level usage.
type Flex = runtime.Flex
type FlexChild = runtime.FlexChild
type FlexDirection = runtime.FlexDirection

const (
	Column = runtime.Column
	Row    = runtime.Row
)

// VBox creates a vertical flex container.
func VBox(children ...runtime.FlexChild) *runtime.Flex {
	return runtime.VBox(children...)
}

// HBox creates a horizontal flex container.
func HBox(children ...runtime.FlexChild) *runtime.Flex {
	return runtime.HBox(children...)
}

// Flex child helpers.
func Fixed(w runtime.Widget) runtime.FlexChild { return runtime.Fixed(w) }
func Flexible(w runtime.Widget, grow float64) runtime.FlexChild { return runtime.Flexible(w, grow) }
func Expanded(w runtime.Widget) runtime.FlexChild { return runtime.Expanded(w) }
func Sized(w runtime.Widget, basis int) runtime.FlexChild { return runtime.Sized(w, basis) }
func Space() runtime.FlexChild { return runtime.Space() }
func FixedSpace(size int) runtime.FlexChild { return runtime.FixedSpace(size) }
