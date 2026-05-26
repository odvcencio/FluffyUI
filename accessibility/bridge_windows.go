//go:build windows && cgo

package accessibility

import (
	"m31labs.dev/fluffyui/accessibility/uiautomation"
)

// NewBridge returns a UI Automation bridge for Windows.
func NewBridge() Bridge {
	return &uiaBridgeWrapper{bridge: uiautomation.NewBridge()}
}

// uiaBridgeWrapper wraps the uiautomation.Bridge to implement accessibility.Bridge.
type uiaBridgeWrapper struct {
	bridge *uiautomation.Bridge
}

func (w *uiaBridgeWrapper) Register(app BridgeApp) error {
	if app == nil {
		return w.bridge.Register(nil)
	}
	return w.bridge.Register(&uiaAppAdapter{app: app})
}

func (w *uiaBridgeWrapper) Announce(text string, priority Priority) error {
	return w.bridge.Announce(text, uiautomation.Priority(priority))
}

func (w *uiaBridgeWrapper) UpdateTree() error {
	return w.bridge.UpdateTree()
}

func (w *uiaBridgeWrapper) NotifyFocusChange(widget Accessible) error {
	if widget == nil {
		return w.bridge.NotifyFocusChange(nil)
	}
	return w.bridge.NotifyFocusChange(&uiaAccessibleAdapter{accessible: widget})
}

func (w *uiaBridgeWrapper) NotifyValueChange(widget Accessible, oldValue, newValue string) error {
	if widget == nil {
		return w.bridge.NotifyValueChange(nil, oldValue, newValue)
	}
	return w.bridge.NotifyValueChange(&uiaAccessibleAdapter{accessible: widget}, oldValue, newValue)
}

func (w *uiaBridgeWrapper) NotifyStateChange(widget Accessible, state string, value bool) error {
	if widget == nil {
		return w.bridge.NotifyStateChange(nil, state, value)
	}
	return w.bridge.NotifyStateChange(&uiaAccessibleAdapter{accessible: widget}, state, value)
}

func (w *uiaBridgeWrapper) Close() error {
	return w.bridge.Close()
}

// uiaAppAdapter adapts accessibility.BridgeApp to uiautomation.BridgeApp.
type uiaAppAdapter struct {
	app BridgeApp
}

func (a *uiaAppAdapter) Name() string {
	return a.app.Name()
}

func (a *uiaAppAdapter) RootAccessible() uiautomation.Accessible {
	root := a.app.RootAccessible()
	if root == nil {
		return nil
	}
	return &uiaAccessibleAdapter{accessible: root}
}

func (a *uiaAppAdapter) FocusedAccessible() uiautomation.Accessible {
	focused := a.app.FocusedAccessible()
	if focused == nil {
		return nil
	}
	return &uiaAccessibleAdapter{accessible: focused}
}

func (a *uiaAppAdapter) AccessibleAt(path string) uiautomation.Accessible {
	acc := a.app.AccessibleAt(path)
	if acc == nil {
		return nil
	}
	return &uiaAccessibleAdapter{accessible: acc}
}

func (a *uiaAppAdapter) ChildAccessibles(parent uiautomation.Accessible) []uiautomation.Accessible {
	var parentAcc Accessible
	if adapter, ok := parent.(*uiaAccessibleAdapter); ok {
		parentAcc = adapter.accessible
	}
	if parentAcc == nil {
		return nil
	}

	children := a.app.ChildAccessibles(parentAcc)
	result := make([]uiautomation.Accessible, len(children))
	for i, child := range children {
		result[i] = &uiaAccessibleAdapter{accessible: child}
	}
	return result
}

// uiaAccessibleAdapter adapts accessibility.Accessible to uiautomation.Accessible.
type uiaAccessibleAdapter struct {
	accessible Accessible
}

func (a *uiaAccessibleAdapter) AccessibleRole() string {
	return string(a.accessible.AccessibleRole())
}

func (a *uiaAccessibleAdapter) AccessibleLabel() string {
	return a.accessible.AccessibleLabel()
}

func (a *uiaAccessibleAdapter) AccessibleDescription() string {
	return a.accessible.AccessibleDescription()
}

func (a *uiaAccessibleAdapter) AccessibleState() uiautomation.StateSet {
	s := a.accessible.AccessibleState()
	return uiautomation.StateSet{
		Checked:  s.Checked,
		Expanded: s.Expanded,
		Selected: s.Selected,
		Disabled: s.Disabled,
		ReadOnly: s.ReadOnly,
		Required: s.Required,
		Invalid:  s.Invalid,
	}
}

func (a *uiaAccessibleAdapter) AccessibleValue() *uiautomation.ValueInfo {
	v := a.accessible.AccessibleValue()
	if v == nil {
		return nil
	}
	return &uiautomation.ValueInfo{
		Min:     v.Min,
		Max:     v.Max,
		Current: v.Current,
		Text:    v.Text,
	}
}
