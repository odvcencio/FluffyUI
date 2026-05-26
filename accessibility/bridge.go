// Package accessibility provides accessibility primitives for widgets.
package accessibility

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"m31labs.dev/fluffyui/backend"
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

	// WAI-ARIA 1.2 additional roles
	RoleCombobox   Role = "combobox"
	RoleSwitch     Role = "switch"
	RoleSpinButton Role = "spinbutton"
	RoleHeading    Role = "heading"
	RoleLink       Role = "link"
	RoleSeparator  Role = "separator"
	RoleLog        Role = "log"
	RoleTimer      Role = "timer"
	RoleFeed       Role = "feed"
	RoleToolbar    Role = "toolbar"
	RoleSearchbox  Role = "searchbox"
	RoleNone       Role = "none"
	RoleImg        Role = "img"
	RoleNote       Role = "note"
	RoleScrollbar  Role = "scrollbar"

	// WAI-ARIA 1.2 composite and structural roles
	RoleListbox          Role = "listbox"
	RoleOption           Role = "option"
	RoleRadioGroup       Role = "radiogroup"
	RoleGrid             Role = "grid"
	RoleGridCell         Role = "gridcell"
	RoleColumnHeader     Role = "columnheader"
	RoleRowHeader        Role = "rowheader"
	RoleRowGroup         Role = "rowgroup"
	RoleAlertDialog      Role = "alertdialog"
	RoleMenuItemCheckbox Role = "menuitemcheckbox"
	RoleMenuItemRadio    Role = "menuitemradio"
	RoleMenuBar          Role = "menubar"
	RoleTreeGrid         Role = "treegrid"
	RoleDocument         Role = "document"
	RoleMarquee          Role = "marquee"
	RolePresentation     Role = "presentation"
	RoleTooltip          Role = "tooltip"
	RoleMeter            Role = "meter"

	// WAI-ARIA landmark roles
	RoleForm          Role = "form"
	RoleNavigation    Role = "navigation"
	RoleComplementary Role = "complementary"

	// WAI-ARIA 1.3 roles
	RoleComment    Role = "comment"
	RoleMark       Role = "mark"
	RoleSuggestion Role = "suggestion"
	RoleCode       Role = "code"
	RoleTime       Role = "time"
	RoleImage      Role = "image"
)

// Live describes how content changes are announced to assistive technology.
// Mirrors aria-live: off (default), polite (wait for idle), assertive (interrupt).
type Live string

const (
	// LiveOff suppresses automatic announcements (default).
	LiveOff Live = ""
	// LivePolite announces changes when the user is idle.
	LivePolite Live = "polite"
	// LiveAssertive interrupts current speech to announce changes.
	LiveAssertive Live = "assertive"
)

// Relevant describes which mutations in a live region trigger announcements.
// Mirrors aria-relevant: additions, removals, text, all.
type Relevant string

const (
	// RelevantAdditions announces new children added to the region.
	RelevantAdditions Relevant = "additions"
	// RelevantRemovals announces children removed from the region.
	RelevantRemovals Relevant = "removals"
	// RelevantText announces text content changes.
	RelevantText Relevant = "text"
	// RelevantAll announces all mutations.
	RelevantAll Relevant = "all"
)

// Landmark describes the structural significance of a widget for navigation.
// Mirrors ARIA landmark roles: navigation, main, search, form, banner, etc.
type Landmark string

const (
	LandmarkNone          Landmark = ""
	LandmarkNavigation    Landmark = "navigation"
	LandmarkMain          Landmark = "main"
	LandmarkSearch        Landmark = "search"
	LandmarkForm          Landmark = "form"
	LandmarkBanner        Landmark = "banner"
	LandmarkContentInfo   Landmark = "contentinfo"
	LandmarkRegion        Landmark = "region"
	LandmarkComplementary Landmark = "complementary"
)

// Accessible is implemented by widgets that expose accessibility metadata.
type Accessible interface {
	// Identity
	AccessibleRole() Role
	AccessibleLabel() string
	AccessibleDescription() string
	AccessibleState() StateSet
	AccessibleValue() *ValueInfo

	// Live regions — mirrors aria-live, aria-relevant, aria-atomic
	AccessibleLive() Live
	AccessibleRelevant() Relevant
	AccessibleAtomic() bool

	// Landmarks — mirrors ARIA landmark roles
	AccessibleLandmark() Landmark

	// Relationships — mirrors aria-labelledby, aria-describedby, aria-controls, aria-owns, aria-flowto
	AccessibleLabelledBy() string
	AccessibleDescribedBy() string
	AccessibleControls() string
	AccessibleOwns() string
	AccessibleFlowTo() string

	// WAI-ARIA 1.2/1.3 properties
	AccessibleLevel() int
	AccessibleOrientation() string
	AccessibleActiveDescendant() string
	AccessiblePosInSet() int
	AccessibleSetSize() int
	AccessibleHasPopup() string
	AccessibleErrorMessage() string
	AccessibleCurrent() string
	AccessibleAutocomplete() string
	AccessiblePlaceholder() string
	AccessibleSort() string
	AccessibleKeyShortcuts() string
	AccessibleDetails() string
	AccessibleRoleDescription() string
}

// StateSet describes the state of a widget.
type StateSet struct {
	Checked         *bool // nil = not applicable
	Expanded        *bool
	Pressed         *bool // nil = not applicable (tri-state, for toggle buttons)
	Selected        bool
	Disabled        bool
	ReadOnly        bool
	Required        bool
	Invalid         bool
	Hidden          bool
	Busy            bool
	Modal           bool
	Multiline       bool // aria-multiline
	Multiselectable bool // aria-multiselectable
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
	if s.Pressed != nil {
		if *s.Pressed {
			out = append(out, "pressed")
		} else {
			out = append(out, "not pressed")
		}
	}
	if s.Hidden {
		out = append(out, "hidden")
	}
	if s.Busy {
		out = append(out, "busy")
	}
	if s.Modal {
		out = append(out, "modal")
	}
	if s.Multiline {
		out = append(out, "multiline")
	}
	if s.Multiselectable {
		out = append(out, "multiselectable")
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

	// Live region behavior (aria-live, aria-relevant, aria-atomic)
	Live     Live
	Relevant Relevant
	Atomic   bool

	// Relationships (aria-labelledby, aria-describedby, aria-controls, aria-owns, aria-flowto)
	LabelledBy  string
	DescribedBy string
	Controls    string
	Owns        string
	FlowTo      string

	// Landmark (aria landmark roles)
	Landmark Landmark

	// WAI-ARIA 1.2/1.3 properties
	Level            int    // aria-level (heading level, tree depth)
	Orientation      string // aria-orientation (horizontal/vertical)
	ActiveDescendant string // aria-activedescendant (ID of focused descendant)
	PosInSet         int    // aria-posinset (position in set, 1-based)
	SetSize          int    // aria-setsize (total items in set)
	HasPopup         string // aria-haspopup (false/true/menu/listbox/tree/grid/dialog)
	ErrorMessage     string // aria-errormessage (ID of error element)
	Current          string // aria-current (page/step/location/date/time/true)
	Autocomplete     string // aria-autocomplete (inline/list/both/none)
	Placeholder      string // aria-placeholder
	Sort             string // aria-sort (ascending/descending/none/other)
	KeyShortcuts     string // aria-keyshortcuts
	Details          string // aria-details (ID of detailed description)
	RoleDescription  string // aria-roledescription (custom role label for AT)
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

// AccessibleLive returns the live region behavior.
func (b *Base) AccessibleLive() Live {
	if b == nil {
		return LiveOff
	}
	return b.Live
}

// AccessibleRelevant returns which mutations trigger announcements.
func (b *Base) AccessibleRelevant() Relevant {
	if b == nil {
		return ""
	}
	return b.Relevant
}

// AccessibleAtomic returns whether the entire region is announced on change.
func (b *Base) AccessibleAtomic() bool {
	if b == nil {
		return false
	}
	return b.Atomic
}

// AccessibleLandmark returns the landmark designation.
func (b *Base) AccessibleLandmark() Landmark {
	if b == nil {
		return LandmarkNone
	}
	return b.Landmark
}

// AccessibleLabelledBy returns the ID of the widget that labels this one.
func (b *Base) AccessibleLabelledBy() string {
	if b == nil {
		return ""
	}
	return b.LabelledBy
}

// AccessibleDescribedBy returns the ID of the widget that describes this one.
func (b *Base) AccessibleDescribedBy() string {
	if b == nil {
		return ""
	}
	return b.DescribedBy
}

// AccessibleControls returns the ID of the widget this one controls.
func (b *Base) AccessibleControls() string {
	if b == nil {
		return ""
	}
	return b.Controls
}

// AccessibleOwns returns the ID of the widget this one owns.
func (b *Base) AccessibleOwns() string {
	if b == nil {
		return ""
	}
	return b.Owns
}

// AccessibleFlowTo returns the ID of the next widget in reading order.
func (b *Base) AccessibleFlowTo() string {
	if b == nil {
		return ""
	}
	return b.FlowTo
}

// SetLive updates the live region behavior.
func (b *Base) SetLive(live Live) {
	if b == nil {
		return
	}
	b.Live = live
}

// SetRelevant updates which mutations trigger announcements.
func (b *Base) SetRelevant(relevant Relevant) {
	if b == nil {
		return
	}
	b.Relevant = relevant
}

// SetAtomic updates whether the entire region is announced on change.
func (b *Base) SetAtomic(atomic bool) {
	if b == nil {
		return
	}
	b.Atomic = atomic
}

// SetLandmark updates the landmark designation.
func (b *Base) SetLandmark(landmark Landmark) {
	if b == nil {
		return
	}
	b.Landmark = landmark
}

// SetLabelledBy updates the ID of the widget that labels this one.
func (b *Base) SetLabelledBy(id string) {
	if b == nil {
		return
	}
	b.LabelledBy = id
}

// SetDescribedBy updates the ID of the widget that describes this one.
func (b *Base) SetDescribedBy(id string) {
	if b == nil {
		return
	}
	b.DescribedBy = id
}

// SetControls updates the ID of the widget this one controls.
func (b *Base) SetControls(id string) {
	if b == nil {
		return
	}
	b.Controls = id
}

// SetOwns updates the ID of the widget this one owns.
func (b *Base) SetOwns(id string) {
	if b == nil {
		return
	}
	b.Owns = id
}

// SetFlowTo updates the ID of the next widget in reading order.
func (b *Base) SetFlowTo(id string) {
	if b == nil {
		return
	}
	b.FlowTo = id
}

// AccessibleLevel returns the heading level or tree depth.
func (b *Base) AccessibleLevel() int {
	if b == nil {
		return 0
	}
	return b.Level
}

// AccessibleOrientation returns the orientation (horizontal/vertical).
func (b *Base) AccessibleOrientation() string {
	if b == nil {
		return ""
	}
	return b.Orientation
}

// AccessibleActiveDescendant returns the ID of the active descendant.
func (b *Base) AccessibleActiveDescendant() string {
	if b == nil {
		return ""
	}
	return b.ActiveDescendant
}

// AccessiblePosInSet returns the 1-based position in the set.
func (b *Base) AccessiblePosInSet() int {
	if b == nil {
		return 0
	}
	return b.PosInSet
}

// AccessibleSetSize returns the total items in the set.
func (b *Base) AccessibleSetSize() int {
	if b == nil {
		return 0
	}
	return b.SetSize
}

// AccessibleHasPopup returns the popup type.
func (b *Base) AccessibleHasPopup() string {
	if b == nil {
		return ""
	}
	return b.HasPopup
}

// AccessibleErrorMessage returns the ID of the error message element.
func (b *Base) AccessibleErrorMessage() string {
	if b == nil {
		return ""
	}
	return b.ErrorMessage
}

// AccessibleCurrent returns the current indicator (page/step/location/date/time/true).
func (b *Base) AccessibleCurrent() string {
	if b == nil {
		return ""
	}
	return b.Current
}

// AccessibleAutocomplete returns the autocomplete mode.
func (b *Base) AccessibleAutocomplete() string {
	if b == nil {
		return ""
	}
	return b.Autocomplete
}

// AccessiblePlaceholder returns the placeholder text.
func (b *Base) AccessiblePlaceholder() string {
	if b == nil {
		return ""
	}
	return b.Placeholder
}

// AccessibleSort returns the sort direction.
func (b *Base) AccessibleSort() string {
	if b == nil {
		return ""
	}
	return b.Sort
}

// AccessibleKeyShortcuts returns the keyboard shortcuts.
func (b *Base) AccessibleKeyShortcuts() string {
	if b == nil {
		return ""
	}
	return b.KeyShortcuts
}

// AccessibleDetails returns the ID of the details element.
func (b *Base) AccessibleDetails() string {
	if b == nil {
		return ""
	}
	return b.Details
}

// AccessibleRoleDescription returns the custom role description.
func (b *Base) AccessibleRoleDescription() string {
	if b == nil {
		return ""
	}
	return b.RoleDescription
}

// SetLevel updates the heading level or tree depth.
func (b *Base) SetLevel(level int) {
	if b == nil {
		return
	}
	b.Level = level
}

// SetOrientation updates the orientation.
func (b *Base) SetOrientation(orientation string) {
	if b == nil {
		return
	}
	b.Orientation = orientation
}

// SetActiveDescendant updates the ID of the active descendant.
func (b *Base) SetActiveDescendant(id string) {
	if b == nil {
		return
	}
	b.ActiveDescendant = id
}

// SetPosInSet updates the 1-based position in the set.
func (b *Base) SetPosInSet(pos int) {
	if b == nil {
		return
	}
	b.PosInSet = pos
}

// SetSetSize updates the total items in the set.
func (b *Base) SetSetSize(size int) {
	if b == nil {
		return
	}
	b.SetSize = size
}

// SetHasPopup updates the popup type.
func (b *Base) SetHasPopup(popup string) {
	if b == nil {
		return
	}
	b.HasPopup = popup
}

// SetErrorMessage updates the ID of the error message element.
func (b *Base) SetErrorMessage(id string) {
	if b == nil {
		return
	}
	b.ErrorMessage = id
}

// SetCurrent updates the current indicator.
func (b *Base) SetCurrent(current string) {
	if b == nil {
		return
	}
	b.Current = current
}

// SetAutocomplete updates the autocomplete mode.
func (b *Base) SetAutocomplete(autocomplete string) {
	if b == nil {
		return
	}
	b.Autocomplete = autocomplete
}

// SetPlaceholder updates the placeholder text.
func (b *Base) SetPlaceholder(placeholder string) {
	if b == nil {
		return
	}
	b.Placeholder = placeholder
}

// SetSort updates the sort direction.
func (b *Base) SetSort(sort string) {
	if b == nil {
		return
	}
	b.Sort = sort
}

// SetKeyShortcuts updates the keyboard shortcuts.
func (b *Base) SetKeyShortcuts(shortcuts string) {
	if b == nil {
		return
	}
	b.KeyShortcuts = shortcuts
}

// SetDetails updates the ID of the details element.
func (b *Base) SetDetails(id string) {
	if b == nil {
		return
	}
	b.Details = id
}

// SetRoleDescription updates the custom role description.
func (b *Base) SetRoleDescription(desc string) {
	if b == nil {
		return
	}
	b.RoleDescription = desc
}

// HasExtendedARIA returns true if any WAI-ARIA 1.2/1.3 extended properties are set.
// This allows callers to skip serialization/processing of extended properties when none are in use.
func (b *Base) HasExtendedARIA() bool {
	if b == nil {
		return false
	}
	return b.Level != 0 || b.Orientation != "" ||
		b.ActiveDescendant != "" || b.PosInSet != 0 || b.SetSize != 0 ||
		b.HasPopup != "" || b.ErrorMessage != "" || b.Current != "" ||
		b.Autocomplete != "" || b.Placeholder != "" || b.Sort != "" ||
		b.KeyShortcuts != "" || b.Details != "" || b.RoleDescription != ""
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

const (
	// defaultMaxHistory limits SimpleAnnouncer history size to prevent unbounded growth.
	defaultMaxHistory = 100
)

// SimpleAnnouncer stores announcements in memory.
type SimpleAnnouncer struct {
	mu        sync.Mutex
	history   []Announcement
	onMessage func(Announcement)
	speaker   Speaker

	// speech state
	debounceTimer    *time.Timer
	speechCancel     context.CancelFunc
	speechAssertive  bool // true if current speech is assertive priority
}

// SetSpeaker attaches a TTS speaker to the announcer.
// When set, announcements are automatically spoken.
func (a *SimpleAnnouncer) SetSpeaker(s Speaker) {
	if a == nil {
		return
	}
	a.mu.Lock()
	a.speaker = s
	a.mu.Unlock()
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

// CloseSpeaker stops any pending speech and closes the speaker.
func (a *SimpleAnnouncer) CloseSpeaker() {
	if a == nil {
		return
	}
	a.mu.Lock()
	if a.debounceTimer != nil {
		a.debounceTimer.Stop()
		a.debounceTimer = nil
	}
	if a.speechCancel != nil {
		a.speechCancel()
	}
	s := a.speaker
	a.speaker = nil
	a.mu.Unlock()
	if s != nil {
		_ = s.Close()
	}
}

// addToHistory appends an announcement to the history, maintaining a bounded size.
// When the history exceeds defaultMaxHistory, the oldest entries are removed.
// Must be called with a.mu held.
func (a *SimpleAnnouncer) addToHistory(ann Announcement) {
	a.history = append(a.history, ann)
	if len(a.history) > defaultMaxHistory {
		// Shift to avoid holding references to old entries.
		copy(a.history, a.history[len(a.history)-defaultMaxHistory:])
		a.history = a.history[:defaultMaxHistory]
	}
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
	a.addToHistory(announcement)
	cb := a.onMessage
	speaker := a.speaker
	a.mu.Unlock()
	if cb != nil {
		cb(announcement)
	}
	if speaker != nil {
		a.dispatchSpeech(msg, priority, speaker)
	}
}

func (a *SimpleAnnouncer) dispatchSpeech(text string, priority Priority, speaker Speaker) {
	assertive := priority == PriorityAssertive || priority == PriorityUrgent || priority == PriorityHigh

	if assertive {
		// Interrupt any current speech and speak immediately.
		a.mu.Lock()
		if a.debounceTimer != nil {
			a.debounceTimer.Stop()
			a.debounceTimer = nil
		}
		if a.speechCancel != nil {
			a.speechCancel()
		}
		ctx, cancel := context.WithCancel(context.Background())
		a.speechCancel = cancel
		a.speechAssertive = true
		a.mu.Unlock()
		_ = speaker.Stop()
		go func() {
			_ = speaker.Speak(ctx, text)
			a.mu.Lock()
			a.speechAssertive = false
			a.mu.Unlock()
		}()
		return
	}

	// Polite: debounce 50ms, speak the latest message.
	// Never interrupts in-progress assertive speech — waits for it to finish.
	a.mu.Lock()
	if a.debounceTimer != nil {
		a.debounceTimer.Stop()
	}
	a.debounceTimer = time.AfterFunc(50*time.Millisecond, func() {
		a.mu.Lock()
		// Only cancel previous speech if it's not assertive.
		if !a.speechAssertive && a.speechCancel != nil {
			a.speechCancel()
		}
		ctx, cancel := context.WithCancel(context.Background())
		a.speechCancel = cancel
		a.mu.Unlock()
		// speaker.Speak blocks until any in-progress speech finishes (speaker mutex).
		_ = speaker.Speak(ctx, text)
	})
	a.mu.Unlock()
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
	if level := widget.AccessibleLevel(); level > 0 {
		parts = append(parts, fmt.Sprintf("level %d", level))
	}
	if orientation := widget.AccessibleOrientation(); orientation != "" {
		parts = append(parts, orientation)
	}
	if current := widget.AccessibleCurrent(); current != "" {
		parts = append(parts, "current "+current)
	}
	if pos, size := widget.AccessiblePosInSet(), widget.AccessibleSetSize(); pos > 0 && size > 0 {
		parts = append(parts, fmt.Sprintf("%d of %d", pos, size))
	}
	if errMsg := widget.AccessibleErrorMessage(); errMsg != "" {
		parts = append(parts, "error: "+errMsg)
	}
	if roleDesc := widget.AccessibleRoleDescription(); roleDesc != "" && role != "" {
		for i, p := range parts {
			if p == role {
				parts[i] = roleDesc
				break
			}
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
