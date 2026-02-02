package accessibility

import (
	"strings"
	"testing"
)

func TestFormatChange(t *testing.T) {
	checked := true
	base := &Base{
		Role:  RoleButton,
		Label: "Save",
		State: StateSet{
			Checked:  &checked,
			Disabled: true,
		},
	}
	msg := FormatChange(base)
	if msg == "" {
		t.Fatalf("expected message, got empty")
	}
	if !containsAll(msg, []string{"Save", "button"}) {
		t.Fatalf("unexpected message: %q", msg)
	}
}

func TestSimpleAnnouncer(t *testing.T) {
	a := &SimpleAnnouncer{}
	a.Announce("hello", PriorityPolite)
	history := a.History()
	if len(history) != 1 {
		t.Fatalf("expected 1 announcement, got %d", len(history))
	}
	if history[0].Message != "hello" {
		t.Fatalf("unexpected message %q", history[0].Message)
	}
}

func containsAll(message string, parts []string) bool {
	for _, part := range parts {
		if !strings.Contains(message, part) {
			return false
		}
	}
	return true
}

// MockBridge implements Bridge for testing.
type MockBridge struct {
	registered       bool
	closed           bool
	announcements    []string
	focusChanges     []Accessible
	valueChanges     []valueChange
	stateChanges     []stateChange
	treeUpdates      int
	registerError    error
	announceError    error
	focusError       error
	valueError       error
	stateError       error
	treeError        error
	closeError       error
}

type valueChange struct {
	widget   Accessible
	oldValue string
	newValue string
}

type stateChange struct {
	widget Accessible
	state  string
	value  bool
}

func (m *MockBridge) Register(app BridgeApp) error {
	if m.registerError != nil {
		return m.registerError
	}
	m.registered = true
	return nil
}

func (m *MockBridge) Announce(text string, priority Priority) error {
	if m.announceError != nil {
		return m.announceError
	}
	if text != "" {
		m.announcements = append(m.announcements, text)
	}
	return nil
}

func (m *MockBridge) UpdateTree() error {
	if m.treeError != nil {
		return m.treeError
	}
	m.treeUpdates++
	return nil
}

func (m *MockBridge) NotifyFocusChange(widget Accessible) error {
	if m.focusError != nil {
		return m.focusError
	}
	m.focusChanges = append(m.focusChanges, widget)
	return nil
}

func (m *MockBridge) NotifyValueChange(widget Accessible, oldValue, newValue string) error {
	if m.valueError != nil {
		return m.valueError
	}
	m.valueChanges = append(m.valueChanges, valueChange{widget, oldValue, newValue})
	return nil
}

func (m *MockBridge) NotifyStateChange(widget Accessible, state string, value bool) error {
	if m.stateError != nil {
		return m.stateError
	}
	m.stateChanges = append(m.stateChanges, stateChange{widget, state, value})
	return nil
}

func (m *MockBridge) Close() error {
	if m.closeError != nil {
		return m.closeError
	}
	m.closed = true
	return nil
}

// MockBridgeApp implements BridgeApp for testing.
type MockBridgeApp struct {
	name     string
	root     Accessible
	focused  Accessible
	children map[Accessible][]Accessible
}

func (a *MockBridgeApp) Name() string {
	return a.name
}

func (a *MockBridgeApp) RootAccessible() Accessible {
	return a.root
}

func (a *MockBridgeApp) FocusedAccessible() Accessible {
	return a.focused
}

func (a *MockBridgeApp) AccessibleAt(path string) Accessible {
	return nil
}

func (a *MockBridgeApp) ChildAccessibles(parent Accessible) []Accessible {
	if a.children == nil {
		return nil
	}
	return a.children[parent]
}

func TestMockBridgeRegister(t *testing.T) {
	bridge := &MockBridge{}
	app := &MockBridgeApp{name: "TestApp"}

	err := bridge.Register(app)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if !bridge.registered {
		t.Fatal("expected bridge to be registered")
	}
}

func TestMockBridgeAnnounce(t *testing.T) {
	bridge := &MockBridge{}

	tests := []struct {
		name     string
		text     string
		priority Priority
		wantLen  int
	}{
		{"polite", "Hello", PriorityPolite, 1},
		{"assertive", "Alert!", PriorityAssertive, 2},
		{"urgent", "Critical", PriorityUrgent, 3},
		{"empty ignored", "", PriorityPolite, 3},
		{"low", "Low priority", PriorityLow, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = bridge.Announce(tt.text, tt.priority)
			if len(bridge.announcements) != tt.wantLen {
				t.Errorf("expected %d announcements, got %d", tt.wantLen, len(bridge.announcements))
			}
		})
	}
}

func TestMockBridgeFocusChange(t *testing.T) {
	bridge := &MockBridge{}
	widget := &Base{Role: RoleButton, Label: "Submit"}

	err := bridge.NotifyFocusChange(widget)
	if err != nil {
		t.Fatalf("NotifyFocusChange failed: %v", err)
	}
	if len(bridge.focusChanges) != 1 {
		t.Fatalf("expected 1 focus change, got %d", len(bridge.focusChanges))
	}
	if bridge.focusChanges[0].AccessibleLabel() != "Submit" {
		t.Errorf("unexpected widget label: %s", bridge.focusChanges[0].AccessibleLabel())
	}
}

func TestMockBridgeValueChange(t *testing.T) {
	bridge := &MockBridge{}
	widget := &Base{Role: RoleSlider, Label: "Volume"}

	err := bridge.NotifyValueChange(widget, "50%", "75%")
	if err != nil {
		t.Fatalf("NotifyValueChange failed: %v", err)
	}
	if len(bridge.valueChanges) != 1 {
		t.Fatalf("expected 1 value change, got %d", len(bridge.valueChanges))
	}
	vc := bridge.valueChanges[0]
	if vc.oldValue != "50%" || vc.newValue != "75%" {
		t.Errorf("unexpected values: old=%s, new=%s", vc.oldValue, vc.newValue)
	}
}

func TestMockBridgeStateChange(t *testing.T) {
	bridge := &MockBridge{}
	widget := &Base{Role: RoleCheckbox, Label: "Enabled"}

	err := bridge.NotifyStateChange(widget, "checked", true)
	if err != nil {
		t.Fatalf("NotifyStateChange failed: %v", err)
	}
	if len(bridge.stateChanges) != 1 {
		t.Fatalf("expected 1 state change, got %d", len(bridge.stateChanges))
	}
	sc := bridge.stateChanges[0]
	if sc.state != "checked" || !sc.value {
		t.Errorf("unexpected state: %s=%v", sc.state, sc.value)
	}
}

func TestMockBridgeUpdateTree(t *testing.T) {
	bridge := &MockBridge{}

	for i := 0; i < 3; i++ {
		err := bridge.UpdateTree()
		if err != nil {
			t.Fatalf("UpdateTree failed: %v", err)
		}
	}
	if bridge.treeUpdates != 3 {
		t.Errorf("expected 3 tree updates, got %d", bridge.treeUpdates)
	}
}

func TestMockBridgeClose(t *testing.T) {
	bridge := &MockBridge{}

	err := bridge.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	if !bridge.closed {
		t.Fatal("expected bridge to be closed")
	}
}

func TestBridgeAnnouncerWrapper(t *testing.T) {
	mock := &MockBridge{}
	announcer := NewBridgeAnnouncer(mock)

	announcer.Announce("Test message", PriorityPolite)
	if len(mock.announcements) != 1 {
		t.Fatalf("expected 1 announcement, got %d", len(mock.announcements))
	}
	if mock.announcements[0] != "Test message" {
		t.Errorf("unexpected message: %s", mock.announcements[0])
	}
}

func TestBridgeAnnouncerChange(t *testing.T) {
	mock := &MockBridge{}
	announcer := NewBridgeAnnouncer(mock)

	widget := &Base{
		Role:  RoleButton,
		Label: "Click me",
	}

	announcer.AnnounceChange(widget)
	if len(mock.announcements) != 1 {
		t.Fatalf("expected 1 announcement, got %d", len(mock.announcements))
	}
	if !strings.Contains(mock.announcements[0], "Click me") {
		t.Errorf("expected message to contain 'Click me', got: %s", mock.announcements[0])
	}
}

func TestBridgeAnnouncerNilBridge(t *testing.T) {
	announcer := NewBridgeAnnouncer(nil)
	// Should not panic
	announcer.Announce("Test", PriorityPolite)
	announcer.AnnounceChange(&Base{Label: "Test"})
}

func TestBridgeAnnouncerNilWidget(t *testing.T) {
	mock := &MockBridge{}
	announcer := NewBridgeAnnouncer(mock)
	// Should not add an announcement for nil widget
	announcer.AnnounceChange(nil)
	if len(mock.announcements) != 0 {
		t.Errorf("expected no announcements for nil widget, got %d", len(mock.announcements))
	}
}

func TestNewBridge(t *testing.T) {
	bridge := NewBridge()
	if bridge == nil {
		t.Fatal("NewBridge returned nil")
	}

	// Should work without errors on any platform
	err := bridge.Register(nil)
	if err != nil {
		t.Errorf("Register with nil app failed: %v", err)
	}

	err = bridge.Announce("Test", PriorityPolite)
	if err != nil {
		t.Errorf("Announce failed: %v", err)
	}

	err = bridge.UpdateTree()
	if err != nil {
		t.Errorf("UpdateTree failed: %v", err)
	}

	err = bridge.NotifyFocusChange(nil)
	if err != nil {
		t.Errorf("NotifyFocusChange failed: %v", err)
	}

	err = bridge.NotifyValueChange(nil, "", "")
	if err != nil {
		t.Errorf("NotifyValueChange failed: %v", err)
	}

	err = bridge.NotifyStateChange(nil, "", false)
	if err != nil {
		t.Errorf("NotifyStateChange failed: %v", err)
	}

	err = bridge.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

func TestPriorityConstants(t *testing.T) {
	// Verify priority ordering makes sense
	if PriorityPolite > PriorityAssertive {
		t.Error("PriorityPolite should be less than PriorityAssertive")
	}
	if PriorityHigh > PriorityUrgent {
		t.Error("PriorityHigh should be less than PriorityUrgent")
	}
}

func TestRoleConstants(t *testing.T) {
	roles := []Role{
		RoleButton, RoleCheckbox, RoleRadio, RoleTextbox,
		RoleList, RoleListItem, RoleTable, RoleRow, RoleCell,
		RoleSlider, RoleTree, RoleTreeItem, RoleMenu, RoleMenuItem,
		RoleTab, RoleTabList, RoleTabPanel, RoleDialog, RoleAlert,
		RoleStatus, RoleProgressBar, RoleGroup, RoleText, RoleChart,
		RoleWindow, RoleApplication,
	}

	for _, role := range roles {
		if role == "" {
			t.Error("found empty role constant")
		}
	}
}

func TestStateSetWithNilPointers(t *testing.T) {
	state := StateSet{
		Selected: true,
		Disabled: false,
	}

	strs := state.Strings()
	if len(strs) != 1 {
		t.Errorf("expected 1 state string, got %d: %v", len(strs), strs)
	}
	if strs[0] != "selected" {
		t.Errorf("expected 'selected', got %s", strs[0])
	}
}

func TestStateSetAllStates(t *testing.T) {
	checked := true
	expanded := true
	state := StateSet{
		Checked:  &checked,
		Expanded: &expanded,
		Selected: true,
		Disabled: true,
		ReadOnly: true,
		Required: true,
		Invalid:  true,
	}

	strs := state.Strings()
	expected := []string{"checked", "expanded", "selected", "disabled", "read-only", "required", "invalid"}

	if len(strs) != len(expected) {
		t.Errorf("expected %d state strings, got %d: %v", len(expected), len(strs), strs)
	}

	for i, exp := range expected {
		if strs[i] != exp {
			t.Errorf("state[%d]: expected %s, got %s", i, exp, strs[i])
		}
	}
}

func TestBaseNilSafety(t *testing.T) {
	var base *Base

	// All methods should handle nil receiver gracefully
	if base.AccessibleRole() != "" {
		t.Error("nil Base.AccessibleRole should return empty string")
	}
	if base.AccessibleLabel() != "" {
		t.Error("nil Base.AccessibleLabel should return empty string")
	}
	if base.AccessibleDescription() != "" {
		t.Error("nil Base.AccessibleDescription should return empty string")
	}
	state := base.AccessibleState()
	if state.Selected || state.Disabled {
		t.Error("nil Base.AccessibleState should return zero StateSet")
	}
	if base.AccessibleValue() != nil {
		t.Error("nil Base.AccessibleValue should return nil")
	}

	// Setters should not panic
	base.SetRole(RoleButton)
	base.SetLabel("Test")
	base.SetDescription("Desc")
	base.SetState(StateSet{Selected: true})
	base.SetValue(&ValueInfo{Text: "100"})
}

func TestSimpleAnnouncerNilSafety(t *testing.T) {
	var a *SimpleAnnouncer

	// All methods should handle nil receiver gracefully
	a.Announce("test", PriorityPolite)
	a.AnnounceChange(&Base{Label: "Test"})
	a.SetOnMessage(func(Announcement) {})

	history := a.History()
	if history != nil {
		t.Error("nil SimpleAnnouncer.History should return nil")
	}
}

func TestFormatChangeNil(t *testing.T) {
	result := FormatChange(nil)
	if result != "" {
		t.Errorf("FormatChange(nil) should return empty string, got %q", result)
	}
}

func TestFormatChangeComplete(t *testing.T) {
	checked := true
	base := &Base{
		Role:        RoleCheckbox,
		Label:       "Accept Terms",
		Description: "You must accept the terms to continue",
		State:       StateSet{Checked: &checked},
		Value:       &ValueInfo{Text: "Selected"},
	}

	result := FormatChange(base)

	expected := []string{"Accept Terms", "checkbox", "must accept", "checked", "Selected"}
	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("FormatChange result missing %q: %s", exp, result)
		}
	}
}

func TestMockBridgeAppInterface(t *testing.T) {
	root := &Base{Role: RoleWindow, Label: "Main Window"}
	child1 := &Base{Role: RoleButton, Label: "Button 1"}
	child2 := &Base{Role: RoleButton, Label: "Button 2"}

	app := &MockBridgeApp{
		name:    "TestApp",
		root:    root,
		focused: child1,
		children: map[Accessible][]Accessible{
			root: {child1, child2},
		},
	}

	if app.Name() != "TestApp" {
		t.Errorf("expected name 'TestApp', got %s", app.Name())
	}

	if app.RootAccessible() != root {
		t.Error("RootAccessible returned wrong value")
	}

	if app.FocusedAccessible() != child1 {
		t.Error("FocusedAccessible returned wrong value")
	}

	children := app.ChildAccessibles(root)
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}

	// No children for non-container
	children = app.ChildAccessibles(child1)
	if len(children) != 0 {
		t.Errorf("expected 0 children for button, got %d", len(children))
	}
}
