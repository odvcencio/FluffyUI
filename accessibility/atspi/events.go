//go:build linux

package atspi

import (
	"os"
	"strconv"
	"strings"
)

// Event types for AT-SPI
const (
	EventFocus          = "focus:"
	EventStateChanged   = "object:state-changed:"
	EventPropertyChange = "object:property-change:"
	EventChildrenChange = "object:children-changed:"
	EventTextChanged    = "object:text-changed:"
	EventValueChanged   = "object:value-changed:"
	EventAnnouncement   = "screen-reader:announcement:"
)

// State change event names
const (
	StateChangedChecked   = "checked"
	StateChangedExpanded  = "expanded"
	StateChangedSelected  = "selected"
	StateChangedFocused   = "focused"
	StateChangedEnabled   = "enabled"
	StateChangedVisible   = "visible"
	StateChangedShowing   = "showing"
	StateChangedActive    = "active"
	StateChangedSensitive = "sensitive"
)

// Property change event names
const (
	PropertyName        = "accessible-name"
	PropertyDescription = "accessible-description"
	PropertyRole        = "accessible-role"
	PropertyValue       = "accessible-value"
	PropertyParent      = "accessible-parent"
)

// Children change event detail
const (
	ChildrenChangeAdd    = "add"
	ChildrenChangeRemove = "remove"
)

// Event represents an AT-SPI event.
type Event struct {
	Type     string
	Source   string
	Detail1  int32
	Detail2  int32
	Sender   string
	AppPath  string
	AnyData  interface{}
}

// NewFocusEvent creates a focus event.
func NewFocusEvent(sourcePath string, detail int32) Event {
	return Event{
		Type:    EventFocus,
		Source:  sourcePath,
		Detail1: detail,
	}
}

// NewStateChangedEvent creates a state change event.
func NewStateChangedEvent(sourcePath, stateName string, value bool) Event {
	detail := int32(0)
	if value {
		detail = 1
	}
	return Event{
		Type:    EventStateChanged + stateName,
		Source:  sourcePath,
		Detail1: detail,
	}
}

// NewPropertyChangeEvent creates a property change event.
func NewPropertyChangeEvent(sourcePath, propertyName string, value interface{}) Event {
	return Event{
		Type:    EventPropertyChange + propertyName,
		Source:  sourcePath,
		AnyData: value,
	}
}

// NewChildrenChangeEvent creates a children change event.
func NewChildrenChangeEvent(sourcePath, changeType string, index int32) Event {
	return Event{
		Type:    EventChildrenChange + changeType,
		Source:  sourcePath,
		Detail1: index,
	}
}

// NewAnnouncementEvent creates an announcement event.
func NewAnnouncementEvent(sourcePath, message string, priority int32) Event {
	return Event{
		Type:    EventAnnouncement,
		Source:  sourcePath,
		Detail1: priority,
		AnyData: message,
	}
}

// getATSPIBusAddress returns the AT-SPI bus address.
func getATSPIBusAddress() (string, error) {
	// First check the environment variable
	if addr := os.Getenv("AT_SPI_BUS_ADDRESS"); addr != "" {
		return addr, nil
	}

	// Try to read from the X11 property
	addr, err := getATSPIAddressFromX11()
	if err == nil && addr != "" {
		return addr, nil
	}

	// Try the default session bus
	if addr := os.Getenv("DBUS_SESSION_BUS_ADDRESS"); addr != "" {
		return addr, nil
	}

	// Default Unix socket
	uid := os.Getuid()
	return "unix:path=/run/user/" + strconv.Itoa(uid) + "/at-spi/bus", nil
}

// getATSPIAddressFromX11 tries to read the AT-SPI bus address from X11.
func getATSPIAddressFromX11() (string, error) {
	// This would normally use xcb or xlib to read the AT_SPI_BUS property
	// from the X root window. For now, return empty to fall back to env.
	return "", nil
}

// getProcessID returns the current process ID.
func getProcessID() int {
	return os.Getpid()
}

// connectDBus creates a D-Bus connection to the given address.
// This is a placeholder that will be replaced by the real implementation
// when building with the dbus package.
func connectDBus(addr string) (dbusConn, error) {
	// Try to connect using the real D-Bus package
	return connectRealDBus(addr)
}

// connectRealDBus is the actual D-Bus connection implementation.
// It is defined in dbus_linux.go with the real D-Bus connection logic.
var connectRealDBus = func(addr string) (dbusConn, error) {
	// Default stub implementation returns a mock connection
	// This allows the code to compile without D-Bus support
	return &mockDBusConn{}, nil
}

// mockDBusConn is a no-op D-Bus connection for when D-Bus is not available.
type mockDBusConn struct{}

func (c *mockDBusConn) Close() error { return nil }

func (c *mockDBusConn) Object(dest, path string) dbusObject {
	return &mockDBusObject{}
}

func (c *mockDBusConn) Export(v interface{}, path, iface string) error {
	return nil
}

func (c *mockDBusConn) RequestName(name string, flags uint) error {
	return nil
}

func (c *mockDBusConn) Emit(path, iface, signal string, values ...interface{}) error {
	return nil
}

// mockDBusObject is a no-op D-Bus object for when D-Bus is not available.
type mockDBusObject struct{}

func (o *mockDBusObject) Call(method string, flags uint, args ...interface{}) dbusCall {
	return &mockDBusCall{}
}

func (o *mockDBusObject) GetProperty(prop string) (interface{}, error) {
	return nil, nil
}

// mockDBusCall is a no-op D-Bus call for when D-Bus is not available.
type mockDBusCall struct{}

func (c *mockDBusCall) Store(retvalues ...interface{}) error {
	return nil
}

// ParseEvent parses an AT-SPI event type string.
func ParseEvent(eventType string) (category string, detail string) {
	parts := strings.SplitN(eventType, ":", 3)
	if len(parts) >= 1 {
		category = parts[0]
	}
	if len(parts) >= 2 {
		detail = parts[1]
	}
	return
}
