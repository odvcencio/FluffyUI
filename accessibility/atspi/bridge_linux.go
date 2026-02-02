//go:build linux

// Package atspi provides Linux accessibility support via AT-SPI D-Bus.
package atspi

import (
	"fmt"
	"sync"
)

const (
	// D-Bus service and interface names
	atspiBus         = "org.a11y.Bus"
	atspiRegistry    = "org.a11y.atspi.Registry"
	atspiAccessible  = "org.a11y.atspi.Accessible"
	atspiComponent   = "org.a11y.atspi.Component"
	atspiText        = "org.a11y.atspi.Text"
	atspiValue       = "org.a11y.atspi.Value"
	atspiAction      = "org.a11y.atspi.Action"
	atspiApplication = "org.a11y.atspi.Application"
	atspiEvent       = "org.a11y.atspi.Event"

	// Object paths
	registryPath = "/org/a11y/atspi/registry"
	cachePath    = "/org/a11y/atspi/cache"
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

// Bridge implements accessibility via AT-SPI D-Bus.
type Bridge struct {
	mu       sync.RWMutex
	app      BridgeApp
	tree     *AccessibleTree
	conn     dbusConn
	busAddr  string
	appPath  string
	uniqueID int32
	closed   bool
}

// dbusConn abstracts the D-Bus connection for testing.
type dbusConn interface {
	Close() error
	Object(dest, path string) dbusObject
	Export(v interface{}, path, iface string) error
	RequestName(name string, flags uint) error
	Emit(path, iface, signal string, values ...interface{}) error
}

// dbusObject abstracts D-Bus object calls for testing.
type dbusObject interface {
	Call(method string, flags uint, args ...interface{}) dbusCall
	GetProperty(prop string) (interface{}, error)
}

// dbusCall abstracts D-Bus method call results.
type dbusCall interface {
	Store(retvalues ...interface{}) error
}

// NewBridge creates a new AT-SPI bridge.
func NewBridge() *Bridge {
	return &Bridge{}
}

// Register connects to the AT-SPI registry and exports accessible objects.
func (b *Bridge) Register(app BridgeApp) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return fmt.Errorf("bridge is closed")
	}

	b.app = app

	// Try to connect to AT-SPI bus
	conn, err := b.connectATSPI()
	if err != nil {
		// AT-SPI not available, fall back to no-op mode
		return nil
	}
	b.conn = conn

	// Build and export the accessible tree
	b.tree = NewAccessibleTree(app)

	// Register with AT-SPI registry
	if err := b.registerWithATSPI(); err != nil {
		// Registration failed, fall back to no-op mode
		_ = b.conn.Close()
		b.conn = nil
		return nil
	}

	return nil
}

// connectATSPI connects to the AT-SPI D-Bus bus.
func (b *Bridge) connectATSPI() (dbusConn, error) {
	// Get the AT-SPI bus address
	addr, err := getATSPIBusAddress()
	if err != nil {
		return nil, err
	}
	b.busAddr = addr

	// Connect to the AT-SPI bus
	conn, err := connectDBus(addr)
	if err != nil {
		return nil, fmt.Errorf("connect to AT-SPI bus: %w", err)
	}

	return conn, nil
}

// registerWithATSPI registers this application with the AT-SPI registry.
func (b *Bridge) registerWithATSPI() error {
	if b.conn == nil {
		return nil
	}

	// Generate a unique object path for this application
	appName := "fluffyui"
	if b.app != nil {
		appName = b.app.Name()
	}

	// Request a unique name on the bus
	busName := fmt.Sprintf("org.a11y.atspi.%s.%d", appName, getProcessID())
	if err := b.conn.RequestName(busName, 0); err != nil {
		return fmt.Errorf("request bus name: %w", err)
	}

	b.appPath = fmt.Sprintf("/org/a11y/atspi/%s", appName)

	// Export the accessible tree on the bus
	if err := b.exportAccessibleTree(); err != nil {
		return fmt.Errorf("export accessible tree: %w", err)
	}

	// Register with the AT-SPI registry
	registry := b.conn.Object(atspiRegistry, registryPath)
	call := registry.Call(
		atspiRegistry+".RegisterApplication",
		0,
		b.appPath,
	)
	if err := call.Store(&b.uniqueID); err != nil {
		// Registry not available, but that's okay
		b.uniqueID = -1
	}

	return nil
}

// exportAccessibleTree exports all accessible objects on the D-Bus.
func (b *Bridge) exportAccessibleTree() error {
	if b.conn == nil || b.tree == nil {
		return nil
	}

	// Export the root accessible
	root := b.tree.Root()
	if root == nil {
		return nil
	}

	// Export accessible interface
	if err := b.conn.Export(root, b.appPath, atspiAccessible); err != nil {
		return err
	}

	// Export application interface
	if err := b.conn.Export(root, b.appPath, atspiApplication); err != nil {
		return err
	}

	return nil
}

// Announce sends text to the screen reader.
func (b *Bridge) Announce(text string, priority Priority) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed || b.conn == nil || text == "" {
		return nil
	}

	// Map priority to AT-SPI announcement type
	announcementType := "message"
	if priority >= PriorityHigh {
		announcementType = "alert"
	}

	// Send announcement event
	return b.conn.Emit(
		b.appPath,
		atspiEvent+".Object",
		"Announcement",
		announcementType,
		text,
	)
}

// UpdateTree notifies the screen reader that the widget structure changed.
func (b *Bridge) UpdateTree() error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed || b.conn == nil {
		return nil
	}

	// Rebuild the accessible tree
	if b.tree != nil && b.app != nil {
		b.tree.Rebuild(b.app)
	}

	// Emit tree update event
	return b.conn.Emit(
		b.appPath,
		atspiEvent+".Object",
		"ChildrenChanged",
		"add/remove",
		0, // index
	)
}

// NotifyFocusChange informs the screen reader that focus moved.
func (b *Bridge) NotifyFocusChange(widget Accessible) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed || b.conn == nil {
		return nil
	}

	// Find the accessible object for this widget
	path := b.appPath
	if b.tree != nil {
		if nodePath := b.tree.PathFor(widget); nodePath != "" {
			path = b.appPath + nodePath
		}
	}

	// Emit focus event
	return b.conn.Emit(
		path,
		atspiEvent+".Focus",
		"Focus",
		"",
		true,
	)
}

// NotifyValueChange informs the screen reader that a widget's value changed.
func (b *Bridge) NotifyValueChange(widget Accessible, oldValue, newValue string) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed || b.conn == nil {
		return nil
	}

	// Find the accessible object for this widget
	path := b.appPath
	if b.tree != nil {
		if nodePath := b.tree.PathFor(widget); nodePath != "" {
			path = b.appPath + nodePath
		}
	}

	// Emit value change event
	return b.conn.Emit(
		path,
		atspiEvent+".Object",
		"PropertyChange",
		"accessible-value",
		newValue,
	)
}

// NotifyStateChange informs the screen reader that a widget's state changed.
func (b *Bridge) NotifyStateChange(widget Accessible, state string, value bool) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed || b.conn == nil {
		return nil
	}

	// Find the accessible object for this widget
	path := b.appPath
	if b.tree != nil {
		if nodePath := b.tree.PathFor(widget); nodePath != "" {
			path = b.appPath + nodePath
		}
	}

	// Emit state change event
	stateValue := 0
	if value {
		stateValue = 1
	}
	return b.conn.Emit(
		path,
		atspiEvent+".Object",
		"StateChanged",
		state,
		stateValue,
	)
}

// Close disconnects from the AT-SPI bus.
func (b *Bridge) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}
	b.closed = true

	if b.conn != nil {
		// Unregister from AT-SPI registry
		if b.uniqueID >= 0 {
			registry := b.conn.Object(atspiRegistry, registryPath)
			_ = registry.Call(
				atspiRegistry+".DeregisterApplication",
				0,
				b.appPath,
			).Store()
		}

		return b.conn.Close()
	}

	return nil
}

// IsConnected returns true if the bridge is connected to AT-SPI.
func (b *Bridge) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.conn != nil && !b.closed
}
