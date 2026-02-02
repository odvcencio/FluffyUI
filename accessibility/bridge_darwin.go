//go:build darwin && cgo

package accessibility

import (
	"github.com/odvcencio/fluffyui/accessibility/nsaccessibility"
)

// NewBridge returns an NSAccessibility bridge for macOS.
func NewBridge() Bridge {
	return &nsaBridgeWrapper{bridge: nsaccessibility.NewBridge()}
}

// nsaBridgeWrapper wraps the nsaccessibility.Bridge to implement accessibility.Bridge.
type nsaBridgeWrapper struct {
	bridge *nsaccessibility.Bridge
}

func (w *nsaBridgeWrapper) Register(app BridgeApp) error {
	if app == nil {
		return w.bridge.Register(nil)
	}
	return w.bridge.Register(&nsaAppAdapter{app: app})
}

func (w *nsaBridgeWrapper) Announce(text string, priority Priority) error {
	return w.bridge.Announce(text, nsaccessibility.Priority(priority))
}

func (w *nsaBridgeWrapper) UpdateTree() error {
	return w.bridge.UpdateTree()
}

func (w *nsaBridgeWrapper) NotifyFocusChange(widget Accessible) error {
	if widget == nil {
		return w.bridge.NotifyFocusChange(nil)
	}
	return w.bridge.NotifyFocusChange(&nsaAccessibleAdapter{accessible: widget})
}

func (w *nsaBridgeWrapper) NotifyValueChange(widget Accessible, oldValue, newValue string) error {
	if widget == nil {
		return w.bridge.NotifyValueChange(nil, oldValue, newValue)
	}
	return w.bridge.NotifyValueChange(&nsaAccessibleAdapter{accessible: widget}, oldValue, newValue)
}

func (w *nsaBridgeWrapper) NotifyStateChange(widget Accessible, state string, value bool) error {
	if widget == nil {
		return w.bridge.NotifyStateChange(nil, state, value)
	}
	return w.bridge.NotifyStateChange(&nsaAccessibleAdapter{accessible: widget}, state, value)
}

func (w *nsaBridgeWrapper) Close() error {
	return w.bridge.Close()
}

// nsaAppAdapter adapts accessibility.BridgeApp to nsaccessibility.BridgeApp.
type nsaAppAdapter struct {
	app BridgeApp
}

func (a *nsaAppAdapter) Name() string {
	return a.app.Name()
}

func (a *nsaAppAdapter) RootAccessible() nsaccessibility.Accessible {
	root := a.app.RootAccessible()
	if root == nil {
		return nil
	}
	return &nsaAccessibleAdapter{accessible: root}
}

func (a *nsaAppAdapter) FocusedAccessible() nsaccessibility.Accessible {
	focused := a.app.FocusedAccessible()
	if focused == nil {
		return nil
	}
	return &nsaAccessibleAdapter{accessible: focused}
}

func (a *nsaAppAdapter) AccessibleAt(path string) nsaccessibility.Accessible {
	acc := a.app.AccessibleAt(path)
	if acc == nil {
		return nil
	}
	return &nsaAccessibleAdapter{accessible: acc}
}

func (a *nsaAppAdapter) ChildAccessibles(parent nsaccessibility.Accessible) []nsaccessibility.Accessible {
	var parentAcc Accessible
	if adapter, ok := parent.(*nsaAccessibleAdapter); ok {
		parentAcc = adapter.accessible
	}
	if parentAcc == nil {
		return nil
	}

	children := a.app.ChildAccessibles(parentAcc)
	result := make([]nsaccessibility.Accessible, len(children))
	for i, child := range children {
		result[i] = &nsaAccessibleAdapter{accessible: child}
	}
	return result
}

// nsaAccessibleAdapter adapts accessibility.Accessible to nsaccessibility.Accessible.
type nsaAccessibleAdapter struct {
	accessible Accessible
}

func (a *nsaAccessibleAdapter) AccessibleRole() string {
	return string(a.accessible.AccessibleRole())
}

func (a *nsaAccessibleAdapter) AccessibleLabel() string {
	return a.accessible.AccessibleLabel()
}

func (a *nsaAccessibleAdapter) AccessibleDescription() string {
	return a.accessible.AccessibleDescription()
}

func (a *nsaAccessibleAdapter) AccessibleState() nsaccessibility.StateSet {
	s := a.accessible.AccessibleState()
	return nsaccessibility.StateSet{
		Checked:  s.Checked,
		Expanded: s.Expanded,
		Selected: s.Selected,
		Disabled: s.Disabled,
		ReadOnly: s.ReadOnly,
		Required: s.Required,
		Invalid:  s.Invalid,
	}
}

func (a *nsaAccessibleAdapter) AccessibleValue() *nsaccessibility.ValueInfo {
	v := a.accessible.AccessibleValue()
	if v == nil {
		return nil
	}
	return &nsaccessibility.ValueInfo{
		Min:     v.Min,
		Max:     v.Max,
		Current: v.Current,
		Text:    v.Text,
	}
}
