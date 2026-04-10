//go:build linux

package atspi

import (
	"os"
	"testing"
)

func TestEventConstants(t *testing.T) {
	// Verify event type constants are non-empty
	events := []string{
		EventFocus,
		EventStateChanged,
		EventPropertyChange,
		EventChildrenChange,
		EventTextChanged,
		EventValueChanged,
		EventAnnouncement,
	}

	for _, ev := range events {
		if ev == "" {
			t.Error("found empty event constant")
		}
	}
}

func TestStateChangeConstants(t *testing.T) {
	states := []string{
		StateChangedChecked,
		StateChangedExpanded,
		StateChangedSelected,
		StateChangedFocused,
		StateChangedEnabled,
		StateChangedVisible,
		StateChangedShowing,
		StateChangedActive,
		StateChangedSensitive,
	}

	for _, state := range states {
		if state == "" {
			t.Error("found empty state constant")
		}
	}
}

func TestPropertyConstants(t *testing.T) {
	properties := []string{
		PropertyName,
		PropertyDescription,
		PropertyRole,
		PropertyValue,
		PropertyParent,
	}

	for _, prop := range properties {
		if prop == "" {
			t.Error("found empty property constant")
		}
	}
}

func TestChildrenChangeConstants(t *testing.T) {
	if ChildrenChangeAdd != "add" {
		t.Errorf("ChildrenChangeAdd = %q, want 'add'", ChildrenChangeAdd)
	}
	if ChildrenChangeRemove != "remove" {
		t.Errorf("ChildrenChangeRemove = %q, want 'remove'", ChildrenChangeRemove)
	}
}

func TestNewFocusEvent(t *testing.T) {
	event := NewFocusEvent("/test/path", 1)

	if event.Type != EventFocus {
		t.Errorf("Type = %q, want %q", event.Type, EventFocus)
	}
	if event.Source != "/test/path" {
		t.Errorf("Source = %q, want '/test/path'", event.Source)
	}
	if event.Detail1 != 1 {
		t.Errorf("Detail1 = %d, want 1", event.Detail1)
	}
}

func TestNewStateChangedEvent(t *testing.T) {
	tests := []struct {
		name      string
		stateName string
		value     bool
		wantType  string
		wantDetail int32
	}{
		{"checked true", StateChangedChecked, true, EventStateChanged + StateChangedChecked, 1},
		{"checked false", StateChangedChecked, false, EventStateChanged + StateChangedChecked, 0},
		{"expanded true", StateChangedExpanded, true, EventStateChanged + StateChangedExpanded, 1},
		{"selected false", StateChangedSelected, false, EventStateChanged + StateChangedSelected, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewStateChangedEvent("/test", tt.stateName, tt.value)

			if event.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", event.Type, tt.wantType)
			}
			if event.Detail1 != tt.wantDetail {
				t.Errorf("Detail1 = %d, want %d", event.Detail1, tt.wantDetail)
			}
		})
	}
}

func TestNewPropertyChangeEvent(t *testing.T) {
	event := NewPropertyChangeEvent("/test", PropertyName, "New Name")

	if event.Type != EventPropertyChange+PropertyName {
		t.Errorf("Type = %q, want property change event", event.Type)
	}
	if event.AnyData != "New Name" {
		t.Errorf("AnyData = %v, want 'New Name'", event.AnyData)
	}
}

func TestNewChildrenChangeEvent(t *testing.T) {
	event := NewChildrenChangeEvent("/test", ChildrenChangeAdd, 5)

	if event.Type != EventChildrenChange+ChildrenChangeAdd {
		t.Errorf("Type = %q, want children change add event", event.Type)
	}
	if event.Detail1 != 5 {
		t.Errorf("Detail1 = %d, want 5", event.Detail1)
	}
}

func TestNewAnnouncementEvent(t *testing.T) {
	event := NewAnnouncementEvent("/test", "Hello World", 1)

	if event.Type != EventAnnouncement {
		t.Errorf("Type = %q, want %q", event.Type, EventAnnouncement)
	}
	if event.Detail1 != 1 {
		t.Errorf("Detail1 = %d, want 1", event.Detail1)
	}
	if event.AnyData != "Hello World" {
		t.Errorf("AnyData = %v, want 'Hello World'", event.AnyData)
	}
}

func TestParseEvent(t *testing.T) {
	tests := []struct {
		input      string
		wantCat    string
		wantDetail string
	}{
		{"focus:", "focus", ""},
		{"object:state-changed:", "object", "state-changed"},
		{"object:property-change:accessible-name", "object", "property-change"},
		{"simple", "simple", ""},
		{"", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			cat, detail := ParseEvent(tt.input)
			if cat != tt.wantCat {
				t.Errorf("category = %q, want %q", cat, tt.wantCat)
			}
			if detail != tt.wantDetail {
				t.Errorf("detail = %q, want %q", detail, tt.wantDetail)
			}
		})
	}
}

func TestGetATSPIBusAddress(t *testing.T) {
	// Save and restore environment
	origATSPI := os.Getenv("AT_SPI_BUS_ADDRESS")
	origSession := os.Getenv("DBUS_SESSION_BUS_ADDRESS")
	defer func() {
		os.Setenv("AT_SPI_BUS_ADDRESS", origATSPI)
		os.Setenv("DBUS_SESSION_BUS_ADDRESS", origSession)
	}()

	t.Run("from AT_SPI_BUS_ADDRESS", func(t *testing.T) {
		os.Setenv("AT_SPI_BUS_ADDRESS", "test:atspi")
		addr, err := getATSPIBusAddress()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if addr != "test:atspi" {
			t.Errorf("addr = %q, want 'test:atspi'", addr)
		}
	})

	t.Run("from DBUS_SESSION_BUS_ADDRESS", func(t *testing.T) {
		os.Setenv("AT_SPI_BUS_ADDRESS", "")
		os.Setenv("DBUS_SESSION_BUS_ADDRESS", "test:session")
		addr, err := getATSPIBusAddress()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if addr != "test:session" {
			t.Errorf("addr = %q, want 'test:session'", addr)
		}
	})

	t.Run("default", func(t *testing.T) {
		os.Setenv("AT_SPI_BUS_ADDRESS", "")
		os.Setenv("DBUS_SESSION_BUS_ADDRESS", "")
		addr, err := getATSPIBusAddress()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if addr == "" {
			t.Error("expected non-empty default address")
		}
	})
}

func TestGetProcessID(t *testing.T) {
	pid := getProcessID()
	if pid <= 0 {
		t.Errorf("getProcessID() = %d, expected positive", pid)
	}
	if pid != os.Getpid() {
		t.Errorf("getProcessID() = %d, os.Getpid() = %d", pid, os.Getpid())
	}
}

func TestMockDBusConn(t *testing.T) {
	conn := &mockDBusConn{}

	// All methods should work without error
	if err := conn.Close(); err != nil {
		t.Errorf("Close error: %v", err)
	}

	obj := conn.Object("org.test", "/path")
	if obj == nil {
		t.Fatal("Object returned nil")
	}

	if err := conn.Export(nil, "/path", "org.test"); err != nil {
		t.Errorf("Export error: %v", err)
	}

	if err := conn.RequestName("org.test", 0); err != nil {
		t.Errorf("RequestName error: %v", err)
	}

	if err := conn.Emit("/path", "org.test", "Signal", "arg"); err != nil {
		t.Errorf("Emit error: %v", err)
	}
}

func TestMockDBusObject(t *testing.T) {
	obj := &mockDBusObject{}

	call := obj.Call("Method", 0, "arg")
	if call == nil {
		t.Fatal("Call returned nil")
	}

	prop, err := obj.GetProperty("Prop")
	if err != nil {
		t.Errorf("GetProperty error: %v", err)
	}
	if prop != nil {
		t.Errorf("GetProperty = %v, want nil", prop)
	}
}

func TestMockDBusCall(t *testing.T) {
	call := &mockDBusCall{}

	if err := call.Store(); err != nil {
		t.Errorf("Store error: %v", err)
	}

	var result int
	if err := call.Store(&result); err != nil {
		t.Errorf("Store with arg error: %v", err)
	}
}

func TestConnectDBus(t *testing.T) {
	// connectDBus uses the mock implementation by default
	conn, err := connectDBus("test:address")
	if err != nil {
		t.Errorf("connectDBus error: %v", err)
	}
	if conn == nil {
		t.Fatal("connectDBus returned nil")
	}

	// Verify it's a mock connection
	_ = conn.Close()
}
