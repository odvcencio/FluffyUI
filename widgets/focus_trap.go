package widgets

import (
	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

// FocusTrap wraps a child widget and constrains Tab/Shift+Tab navigation
// within the child subtree. This is critical for modal dialogs, drawers,
// and overlays that must not let keyboard focus escape.
//
// When active, Tab and Shift+Tab cycle only through focusable descendants.
// When inactive, focus navigation passes through normally.
type FocusTrap struct {
	Base
	child      runtime.Widget
	active     bool
	services   runtime.Services
	scope      *runtime.FocusScope // internal scope for trapped focus
	announced  bool
	registered bool // whether focusables have been collected
}

var (
	_ runtime.Widget       = (*FocusTrap)(nil)
	_ runtime.ChildProvider = (*FocusTrap)(nil)
	_ runtime.Bindable     = (*FocusTrap)(nil)
	_ runtime.Unbindable   = (*FocusTrap)(nil)
)

// NewFocusTrap creates a FocusTrap wrapping the given child widget.
// The trap starts active by default.
func NewFocusTrap(child runtime.Widget) *FocusTrap {
	ft := &FocusTrap{
		child:  child,
		active: true,
		scope:  runtime.NewFocusScopeWithPolicy(runtime.AutoFocusFirst),
	}
	ft.Base.Role = accessibility.RoleDialog
	ft.Base.Label = "Focus trap"
	return ft
}

// Bind attaches app services and announces the focus trap.
func (f *FocusTrap) Bind(services runtime.Services) {
	if f == nil {
		return
	}
	f.services = services
}

// Unbind releases app services.
func (f *FocusTrap) Unbind() {
	if f == nil {
		return
	}
	f.services = runtime.Services{}
}

// SetActive enables or disables the focus trap.
// When activated, focus is announced and trapped within children.
// When deactivated, Tab/Shift+Tab pass through to the parent scope.
func (f *FocusTrap) SetActive(active bool) {
	if f == nil {
		return
	}
	prev := f.active
	f.active = active
	if active && !prev {
		f.announced = false
		f.registered = false
		f.refreshScope()
		f.announce()
	}
}

// Active reports whether the focus trap is currently active.
func (f *FocusTrap) Active() bool {
	if f == nil {
		return false
	}
	return f.active
}

// SetLabel sets the accessible label for announcements.
func (f *FocusTrap) SetLabel(label string) {
	if f == nil {
		return
	}
	f.Base.Label = label
}

// SetChild replaces the child widget.
func (f *FocusTrap) SetChild(child runtime.Widget) {
	if f == nil {
		return
	}
	f.child = child
	f.registered = false
}

// Child returns the wrapped child widget.
func (f *FocusTrap) Child() runtime.Widget {
	if f == nil {
		return nil
	}
	return f.child
}

// StyleType returns the selector type name.
func (f *FocusTrap) StyleType() string {
	return "FocusTrap"
}

// ChildWidgets returns the child widget for tree traversal.
func (f *FocusTrap) ChildWidgets() []runtime.Widget {
	if f == nil || f.child == nil {
		return nil
	}
	return []runtime.Widget{f.child}
}

// PathSegment returns a debug path segment.
func (f *FocusTrap) PathSegment(child runtime.Widget) string {
	return "FocusTrap"
}

// Measure returns desired size, delegating to the child.
func (f *FocusTrap) Measure(constraints runtime.Constraints) runtime.Size {
	if f == nil || f.child == nil {
		return constraints.Constrain(runtime.Size{})
	}
	return f.child.Measure(constraints)
}

// Layout positions the child within the given bounds.
func (f *FocusTrap) Layout(bounds runtime.Rect) {
	if f == nil {
		return
	}
	f.Base.Layout(bounds)
	if f.child != nil {
		f.child.Layout(f.ContentBounds())
	}
}

// Render draws the child widget.
func (f *FocusTrap) Render(ctx runtime.RenderContext) {
	if f == nil || f.child == nil {
		return
	}
	if f.active && !f.registered {
		f.refreshScope()
	}
	if f.active && !f.announced {
		f.announce()
	}
	runtime.RenderChild(ctx, f.child)
}

// HandleMessage intercepts Tab/Shift+Tab when the trap is active
// to cycle focus within the child subtree.
func (f *FocusTrap) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if f == nil || f.child == nil {
		return runtime.Unhandled()
	}

	// When active, intercept Tab/Shift+Tab for trapped focus cycling
	if f.active {
		if key, ok := msg.(runtime.KeyMsg); ok && key.Key == terminal.KeyTab {
			if !f.registered {
				f.refreshScope()
			}
			if key.Shift {
				f.scope.FocusPrev()
			} else {
				f.scope.FocusNext()
			}
			return runtime.Handled()
		}
	}

	// Pass all other messages to the child
	return f.child.HandleMessage(msg)
}

// refreshScope rebuilds the internal focus scope from the child tree.
func (f *FocusTrap) refreshScope() {
	if f == nil {
		return
	}
	f.scope.Reset()
	if f.child != nil {
		runtime.RegisterFocusables(f.scope, f.child)
	}
	f.registered = true
}

// announce emits an accessibility announcement when the trap activates.
func (f *FocusTrap) announce() {
	if f == nil || f.announced {
		return
	}
	announcer := f.services.Announcer()
	if announcer == nil {
		return
	}
	label := f.Base.Label
	if label == "" {
		label = "Focus trap"
	}
	announcer.Announce("Focus trapped in "+label, accessibility.PriorityAssertive)
	f.announced = true
}

// Scope returns the internal focus scope for testing and inspection.
func (f *FocusTrap) Scope() *runtime.FocusScope {
	if f == nil {
		return nil
	}
	return f.scope
}
