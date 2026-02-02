//go:build linux && cgo

package accessibility

import (
	"github.com/odvcencio/fluffyui/accessibility/atspi"
)

// NewBridge returns an AT-SPI bridge for Linux.
func NewBridge() Bridge {
	return &atspiBridgeWrapper{bridge: atspi.NewBridge()}
}

// atspiBridgeWrapper wraps the atspi.Bridge to implement accessibility.Bridge.
type atspiBridgeWrapper struct {
	bridge *atspi.Bridge
}

func (w *atspiBridgeWrapper) Register(app BridgeApp) error {
	if app == nil {
		return w.bridge.Register(nil)
	}
	return w.bridge.Register(&atspiAppAdapter{app: app})
}

func (w *atspiBridgeWrapper) Announce(text string, priority Priority) error {
	return w.bridge.Announce(text, atspi.Priority(priority))
}

func (w *atspiBridgeWrapper) UpdateTree() error {
	return w.bridge.UpdateTree()
}

func (w *atspiBridgeWrapper) NotifyFocusChange(widget Accessible) error {
	if widget == nil {
		return w.bridge.NotifyFocusChange(nil)
	}
	return w.bridge.NotifyFocusChange(&atspiAccessibleAdapter{accessible: widget})
}

func (w *atspiBridgeWrapper) NotifyValueChange(widget Accessible, oldValue, newValue string) error {
	if widget == nil {
		return w.bridge.NotifyValueChange(nil, oldValue, newValue)
	}
	return w.bridge.NotifyValueChange(&atspiAccessibleAdapter{accessible: widget}, oldValue, newValue)
}

func (w *atspiBridgeWrapper) NotifyStateChange(widget Accessible, state string, value bool) error {
	if widget == nil {
		return w.bridge.NotifyStateChange(nil, state, value)
	}
	return w.bridge.NotifyStateChange(&atspiAccessibleAdapter{accessible: widget}, state, value)
}

func (w *atspiBridgeWrapper) Close() error {
	return w.bridge.Close()
}

// atspiAppAdapter adapts accessibility.BridgeApp to atspi.BridgeApp.
type atspiAppAdapter struct {
	app BridgeApp
}

func (a *atspiAppAdapter) Name() string {
	return a.app.Name()
}

func (a *atspiAppAdapter) RootAccessible() atspi.Accessible {
	root := a.app.RootAccessible()
	if root == nil {
		return nil
	}
	return &atspiAccessibleAdapter{accessible: root}
}

func (a *atspiAppAdapter) FocusedAccessible() atspi.Accessible {
	focused := a.app.FocusedAccessible()
	if focused == nil {
		return nil
	}
	return &atspiAccessibleAdapter{accessible: focused}
}

func (a *atspiAppAdapter) AccessibleAt(path string) atspi.Accessible {
	acc := a.app.AccessibleAt(path)
	if acc == nil {
		return nil
	}
	return &atspiAccessibleAdapter{accessible: acc}
}

func (a *atspiAppAdapter) ChildAccessibles(parent atspi.Accessible) []atspi.Accessible {
	// Get the underlying accessibility.Accessible from the adapter
	var parentAcc Accessible
	if adapter, ok := parent.(*atspiAccessibleAdapter); ok {
		parentAcc = adapter.accessible
	}
	if parentAcc == nil {
		return nil
	}

	children := a.app.ChildAccessibles(parentAcc)
	result := make([]atspi.Accessible, len(children))
	for i, child := range children {
		result[i] = &atspiAccessibleAdapter{accessible: child}
	}
	return result
}

// atspiAccessibleAdapter adapts accessibility.Accessible to atspi.Accessible.
type atspiAccessibleAdapter struct {
	accessible Accessible
}

func (a *atspiAccessibleAdapter) AccessibleRole() string {
	return string(a.accessible.AccessibleRole())
}

func (a *atspiAccessibleAdapter) AccessibleLabel() string {
	return a.accessible.AccessibleLabel()
}

func (a *atspiAccessibleAdapter) AccessibleDescription() string {
	return a.accessible.AccessibleDescription()
}

func (a *atspiAccessibleAdapter) AccessibleState() atspi.StateSet {
	s := a.accessible.AccessibleState()
	return atspi.StateSet{
		Checked:  s.Checked,
		Expanded: s.Expanded,
		Selected: s.Selected,
		Disabled: s.Disabled,
		ReadOnly: s.ReadOnly,
		Required: s.Required,
		Invalid:  s.Invalid,
	}
}

func (a *atspiAccessibleAdapter) AccessibleValue() *atspi.ValueInfo {
	v := a.accessible.AccessibleValue()
	if v == nil {
		return nil
	}
	return &atspi.ValueInfo{
		Min:     v.Min,
		Max:     v.Max,
		Current: v.Current,
		Text:    v.Text,
	}
}
