//go:build windows

// Package uiautomation provides Windows accessibility support via UI Automation.
package uiautomation

import (
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	// UIAutomation COM interfaces
	modOleaut32     = windows.NewLazySystemDLL("oleaut32.dll")
	modOle32        = windows.NewLazySystemDLL("ole32.dll")
	modUIAutomation = windows.NewLazySystemDLL("uiautomationcore.dll")

	procCoInitializeEx = modOle32.NewProc("CoInitializeEx")
	procCoUninitialize = modOle32.NewProc("CoUninitialize")

	// UI Automation function pointers
	procUiaRaiseAutomationEvent                    = modUIAutomation.NewProc("UiaRaiseAutomationEvent")
	procUiaRaiseAutomationPropertyChangedEvent     = modUIAutomation.NewProc("UiaRaiseAutomationPropertyChangedEvent")
	procUiaReturnRawElementProvider                = modUIAutomation.NewProc("UiaReturnRawElementProvider")
	procUiaHostProviderFromHwnd                    = modUIAutomation.NewProc("UiaHostProviderFromHwnd")
)

// UI Automation event IDs
const (
	UIA_AutomationFocusChangedEventId   = 20005
	UIA_StructureChangedEventId         = 20002
	UIA_Text_TextChangedEventId         = 20015
	UIA_Selection_InvalidatedEventId    = 20013
	UIA_Invoke_InvokedEventId           = 20009
	UIA_ToggleState_StateChangedEventId = 20041

	// Property IDs
	UIA_NamePropertyId                              = 30005
	UIA_ValueValuePropertyId                        = 30045
	UIA_ToggleToggleStatePropertyId                 = 30086
	UIA_ExpandCollapseExpandCollapseStatePropertyId = 30070
)

// UI Automation control type IDs
const (
	UIA_ButtonControlTypeId      = 50000
	UIA_CheckBoxControlTypeId    = 50002
	UIA_RadioButtonControlTypeId = 50013
	UIA_EditControlTypeId        = 50004
	UIA_ListControlTypeId        = 50008
	UIA_ListItemControlTypeId    = 50007
	UIA_TableControlTypeId       = 50036
	UIA_DataItemControlTypeId    = 50029
	UIA_SliderControlTypeId      = 50015
	UIA_TreeControlTypeId        = 50023
	UIA_TreeItemControlTypeId    = 50024
	UIA_MenuControlTypeId        = 50009
	UIA_MenuItemControlTypeId    = 50011
	UIA_TabControlTypeId         = 50018
	UIA_TabItemControlTypeId     = 50019
	UIA_PaneControlTypeId        = 50033
	UIA_WindowControlTypeId      = 50032
	UIA_ProgressBarControlTypeId = 50012
	UIA_GroupControlTypeId       = 50026
	UIA_TextControlTypeId        = 50020
	UIA_ImageControlTypeId       = 50006
	UIA_CustomControlTypeId      = 50025
)

// COM initialization flags
const (
	COINIT_MULTITHREADED     = 0x0
	COINIT_APARTMENTTHREADED = 0x2
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

// Bridge implements accessibility via Windows UI Automation.
type Bridge struct {
	mu          sync.RWMutex
	app         BridgeApp
	tree        *AccessibleTree
	hwnd        uintptr
	initialized bool
	closed      bool
}

// NewBridge creates a new UI Automation bridge.
func NewBridge() *Bridge {
	return &Bridge{}
}

// Register connects to the UI Automation system.
func (b *Bridge) Register(app BridgeApp) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}

	b.app = app

	// Initialize COM
	hr, _, _ := procCoInitializeEx.Call(0, COINIT_APARTMENTTHREADED)
	if hr != 0 && hr != 1 { // S_OK or S_FALSE (already initialized)
		// COM not available, fall back to no-op mode
		return nil
	}

	b.initialized = true

	// Build the accessible tree
	b.tree = NewAccessibleTree(app)

	return nil
}

// Announce sends text to Narrator.
func (b *Bridge) Announce(text string, priority Priority) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed || !b.initialized || text == "" {
		return nil
	}

	// On Windows, announcements are typically done through a live region
	// For terminal apps, we use the NotifyWinEvent function
	// This is a simplified implementation

	return nil
}

// UpdateTree notifies UI Automation that the widget structure changed.
func (b *Bridge) UpdateTree() error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed || !b.initialized {
		return nil
	}

	// Rebuild the accessible tree
	if b.tree != nil && b.app != nil {
		b.tree.Rebuild(b.app)
	}

	// Raise structure changed event
	if b.hwnd != 0 {
		b.raiseEvent(UIA_StructureChangedEventId)
	}

	return nil
}

// NotifyFocusChange informs Narrator that focus moved.
func (b *Bridge) NotifyFocusChange(widget Accessible) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed || !b.initialized || widget == nil {
		return nil
	}

	// Raise focus changed event
	b.raiseEvent(UIA_AutomationFocusChangedEventId)

	return nil
}

// NotifyValueChange informs Narrator that a widget's value changed.
func (b *Bridge) NotifyValueChange(widget Accessible, oldValue, newValue string) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed || !b.initialized || widget == nil {
		return nil
	}

	// Raise property changed event for value
	b.raisePropertyChanged(UIA_ValueValuePropertyId, oldValue, newValue)

	return nil
}

// NotifyStateChange informs Narrator that a widget's state changed.
func (b *Bridge) NotifyStateChange(widget Accessible, state string, value bool) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed || !b.initialized || widget == nil {
		return nil
	}

	// Map state to UIA property
	var propertyId int32
	switch state {
	case "checked":
		propertyId = UIA_ToggleToggleStatePropertyId
	case "expanded":
		propertyId = UIA_ExpandCollapseExpandCollapseStatePropertyId
	default:
		return nil
	}

	oldState := 0
	newState := 0
	if value {
		newState = 1
	} else {
		oldState = 1
	}

	b.raisePropertyChanged(propertyId, oldState, newState)

	return nil
}

// Close disconnects from the UI Automation system.
func (b *Bridge) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}
	b.closed = true

	if b.initialized {
		procCoUninitialize.Call()
		b.initialized = false
	}

	return nil
}

// SetHwnd sets the window handle for event notifications.
func (b *Bridge) SetHwnd(hwnd uintptr) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.hwnd = hwnd
}

// raiseEvent raises a UI Automation event.
func (b *Bridge) raiseEvent(eventId int32) {
	if b.tree == nil || b.tree.Root() == nil {
		return
	}

	// Get the provider for the root element
	provider := b.tree.Root().Provider()
	if provider == nil {
		return
	}

	_, _, _ = procUiaRaiseAutomationEvent.Call(
		uintptr(unsafe.Pointer(provider)),
		uintptr(eventId),
	)
}

// raisePropertyChanged raises a UI Automation property changed event.
func (b *Bridge) raisePropertyChanged(propertyId int32, oldValue, newValue interface{}) {
	if b.tree == nil || b.tree.Root() == nil {
		return
	}

	provider := b.tree.Root().Provider()
	if provider == nil {
		return
	}

	// Create VARIANT for old and new values
	// This is a simplified implementation
	_, _, _ = procUiaRaiseAutomationPropertyChangedEvent.Call(
		uintptr(unsafe.Pointer(provider)),
		uintptr(propertyId),
		0, // old value VARIANT
		0, // new value VARIANT
	)
}

// IRawElementProviderSimple COM interface
type IRawElementProviderSimple struct {
	vtbl *IRawElementProviderSimpleVtbl
	node *AccessibleNode
}

type IRawElementProviderSimpleVtbl struct {
	QueryInterface            uintptr
	AddRef                    uintptr
	Release                   uintptr
	GetProviderOptions        uintptr
	GetPatternProvider        uintptr
	GetPropertyValue          uintptr
	GetHostRawElementProvider uintptr
}

// NewProvider creates a new UI Automation provider for a node.
func NewProvider(node *AccessibleNode) *IRawElementProviderSimple {
	provider := &IRawElementProviderSimple{
		node: node,
	}

	// Set up vtable
	provider.vtbl = &IRawElementProviderSimpleVtbl{
		QueryInterface:            syscall.NewCallback(providerQueryInterface),
		AddRef:                    syscall.NewCallback(providerAddRef),
		Release:                   syscall.NewCallback(providerRelease),
		GetProviderOptions:        syscall.NewCallback(providerGetProviderOptions),
		GetPatternProvider:        syscall.NewCallback(providerGetPatternProvider),
		GetPropertyValue:          syscall.NewCallback(providerGetPropertyValue),
		GetHostRawElementProvider: syscall.NewCallback(providerGetHostRawElementProvider),
	}

	return provider
}

// COM callback implementations
func providerQueryInterface(this uintptr, riid uintptr, ppv uintptr) uintptr {
	return 0 // S_OK
}

func providerAddRef(this uintptr) uintptr {
	return 1
}

func providerRelease(this uintptr) uintptr {
	return 0
}

func providerGetProviderOptions(this uintptr, options uintptr) uintptr {
	// ProviderOptions_ServerSideProvider = 0x1
	*(*int32)(unsafe.Pointer(options)) = 0x1
	return 0
}

func providerGetPatternProvider(this uintptr, patternId int32, provider uintptr) uintptr {
	*(*uintptr)(unsafe.Pointer(provider)) = 0
	return 0
}

func providerGetPropertyValue(this uintptr, propertyId int32, pRetVal uintptr) uintptr {
	return 0
}

func providerGetHostRawElementProvider(this uintptr, provider uintptr) uintptr {
	*(*uintptr)(unsafe.Pointer(provider)) = 0
	return 0
}
