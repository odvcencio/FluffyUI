// Package accessibility provides accessibility primitives for widgets.
package accessibility

import (
	"strings"
	"sync"

	"github.com/odvcencio/fluffyui/backend"
)

// Role describes the semantic role of a widget.
type Role string

// Common accessibility roles.
const (
	RoleButton      Role = "button"
	RoleCheckbox    Role = "checkbox"
	RoleRadio       Role = "radio"
	RoleTextbox     Role = "textbox"
	RoleList        Role = "list"
	RoleListItem    Role = "listitem"
	RoleTable       Role = "table"
	RoleRow         Role = "row"
	RoleCell        Role = "cell"
	RoleSlider      Role = "slider"
	RoleTree        Role = "tree"
	RoleTreeItem    Role = "treeitem"
	RoleMenu        Role = "menu"
	RoleMenuItem    Role = "menuitem"
	RoleTab         Role = "tab"
	RoleTabList     Role = "tablist"
	RoleTabPanel    Role = "tabpanel"
	RoleDialog      Role = "dialog"
	RoleAlert       Role = "alert"
	RoleStatus      Role = "status"
	RoleProgressBar Role = "progressbar"
	RoleGroup       Role = "group"
	RoleText        Role = "text"
	RoleChart       Role = "chart"
	RoleWindow      Role = "window"
	RoleApplication Role = "application"
)

// Accessible is implemented by widgets that expose accessibility metadata.
type Accessible interface {
	AccessibleRole() Role
	AccessibleLabel() string
	AccessibleDescription() string
	AccessibleState() StateSet
	AccessibleValue() *ValueInfo
}

// StateSet describes the state of a widget.
type StateSet struct {
	Checked  *bool // nil = not applicable
	Expanded *bool
	Selected bool
	Disabled bool
	ReadOnly bool
	Required bool
	Invalid  bool
}

// Strings returns human-friendly descriptions of the state.
func (s StateSet) Strings() []string {
	var out []string
	if s.Checked != nil {
		if *s.Checked {
			out = append(out, "checked")
		} else {
			out = append(out, "unchecked")
		}
	}
	if s.Expanded != nil {
		if *s.Expanded {
			out = append(out, "expanded")
		} else {
			out = append(out, "collapsed")
		}
	}
	if s.Selected {
		out = append(out, "selected")
	}
	if s.Disabled {
		out = append(out, "disabled")
	}
	if s.ReadOnly {
		out = append(out, "read-only")
	}
	if s.Required {
		out = append(out, "required")
	}
	if s.Invalid {
		out = append(out, "invalid")
	}
	return out
}

// ValueInfo describes a widget's numeric value.
type ValueInfo struct {
	Min     float64
	Max     float64
	Current float64
	Text    string
}

// Announcer publishes accessibility announcements.
type Announcer interface {
	Announce(message string, priority Priority)
	AnnounceChange(widget Accessible)
}

// Priority describes announcement urgency.
type Priority int

const (
	// PriorityPolite waits for current speech to complete.
	PriorityPolite Priority = iota
	// PriorityAssertive interrupts current speech for important messages.
	PriorityAssertive
	// PriorityLow is lower than polite, can be dropped.
	PriorityLow
	// PriorityMedium is equivalent to PriorityPolite.
	PriorityMedium
	// PriorityHigh is between polite and assertive.
	PriorityHigh
	// PriorityUrgent interrupts current speech immediately.
	PriorityUrgent
)

// Bridge provides cross-platform screen reader integration.
// Implementations connect to platform-specific accessibility APIs:
// - Linux: AT-SPI via D-Bus
// - macOS: NSAccessibility
// - Windows: UI Automation
type Bridge interface {
	// Register connects the application to the platform accessibility system.
	// The app parameter provides access to the widget tree.
	Register(app BridgeApp) error

	// Announce sends text to the screen reader.
	// Priority controls whether the message interrupts current speech.
	Announce(text string, priority Priority) error

	// UpdateTree notifies the screen reader that the widget structure changed.
	// Call this after layout changes.
	UpdateTree() error

	// NotifyFocusChange informs the screen reader that focus moved to a new widget.
	NotifyFocusChange(widget Accessible) error

	// NotifyValueChange informs the screen reader that a widget's value changed.
	// For text inputs, sliders, progress bars, etc.
	NotifyValueChange(widget Accessible, oldValue, newValue string) error

	// NotifyStateChange informs the screen reader that a widget's state changed.
	// For checkboxes, expandable items, etc.
	NotifyStateChange(widget Accessible, state string, value bool) error

	// Close disconnects from the platform accessibility system.
	Close() error
}

// BridgeApp provides the bridge with access to application state.
type BridgeApp interface {
	// Name returns the application name.
	Name() string

	// RootAccessible returns the root accessible element.
	RootAccessible() Accessible

	// FocusedAccessible returns the currently focused element.
	FocusedAccessible() Accessible

	// AccessibleAt returns the accessible at the given path.
	// Path is a slash-separated list of indices from root.
	AccessibleAt(path string) Accessible

	// ChildAccessibles returns the children of the given accessible.
	ChildAccessibles(parent Accessible) []Accessible
}

// FocusStyle defines consistent focus rendering.
type FocusStyle struct {
	Indicator    string
	Style        backend.Style
	HighContrast backend.Style
}

// Base is a helper implementation of Accessible.
type Base struct {
	Role        Role
	Label       string
	Description string
	State       StateSet
	Value       *ValueInfo
}

// AccessibleRole returns the current role.
func (b *Base) AccessibleRole() Role {
	if b == nil {
		return ""
	}
	return b.Role
}

// AccessibleLabel returns the current label.
func (b *Base) AccessibleLabel() string {
	if b == nil {
		return ""
	}
	return b.Label
}

// AccessibleDescription returns the current description.
func (b *Base) AccessibleDescription() string {
	if b == nil {
		return ""
	}
	return b.Description
}

// AccessibleState returns the current state set.
func (b *Base) AccessibleState() StateSet {
	if b == nil {
		return StateSet{}
	}
	return b.State
}

// AccessibleValue returns the current value info.
func (b *Base) AccessibleValue() *ValueInfo {
	if b == nil {
		return nil
	}
	return b.Value
}

// SetRole updates the role.
func (b *Base) SetRole(role Role) {
	if b == nil {
		return
	}
	b.Role = role
}

// SetLabel updates the label.
func (b *Base) SetLabel(label string) {
	if b == nil {
		return
	}
	b.Label = label
}

// SetDescription updates the description.
func (b *Base) SetDescription(description string) {
	if b == nil {
		return
	}
	b.Description = description
}

// SetState updates the state.
func (b *Base) SetState(state StateSet) {
	if b == nil {
		return
	}
	b.State = state
}

// SetValue updates the value.
func (b *Base) SetValue(value *ValueInfo) {
	if b == nil {
		return
	}
	b.Value = value
}

// BoolPtr returns a pointer to a bool.
func BoolPtr(value bool) *bool {
	return &value
}

// Announcement captures a published accessibility message.
type Announcement struct {
	Message  string
	Priority Priority
}

// SimpleAnnouncer stores announcements in memory.
type SimpleAnnouncer struct {
	mu        sync.Mutex
	history   []Announcement
	onMessage func(Announcement)
}

// SetOnMessage sets a callback for new announcements.
func (a *SimpleAnnouncer) SetOnMessage(fn func(Announcement)) {
	if a == nil {
		return
	}
	a.mu.Lock()
	a.onMessage = fn
	a.mu.Unlock()
}

// History returns a copy of announcements.
func (a *SimpleAnnouncer) History() []Announcement {
	if a == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.history) == 0 {
		return nil
	}
	out := make([]Announcement, len(a.history))
	copy(out, a.history)
	return out
}

// Announce publishes a message.
func (a *SimpleAnnouncer) Announce(message string, priority Priority) {
	if a == nil {
		return
	}
	msg := strings.TrimSpace(message)
	if msg == "" {
		return
	}
	announcement := Announcement{Message: msg, Priority: priority}
	a.mu.Lock()
	a.history = append(a.history, announcement)
	cb := a.onMessage
	a.mu.Unlock()
	if cb != nil {
		cb(announcement)
	}
}

// AnnounceChange announces the widget state.
func (a *SimpleAnnouncer) AnnounceChange(widget Accessible) {
	message := FormatChange(widget)
	if message == "" {
		return
	}
	a.Announce(message, PriorityPolite)
}

// FormatChange builds a short description of a widget's state.
func FormatChange(widget Accessible) string {
	if widget == nil {
		return ""
	}
	label := strings.TrimSpace(widget.AccessibleLabel())
	role := strings.TrimSpace(string(widget.AccessibleRole()))
	description := strings.TrimSpace(widget.AccessibleDescription())
	state := widget.AccessibleState()

	var parts []string
	if label != "" {
		parts = append(parts, label)
	}
	if role != "" {
		parts = append(parts, role)
	}
	if description != "" {
		parts = append(parts, description)
	}
	if stateParts := state.Strings(); len(stateParts) > 0 {
		parts = append(parts, strings.Join(stateParts, " "))
	}
	if value := widget.AccessibleValue(); value != nil {
		text := strings.TrimSpace(value.Text)
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, ", ")
}

// BridgeAnnouncer wraps a Bridge to implement the Announcer interface.
type BridgeAnnouncer struct {
	bridge Bridge
}

// NewBridgeAnnouncer creates an Announcer that uses the given Bridge.
func NewBridgeAnnouncer(bridge Bridge) *BridgeAnnouncer {
	return &BridgeAnnouncer{bridge: bridge}
}

// Announce sends text to the screen reader.
func (ba *BridgeAnnouncer) Announce(message string, priority Priority) {
	if ba == nil || ba.bridge == nil {
		return
	}
	_ = ba.bridge.Announce(message, priority)
}

// AnnounceChange announces the widget state.
func (ba *BridgeAnnouncer) AnnounceChange(widget Accessible) {
	if ba == nil {
		return
	}
	message := FormatChange(widget)
	if message == "" {
		return
	}
	ba.Announce(message, PriorityPolite)
}
