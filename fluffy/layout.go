package fluffy

import "github.com/odvcencio/fluffy-ui/runtime"

// Flex re-exports.
type Flex = runtime.Flex
type FlexChild = runtime.FlexChild
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

// Flex child helpers.
func Fixed(w runtime.Widget) runtime.FlexChild { return runtime.Fixed(w) }
func Flexible(w runtime.Widget, grow float64) runtime.FlexChild { return runtime.Flexible(w, grow) }
func Expanded(w runtime.Widget) runtime.FlexChild { return runtime.Expanded(w) }
func Sized(w runtime.Widget, basis int) runtime.FlexChild { return runtime.Sized(w, basis) }
func Space() runtime.FlexChild { return runtime.Space() }
func FixedSpace(size int) runtime.FlexChild { return runtime.FixedSpace(size) }
