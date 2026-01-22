package widgets

import (
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/state"
)

// SignalLabel is a tiny label bound to a signal.
// It demonstrates managing subscriptions in Mount/Unmount with a state.Scheduler.
type SignalLabel struct {
	Base
	source    state.Readable[string]
	scheduler state.Scheduler
	subs      state.Subscriptions
	text      string
	a11yLabel string
	style     backend.Style
	alignment Alignment
	mounted   bool
	styleSet  bool
}

// NewSignalLabel creates a new signal-backed label.
func NewSignalLabel(source state.Readable[string], scheduler state.Scheduler) *SignalLabel {
	label := &SignalLabel{
		source:    source,
		scheduler: scheduler,
		style:     backend.DefaultStyle(),
		alignment: AlignLeft,
	}
	label.subs.SetScheduler(scheduler)
	if source != nil {
		label.text = source.Get()
	}
	return label
}

// Text returns the current label text.
func (s *SignalLabel) Text() string {
	return s.text
}

// SetStyle sets the label style.
func (s *SignalLabel) SetStyle(style backend.Style) {
	s.style = style
	s.styleSet = true
}

// SetAlignment sets text alignment.
func (s *SignalLabel) SetAlignment(align Alignment) {
	s.alignment = align
}

// SetA11yLabel overrides the accessibility label without changing visible text.
func (s *SignalLabel) SetA11yLabel(label string) {
	s.a11yLabel = label
	s.syncA11y()
}

// StyleType returns the selector type name.
func (s *SignalLabel) StyleType() string {
	return "Label"
}

// Measure returns the size needed for the label.
func (s *SignalLabel) Measure(constraints runtime.Constraints) runtime.Size {
	return s.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		return contentConstraints.Constrain(runtime.Size{
			Width:  len(s.text),
			Height: 1,
		})
	})
}

// Render draws the label.
func (s *SignalLabel) Render(ctx runtime.RenderContext) {
	bounds := s.ContentBounds()
	if bounds.Width == 0 || bounds.Height == 0 {
		return
	}
	s.syncA11y()

	text := s.text
	if len(text) > bounds.Width {
		text = truncateString(text, bounds.Width)
	}

	x := bounds.X
	switch s.alignment {
	case AlignCenter:
		x = bounds.X + (bounds.Width-len(text))/2
	case AlignRight:
		x = bounds.X + bounds.Width - len(text)
	}

	baseStyle := resolveBaseStyle(ctx, s, s.style, s.styleSet)
	ctx.Buffer.SetString(x, bounds.Y, text, baseStyle)
}

// Mount subscribes to signal changes.
func (s *SignalLabel) Mount() {
	s.mounted = true
	s.subscribe()
}

// Unmount unsubscribes from signal changes.
func (s *SignalLabel) Unmount() {
	s.mounted = false
	s.subs.Clear()
}

func (s *SignalLabel) subscribe() {
	s.subs.Clear()
	if s.source == nil {
		s.text = ""
		s.syncA11y()
		return
	}
	s.text = s.source.Get()
	s.syncA11y()
	s.subs.Observe(s.source, s.onSignal)
}

func (s *SignalLabel) onSignal() {
	if !s.mounted || s.source == nil {
		return
	}
	s.text = s.source.Get()
	s.syncA11y()
}

func (s *SignalLabel) syncA11y() {
	if s == nil {
		return
	}
	if s.Base.Role == "" {
		s.Base.Role = accessibility.RoleText
	}
	override := strings.TrimSpace(s.a11yLabel)
	if override != "" {
		s.Base.Label = override
		value := strings.TrimSpace(s.text)
		if value != "" {
			s.Base.Value = &accessibility.ValueInfo{Text: value}
		} else {
			s.Base.Value = nil
		}
		return
	}
	label := strings.TrimSpace(s.text)
	if label == "" {
		label = "Label"
	}
	s.Base.Label = label
	s.Base.Value = nil
}
