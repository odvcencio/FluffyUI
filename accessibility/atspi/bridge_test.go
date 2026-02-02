package atspi

import (
	"testing"
)

// mockAccessible implements Accessible for testing.
type mockAccessible struct {
	role        string
	label       string
	description string
	state       StateSet
	value       *ValueInfo
}

func (m *mockAccessible) AccessibleRole() string        { return m.role }
func (m *mockAccessible) AccessibleLabel() string       { return m.label }
func (m *mockAccessible) AccessibleDescription() string { return m.description }
func (m *mockAccessible) AccessibleState() StateSet     { return m.state }
func (m *mockAccessible) AccessibleValue() *ValueInfo   { return m.value }

// mockApp implements BridgeApp for testing.
type mockApp struct {
	name     string
	root     Accessible
	focused  Accessible
	children map[Accessible][]Accessible
}

func (a *mockApp) Name() string {
	return a.name
}

func (a *mockApp) RootAccessible() Accessible {
	return a.root
}

func (a *mockApp) FocusedAccessible() Accessible {
	return a.focused
}

func (a *mockApp) AccessibleAt(path string) Accessible {
	return nil
}

func (a *mockApp) ChildAccessibles(parent Accessible) []Accessible {
	if a.children == nil {
		return nil
	}
	return a.children[parent]
}

func TestNewBridge(t *testing.T) {
	bridge := NewBridge()
	if bridge == nil {
		t.Fatal("NewBridge returned nil")
	}
}

func TestBridgeRegisterNilApp(t *testing.T) {
	bridge := NewBridge()
	err := bridge.Register(nil)
	if err != nil {
		t.Errorf("Register(nil) should not error: %v", err)
	}
}

func TestBridgeRegisterWithApp(t *testing.T) {
	bridge := NewBridge()
	app := &mockApp{
		name: "TestApp",
		root: &mockAccessible{role: "window", label: "Main"},
	}

	err := bridge.Register(app)
	if err != nil {
		t.Errorf("Register failed: %v", err)
	}
}

func TestBridgeAnnounce(t *testing.T) {
	bridge := NewBridge()
	_ = bridge.Register(&mockApp{name: "Test"})

	tests := []struct {
		name     string
		text     string
		priority Priority
	}{
		{"low", "Hello", PriorityLow},
		{"medium", "Hello", PriorityMedium},
		{"high", "Alert!", PriorityHigh},
		{"urgent", "Critical", PriorityUrgent},
		{"empty", "", PriorityLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bridge.Announce(tt.text, tt.priority)
			if err != nil {
				t.Errorf("Announce failed: %v", err)
			}
		})
	}
}

func TestBridgeUpdateTree(t *testing.T) {
	bridge := NewBridge()
	root := &mockAccessible{role: "window", label: "Main"}
	_ = bridge.Register(&mockApp{name: "Test", root: root})

	err := bridge.UpdateTree()
	if err != nil {
		t.Errorf("UpdateTree failed: %v", err)
	}
}

func TestBridgeFocusChange(t *testing.T) {
	bridge := NewBridge()
	_ = bridge.Register(&mockApp{name: "Test"})

	widget := &mockAccessible{role: "button", label: "Submit"}
	err := bridge.NotifyFocusChange(widget)
	if err != nil {
		t.Errorf("NotifyFocusChange failed: %v", err)
	}
}

func TestBridgeValueChange(t *testing.T) {
	bridge := NewBridge()
	_ = bridge.Register(&mockApp{name: "Test"})

	widget := &mockAccessible{role: "slider", label: "Volume"}
	err := bridge.NotifyValueChange(widget, "50", "75")
	if err != nil {
		t.Errorf("NotifyValueChange failed: %v", err)
	}
}

func TestBridgeStateChange(t *testing.T) {
	bridge := NewBridge()
	_ = bridge.Register(&mockApp{name: "Test"})

	widget := &mockAccessible{role: "checkbox", label: "Enabled"}
	err := bridge.NotifyStateChange(widget, "checked", true)
	if err != nil {
		t.Errorf("NotifyStateChange failed: %v", err)
	}
}

func TestBridgeClose(t *testing.T) {
	bridge := NewBridge()
	_ = bridge.Register(&mockApp{name: "Test"})

	err := bridge.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Double close should not error
	err = bridge.Close()
	if err != nil {
		t.Errorf("Double close failed: %v", err)
	}
}

func TestBridgeClosedOperations(t *testing.T) {
	bridge := NewBridge()
	_ = bridge.Register(&mockApp{name: "Test"})
	_ = bridge.Close()

	// Operations after close should not error
	if err := bridge.Announce("Test", PriorityLow); err != nil {
		t.Errorf("Announce after close should not error: %v", err)
	}

	if err := bridge.UpdateTree(); err != nil {
		t.Errorf("UpdateTree after close should not error: %v", err)
	}

	if err := bridge.NotifyFocusChange(nil); err != nil {
		t.Errorf("NotifyFocusChange after close should not error: %v", err)
	}
}

func TestBridgeIsConnected(t *testing.T) {
	bridge := NewBridge()

	// Not connected before register
	if bridge.IsConnected() {
		t.Error("should not be connected before Register")
	}

	_ = bridge.Register(&mockApp{name: "Test"})
	// Note: In test environment, D-Bus may not be available,
	// so IsConnected may return false

	_ = bridge.Close()
	if bridge.IsConnected() {
		t.Error("should not be connected after Close")
	}
}

func TestRoleConversion(t *testing.T) {
	tests := []struct {
		role  string
		atspi uint32
	}{
		{"button", rolePushButton},
		{"checkbox", roleCheckBox},
		{"radio", roleRadioButton},
		{"textbox", roleEntry},
		{"list", roleList},
		{"listitem", roleListItem},
		{"table", roleTable},
		{"row", roleTableRow},
		{"cell", roleTableCell},
		{"slider", roleSlider},
		{"tree", roleTree},
		{"treeitem", roleTreeItem},
		{"menu", roleMenu},
		{"menuitem", roleMenuItem},
		{"tab", rolePageTab},
		{"tablist", rolePageTabList},
		{"tabpanel", rolePanel},
		{"dialog", roleDialog},
		{"alert", roleAlert},
		{"status", roleStatusBar},
		{"progressbar", roleProgressBar},
		{"group", roleFrame},
		{"text", roleLabel},
		{"chart", roleImage},
		{"application", roleApplication},
		{"window", roleWindow},
		{"unknown", roleFiller},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			got := roleToATSPI(tt.role)
			if got != tt.atspi {
				t.Errorf("roleToATSPI(%s) = %d, want %d", tt.role, got, tt.atspi)
			}
		})
	}
}

func TestStateConversion(t *testing.T) {
	t.Run("default states", func(t *testing.T) {
		state := StateSet{}
		bits := stateToATSPI(state)
		if len(bits) != 2 {
			t.Fatalf("expected 2 state words, got %d", len(bits))
		}
		// Should have visible (in high word) and showing (in low word) by default
		if bits[1]&(1<<(stateVisible-32)) == 0 {
			t.Error("should have visible state")
		}
		if bits[0]&(1<<stateShowing) == 0 {
			t.Error("should have showing state")
		}
	})

	t.Run("enabled", func(t *testing.T) {
		state := StateSet{Disabled: false}
		bits := stateToATSPI(state)
		if bits[0]&(1<<stateEnabled) == 0 {
			t.Error("should have enabled state")
		}
		if bits[0]&(1<<stateSensitive) == 0 {
			t.Error("should have sensitive state")
		}
	})

	t.Run("disabled", func(t *testing.T) {
		state := StateSet{Disabled: true}
		bits := stateToATSPI(state)
		if bits[0]&(1<<stateEnabled) != 0 {
			t.Error("should not have enabled state when disabled")
		}
	})

	t.Run("checked", func(t *testing.T) {
		checked := true
		state := StateSet{Checked: &checked}
		bits := stateToATSPI(state)
		if bits[0]&(1<<stateChecked) == 0 {
			t.Error("should have checked state")
		}
	})

	t.Run("unchecked", func(t *testing.T) {
		checked := false
		state := StateSet{Checked: &checked}
		bits := stateToATSPI(state)
		if bits[0]&(1<<stateChecked) != 0 {
			t.Error("should not have checked state when unchecked")
		}
	})

	t.Run("expanded", func(t *testing.T) {
		expanded := true
		state := StateSet{Expanded: &expanded}
		bits := stateToATSPI(state)
		if bits[0]&(1<<stateExpanded) == 0 {
			t.Error("should have expanded state")
		}
	})

	t.Run("selected", func(t *testing.T) {
		state := StateSet{Selected: true}
		bits := stateToATSPI(state)
		if bits[0]&(1<<stateSelected) == 0 {
			t.Error("should have selected state")
		}
	})

	t.Run("editable", func(t *testing.T) {
		state := StateSet{ReadOnly: false}
		bits := stateToATSPI(state)
		if bits[0]&(1<<stateEditable) == 0 {
			t.Error("should have editable state")
		}
	})

	t.Run("readonly", func(t *testing.T) {
		state := StateSet{ReadOnly: true}
		bits := stateToATSPI(state)
		if bits[0]&(1<<stateEditable) != 0 {
			t.Error("should not have editable state when readonly")
		}
	})

	t.Run("required", func(t *testing.T) {
		state := StateSet{Required: true}
		bits := stateToATSPI(state)
		if bits[0]&(1<<stateRequired) == 0 {
			t.Error("should have required state")
		}
	})

	t.Run("invalid", func(t *testing.T) {
		state := StateSet{Invalid: true}
		bits := stateToATSPI(state)
		// Invalid is in high word
		if bits[1]&(1<<(stateInvalidEntry-32)) == 0 {
			t.Error("should have invalid state")
		}
	})
}
