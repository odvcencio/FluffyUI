package widgets

import (
	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/runtime"
)

// TooltipTrigger describes how a tooltip is activated.
type TooltipTrigger int

const (
	TooltipHover TooltipTrigger = iota
	TooltipFocus
	TooltipClick
)

// Tooltip displays content anchored to a target widget.
type Tooltip struct {
	Base

	target    runtime.Widget
	content   runtime.Widget
	trigger   TooltipTrigger
	placement PopoverPlacement
	gap       int
	open      bool
}

// TooltipOption configures a tooltip.
type TooltipOption = Option[Tooltip]

// NewTooltip creates a tooltip wrapper.
func NewTooltip(target runtime.Widget, content runtime.Widget, opts ...TooltipOption) *Tooltip {
	t := &Tooltip{
		target:    target,
		content:   content,
		trigger:   TooltipHover,
		placement: PopoverAuto,
	}
	t.Base.Role = accessibility.RoleGroup
	for _, opt := range opts {
		if opt != nil {
			opt(t)
		}
	}
	return t
}

// WithTooltipTrigger sets the activation trigger.
func WithTooltipTrigger(trigger TooltipTrigger) TooltipOption {
	return func(t *Tooltip) {
		if t == nil {
			return
		}
		t.trigger = trigger
	}
}

// WithTooltipPlacement sets the popover placement.
func WithTooltipPlacement(placement PopoverPlacement) TooltipOption {
	return func(t *Tooltip) {
		if t == nil {
			return
		}
		t.placement = placement
	}
}

// WithTooltipGap sets the gap between target and tooltip.
func WithTooltipGap(gap int) TooltipOption {
	return func(t *Tooltip) {
		if t == nil {
			return
		}
		t.gap = gap
	}
}

// Measure returns the target size within constraints.
func (t *Tooltip) Measure(constraints runtime.Constraints) runtime.Size {
	return t.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		if t != nil && t.target != nil {
			return t.target.Measure(contentConstraints)
		}
		return contentConstraints.MinSize()
	})
}

// Layout assigns bounds to the target.
func (t *Tooltip) Layout(bounds runtime.Rect) {
	if t == nil {
		return
	}
	t.Base.Layout(bounds)
	if t.target == nil {
		return
	}
	content := t.ContentBounds()
	t.target.Layout(content)
}

// Render draws the target.
func (t *Tooltip) Render(ctx runtime.RenderContext) {
	if t == nil {
		return
	}
	runtime.RenderChild(ctx, t.target)
}

// HandleMessage forwards messages to the target and manages tooltip state.
func (t *Tooltip) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t == nil {
		return runtime.Unhandled()
	}

	if focusMsg, ok := msg.(runtime.FocusChangedMsg); ok {
		return t.handleFocusChanged(focusMsg)
	}

	result := runtime.Unhandled()
	if t.target != nil {
		result = t.target.HandleMessage(msg)
	}

	if mouseMsg, ok := msg.(runtime.MouseMsg); ok {
		return t.handleMouse(mouseMsg, result)
	}

	return result
}

// ChildWidgets returns the tooltip target.
func (t *Tooltip) ChildWidgets() []runtime.Widget {
	if t == nil || t.target == nil {
		return nil
	}
	return []runtime.Widget{t.target}
}

// HitSelf ensures the tooltip receives mouse events for its bounds.
func (t *Tooltip) HitSelf() bool {
	return true
}

func (t *Tooltip) handleMouse(msg runtime.MouseMsg, result runtime.HandleResult) runtime.HandleResult {
	if t == nil {
		return result
	}
	inside := t.anchorRect().Contains(msg.X, msg.Y)
	if t.trigger == TooltipHover {
		if (msg.Action == runtime.MouseMove || msg.Action == runtime.MousePress) && inside {
			return t.mergeResult(result, t.openTooltip(true))
		}
		return result
	}
	if t.trigger == TooltipClick && msg.Action == runtime.MousePress && inside {
		if t.open {
			return t.mergeResult(result, commandResult(runtime.PopOverlay{}))
		}
		return t.mergeResult(result, t.openTooltip(false))
	}
	return result
}

func (t *Tooltip) handleFocusChanged(msg runtime.FocusChangedMsg) runtime.HandleResult {
	if t == nil || t.trigger != TooltipFocus {
		return runtime.Unhandled()
	}
	wasFocused := t.containsFocusable(msg.Prev)
	isFocused := t.containsFocusable(msg.Next)
	if isFocused && !t.open {
		return t.openTooltip(false)
	}
	if wasFocused && !isFocused && t.open {
		return commandResult(runtime.PopOverlay{})
	}
	return runtime.Unhandled()
}

func (t *Tooltip) containsFocusable(target runtime.Focusable) bool {
	if t == nil || target == nil {
		return false
	}
	return widgetContains(t.target, target)
}

func widgetContains(root runtime.Widget, target runtime.Widget) bool {
	if root == nil || target == nil {
		return false
	}
	if root == target {
		return true
	}
	if container, ok := root.(runtime.ChildProvider); ok {
		for _, child := range container.ChildWidgets() {
			if widgetContains(child, target) {
				return true
			}
		}
	}
	return false
}

func (t *Tooltip) anchorRect() runtime.Rect {
	if t != nil && t.target != nil {
		if bounds, ok := t.target.(runtime.BoundsProvider); ok {
			return bounds.Bounds()
		}
	}
	if t == nil {
		return runtime.Rect{}
	}
	return t.Bounds()
}

func (t *Tooltip) openTooltip(dismissOnMoveOutside bool) runtime.HandleResult {
	if t == nil || t.open || t.content == nil {
		return runtime.Unhandled()
	}
	anchor := t.anchorRect()
	options := []PopoverOption{
		WithPopoverPlacement(t.placement),
		WithPopoverGap(t.gap),
		WithPopoverOnClose(func() { t.open = false }),
		WithPopoverDismissOnEscape(true),
	}
	if dismissOnMoveOutside {
		options = append(options, WithPopoverDismissOnMoveOutside(true))
	}
	if t.trigger == TooltipClick {
		options = append(options, WithPopoverDismissOnOutside(true))
	}
	popover := NewPopover(anchor, t.content, options...)
	t.open = true
	return commandResult(runtime.PushOverlay{Widget: popover, Modal: false})
}

func (t *Tooltip) mergeResult(base runtime.HandleResult, extra runtime.HandleResult) runtime.HandleResult {
	if len(extra.Commands) == 0 {
		return base
	}
	base.Commands = append(base.Commands, extra.Commands...)
	if extra.Handled {
		base.Handled = true
	}
	return base
}

func commandResult(cmd runtime.Command) runtime.HandleResult {
	if cmd == nil {
		return runtime.Unhandled()
	}
	return runtime.HandleResult{Handled: false, Commands: []runtime.Command{cmd}}
}

var _ runtime.Widget = (*Tooltip)(nil)
var _ runtime.ChildProvider = (*Tooltip)(nil)
var _ runtime.HitSelfProvider = (*Tooltip)(nil)
