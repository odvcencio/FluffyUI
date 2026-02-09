package widgets

import (
	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
)

// FixedSpacer is an empty widget with a fixed size, useful for adding
// precise spacing between widgets in a layout. Unlike runtime.FixedSpace
// which is a flex child helper, FixedSpacer is a full widget that can
// participate in the accessibility tree and stylesheet system.
type FixedSpacer struct {
	Base
	width  int
	height int
}

// NewFixedSpacer creates a fixed-size empty widget.
// The spacer occupies exactly the given width and height and renders nothing.
func NewFixedSpacer(width, height int) *FixedSpacer {
	s := &FixedSpacer{width: width, height: height}
	s.Base.Role = accessibility.RolePresentation
	return s
}

// Measure returns the fixed size, clamped to the given constraints.
func (s *FixedSpacer) Measure(constraints runtime.Constraints) runtime.Size {
	if s == nil {
		return constraints.MinSize()
	}
	return constraints.Constrain(runtime.Size{Width: s.width, Height: s.height})
}

// Render is a no-op; the spacer is invisible.
func (s *FixedSpacer) Render(ctx runtime.RenderContext) {
	// Spacer renders nothing.
}

// HandleMessage returns unhandled; spacers do not process input.
func (s *FixedSpacer) HandleMessage(msg runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}

// StyleType returns the selector type name for stylesheets.
func (s *FixedSpacer) StyleType() string {
	return "FixedSpacer"
}

var _ runtime.Widget = (*FixedSpacer)(nil)
