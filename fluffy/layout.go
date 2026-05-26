package fluffy

import "m31labs.dev/fluffyui/runtime"

// Flex is a container that lays out children along a horizontal or vertical axis.
type Flex = runtime.Flex

// FlexChild wraps a widget with flex layout properties (grow, shrink, basis).
type FlexChild = runtime.FlexChild

// FlexDirection specifies the main axis: Row (horizontal) or Column (vertical).
type FlexDirection = runtime.FlexDirection

const (
	Row    = runtime.Row
	Column = runtime.Column
)

// VStack creates a vertical stack of fixed children.
func VStack(children ...runtime.Widget) *runtime.Flex {
	flexChildren := make([]runtime.FlexChild, 0, len(children))
	for _, child := range children {
		if child == nil {
			continue
		}
		flexChildren = append(flexChildren, runtime.Fixed(child))
	}
	return runtime.VBox(flexChildren...)
}

// HStack creates a horizontal stack of fixed children.
func HStack(children ...runtime.Widget) *runtime.Flex {
	flexChildren := make([]runtime.FlexChild, 0, len(children))
	for _, child := range children {
		if child == nil {
			continue
		}
		flexChildren = append(flexChildren, runtime.Fixed(child))
	}
	return runtime.HBox(flexChildren...)
}

// VFlex creates a vertical flex container with explicit flex children.
func VFlex(children ...runtime.FlexChild) *runtime.Flex {
	return runtime.VBox(children...)
}

// HFlex creates a horizontal flex container with explicit flex children.
func HFlex(children ...runtime.FlexChild) *runtime.Flex {
	return runtime.HBox(children...)
}

// Fixed wraps a widget as a flex child that uses its natural measured size.
func Fixed(w runtime.Widget) runtime.FlexChild { return runtime.Fixed(w) }

// Flexible wraps a widget as a flex child that grows by the given factor.
// A grow of 1 shares remaining space equally with other flexible children.
func Flexible(w runtime.Widget, grow float64) runtime.FlexChild { return runtime.Flexible(w, grow) }

// Expanded wraps a widget as a flex child that fills all remaining space (grow=1).
func Expanded(w runtime.Widget) runtime.FlexChild { return runtime.Expanded(w) }

// Sized wraps a widget as a flex child with a fixed basis size in the main axis.
func Sized(w runtime.Widget, basis int) runtime.FlexChild { return runtime.Sized(w, basis) }

// Space creates an expanding empty spacer for pushing siblings apart.
func Space() runtime.FlexChild { return runtime.Space() }

// FixedSpace creates an empty spacer with a fixed size in the main axis.
func FixedSpace(size int) runtime.FlexChild { return runtime.FixedSpace(size) }
