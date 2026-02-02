//go:build (linux || darwin || windows) && !cgo

package accessibility

// NewBridge returns a no-op bridge when CGo is not available.
// Platform-specific accessibility APIs require CGo for:
// - Linux: D-Bus communication for AT-SPI
// - macOS: Objective-C bindings for NSAccessibility
// - Windows: COM interfaces for UI Automation
func NewBridge() Bridge {
	return newStubBridge()
}

// newStubBridge creates a no-op bridge implementation.
func newStubBridge() Bridge {
	return &stubBridge{}
}

// stubBridge is a no-op implementation of Bridge.
type stubBridge struct{}

func (b *stubBridge) Register(app BridgeApp) error {
	return nil
}

func (b *stubBridge) Announce(text string, priority Priority) error {
	return nil
}

func (b *stubBridge) UpdateTree() error {
	return nil
}

func (b *stubBridge) NotifyFocusChange(widget Accessible) error {
	return nil
}

func (b *stubBridge) NotifyValueChange(widget Accessible, oldValue, newValue string) error {
	return nil
}

func (b *stubBridge) NotifyStateChange(widget Accessible, state string, value bool) error {
	return nil
}

func (b *stubBridge) Close() error {
	return nil
}
