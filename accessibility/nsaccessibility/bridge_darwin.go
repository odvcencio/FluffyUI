//go:build darwin

// Package nsaccessibility provides macOS accessibility support via NSAccessibility.
package nsaccessibility

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework AppKit

#include <stdlib.h>

// Forward declarations for Objective-C functions
void nsaccessibility_init(void);
void nsaccessibility_cleanup(void);
void nsaccessibility_announce(const char* message, int priority);
void nsaccessibility_notify_focus_change(const char* label, const char* role);
void nsaccessibility_notify_value_change(const char* label, const char* old_value, const char* new_value);
void nsaccessibility_notify_state_change(const char* label, const char* state, int value);
int nsaccessibility_is_voiceover_running(void);
*/
import "C"

import (
	"sync"
	"unsafe"
)

// Priority levels for announcements.
type Priority int

const (
	PriorityLow Priority = iota
	PriorityMedium
	PriorityHigh
	PriorityUrgent
)

// Accessible provides accessibility metadata.
type Accessible interface {
	AccessibleRole() string
	AccessibleLabel() string
	AccessibleDescription() string
	AccessibleState() StateSet
	AccessibleValue() *ValueInfo
}

// StateSet describes the state of a widget.
type StateSet struct {
	Checked  *bool
	Expanded *bool
	Selected bool
	Disabled bool
	ReadOnly bool
	Required bool
	Invalid  bool
}

// ValueInfo describes a widget's value.
type ValueInfo struct {
	Min     float64
	Max     float64
	Current float64
	Text    string
}

// BridgeApp provides access to the application state.
type BridgeApp interface {
	Name() string
	RootAccessible() Accessible
	FocusedAccessible() Accessible
	AccessibleAt(path string) Accessible
	ChildAccessibles(parent Accessible) []Accessible
}

// Bridge implements accessibility via NSAccessibility for macOS.
type Bridge struct {
	mu     sync.RWMutex
	app    BridgeApp
	tree   *AccessibleTree
	closed bool
}

// NewBridge creates a new NSAccessibility bridge.
func NewBridge() *Bridge {
	return &Bridge{}
}

// Register connects to the NSAccessibility system.
func (b *Bridge) Register(app BridgeApp) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.app = app

	// Initialize Cocoa accessibility
	C.nsaccessibility_init()

	// Build the accessible tree
	b.tree = NewAccessibleTree(app)

	return nil
}

// Announce sends text to VoiceOver.
func (b *Bridge) Announce(text string, priority Priority) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed || text == "" {
		return nil
	}

	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	// Map priority to NSAccessibility priority
	cPriority := C.int(0) // Normal priority
	if priority >= PriorityHigh {
		cPriority = C.int(1) // High priority
	}

	C.nsaccessibility_announce(cText, cPriority)
	return nil
}

// UpdateTree notifies VoiceOver that the widget structure changed.
func (b *Bridge) UpdateTree() error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return nil
	}

	// Rebuild the accessible tree
	if b.tree != nil && b.app != nil {
		b.tree.Rebuild(b.app)
	}

	// macOS handles tree updates automatically through accessibility elements
	return nil
}

// NotifyFocusChange informs VoiceOver that focus moved.
func (b *Bridge) NotifyFocusChange(widget Accessible) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed || widget == nil {
		return nil
	}

	label := widget.AccessibleLabel()
	role := widget.AccessibleRole()

	cLabel := C.CString(label)
	cRole := C.CString(role)
	defer C.free(unsafe.Pointer(cLabel))
	defer C.free(unsafe.Pointer(cRole))

	C.nsaccessibility_notify_focus_change(cLabel, cRole)
	return nil
}

// NotifyValueChange informs VoiceOver that a widget's value changed.
func (b *Bridge) NotifyValueChange(widget Accessible, oldValue, newValue string) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed || widget == nil {
		return nil
	}

	label := widget.AccessibleLabel()

	cLabel := C.CString(label)
	cOldValue := C.CString(oldValue)
	cNewValue := C.CString(newValue)
	defer C.free(unsafe.Pointer(cLabel))
	defer C.free(unsafe.Pointer(cOldValue))
	defer C.free(unsafe.Pointer(cNewValue))

	C.nsaccessibility_notify_value_change(cLabel, cOldValue, cNewValue)
	return nil
}

// NotifyStateChange informs VoiceOver that a widget's state changed.
func (b *Bridge) NotifyStateChange(widget Accessible, state string, value bool) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed || widget == nil {
		return nil
	}

	label := widget.AccessibleLabel()

	cLabel := C.CString(label)
	cState := C.CString(state)
	defer C.free(unsafe.Pointer(cLabel))
	defer C.free(unsafe.Pointer(cState))

	cValue := C.int(0)
	if value {
		cValue = C.int(1)
	}

	C.nsaccessibility_notify_state_change(cLabel, cState, cValue)
	return nil
}

// Close disconnects from the NSAccessibility system.
func (b *Bridge) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}
	b.closed = true

	C.nsaccessibility_cleanup()
	return nil
}

// IsVoiceOverRunning returns true if VoiceOver is currently running.
func (b *Bridge) IsVoiceOverRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return false
	}

	return C.nsaccessibility_is_voiceover_running() != 0
}
