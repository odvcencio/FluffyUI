package widgets

import "github.com/odvcencio/fluffyui/runtime"

// Flex re-exports runtime flex layout types for widget-level usage.
type Flex = runtime.Flex
type FlexChild = runtime.FlexChild
type FlexDirection = runtime.FlexDirection

// FlexDirection constants.
const (
	FlexColumn = runtime.Column
	FlexRow    = runtime.Row
)

// VBox creates a vertical flex container.
func VBox(children ...runtime.FlexChild) *runtime.Flex {
	return runtime.VBox(children...)
}

// HBox creates a horizontal flex container.
func HBox(children ...runtime.FlexChild) *runtime.Flex {
	return runtime.HBox(children...)
}

// Flex child helpers - these delegate to runtime.
func FlexFixed(w runtime.Widget) runtime.FlexChild { return runtime.Fixed(w) }
func FlexFlexible(w runtime.Widget, grow float64) runtime.FlexChild { return runtime.Flexible(w, grow) }
func FlexExpanded(w runtime.Widget) runtime.FlexChild { return runtime.Expanded(w) }
func FlexSized(w runtime.Widget, basis int) runtime.FlexChild { return runtime.Sized(w, basis) }
func FlexSpace() runtime.FlexChild { return runtime.Space() }
func FlexFixedSpace(size int) runtime.FlexChild { return runtime.FixedSpace(size) }
