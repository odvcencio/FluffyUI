package widgets

import (
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
)

const (
	skeletonCharLight = '\u2591' // ░
	skeletonCharMed   = '\u2592' // ▒
)

// Skeleton is an animated placeholder widget that indicates loading state.
// It renders alternating block characters that shift with each tick.
type Skeleton struct {
	Component

	width   int
	height  int
	animate bool
	frame   int

	style    backend.Style
	styleSet bool
}

// NewSkeleton creates a skeleton placeholder with the given dimensions.
func NewSkeleton(width, height int) *Skeleton {
	if width < 1 {
		width = 10
	}
	if height < 1 {
		height = 1
	}
	s := &Skeleton{
		width:   width,
		height:  height,
		animate: true,
		style:   backend.DefaultStyle(),
	}
	s.Base.Role = accessibility.RoleStatus
	s.Base.Live = accessibility.LivePolite
	s.Base.Label = "Loading"
	s.Base.State.Busy = true
	return s
}

// SetAnimate enables or disables animation.
func (s *Skeleton) SetAnimate(animate bool) {
	if s == nil {
		return
	}
	s.animate = animate
}

// IsAnimating returns whether animation is enabled.
func (s *Skeleton) IsAnimating() bool {
	if s == nil {
		return false
	}
	return s.animate
}

// Frame returns the current animation frame.
func (s *Skeleton) Frame() int {
	if s == nil {
		return 0
	}
	return s.frame
}

// SetStyle sets the widget style.
func (s *Skeleton) SetStyle(style backend.Style) {
	if s == nil {
		return
	}
	s.style = style
	s.styleSet = true
}

// StyleType returns the selector type name.
func (s *Skeleton) StyleType() string {
	return "Skeleton"
}

// Measure returns the desired size.
func (s *Skeleton) Measure(constraints runtime.Constraints) runtime.Size {
	return s.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		return contentConstraints.Constrain(runtime.Size{Width: s.width, Height: s.height})
	})
}

// Render draws the skeleton placeholder.
func (s *Skeleton) Render(ctx runtime.RenderContext) {
	if s == nil {
		return
	}
	s.syncA11y()
	bounds := s.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	style := resolveBaseStyle(ctx, s, s.style, s.styleSet)

	for row := 0; row < bounds.Height; row++ {
		var sb strings.Builder
		for col := 0; col < bounds.Width; col++ {
			if (col+row+s.frame)%2 == 0 {
				sb.WriteRune(skeletonCharLight)
			} else {
				sb.WriteRune(skeletonCharMed)
			}
		}
		ctx.Buffer.SetString(bounds.X, bounds.Y+row, sb.String(), style)
	}
}

// HandleMessage advances animation on ticks.
func (s *Skeleton) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if s == nil {
		return runtime.Unhandled()
	}
	if _, ok := msg.(runtime.TickMsg); ok {
		if s.animate {
			s.frame++
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (s *Skeleton) syncA11y() {
	if s == nil {
		return
	}
	if s.Base.Role == "" {
		s.Base.Role = accessibility.RoleStatus
	}
	if s.Base.Live == "" {
		s.Base.Live = accessibility.LivePolite
	}
	s.Base.State.Busy = true
	if s.Base.Label == "" {
		s.Base.Label = "Loading"
	}
}

var _ runtime.Widget = (*Skeleton)(nil)
var _ runtime.Bindable = (*Skeleton)(nil)
