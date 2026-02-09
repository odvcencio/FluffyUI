package widgets

import (
	"fmt"
	"strings"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// NotificationLevel indicates notification severity.
type NotificationLevel int

const (
	// NotificationInfo is for informational notifications.
	NotificationInfo NotificationLevel = iota
	// NotificationSuccess is for success notifications.
	NotificationSuccess
	// NotificationWarning is for warning notifications.
	NotificationWarning
	// NotificationError is for error notifications.
	NotificationError
)

// String returns a string representation of the notification level.
func (l NotificationLevel) String() string {
	switch l {
	case NotificationInfo:
		return "info"
	case NotificationSuccess:
		return "success"
	case NotificationWarning:
		return "warning"
	case NotificationError:
		return "error"
	default:
		return "unknown"
	}
}

// notificationLevelIcon returns a text indicator for the level.
func notificationLevelIcon(level NotificationLevel) string {
	switch level {
	case NotificationSuccess:
		return "[v]"
	case NotificationWarning:
		return "[!]"
	case NotificationError:
		return "[x]"
	default:
		return "[i]"
	}
}

// Notification represents a persistent notification entry.
type Notification struct {
	ID      string
	Title   string
	Body    string
	Level   NotificationLevel
	Time    time.Time
	Read    bool
	Actions []NotificationAction
}

// NotificationAction represents an optional action button on a notification.
type NotificationAction struct {
	Label   string
	OnClick func()
}

const (
	notificationDefaultMax = 50
)

// NotificationCenter is a persistent, dismissible notification panel.
// When collapsed it shows an unread count badge; when expanded it shows
// a navigable list of notifications.
type NotificationCenter struct {
	FocusableBase

	notifications []Notification
	maxItems      int
	selected      int
	expanded      bool
	services      runtime.Services
	onDismiss     func(id string)
	label         string

	style         backend.Style
	selectedStyle backend.Style
	headerStyle   backend.Style
}

// NewNotificationCenter creates a new notification center widget.
func NewNotificationCenter() *NotificationCenter {
	nc := &NotificationCenter{
		maxItems:      notificationDefaultMax,
		label:         "Notification Center",
		style:         backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
		headerStyle:   backend.DefaultStyle().Bold(true),
	}
	nc.Base.Role = accessibility.RoleLog
	nc.Base.Live = accessibility.LivePolite
	nc.Base.Relevant = accessibility.RelevantAdditions
	nc.syncA11y()
	return nc
}

// Bind attaches app services.
func (nc *NotificationCenter) Bind(services runtime.Services) {
	if nc == nil {
		return
	}
	nc.services = services
}

// Unbind releases app services.
func (nc *NotificationCenter) Unbind() {
	if nc == nil {
		return
	}
	nc.services = runtime.Services{}
}

// StyleType returns the selector type name.
func (nc *NotificationCenter) StyleType() string {
	return "NotificationCenter"
}

// Add adds a notification. It announces new notifications via the a11y announcer.
func (nc *NotificationCenter) Add(n Notification) {
	if nc == nil {
		return
	}
	if n.Time.IsZero() {
		n.Time = time.Now()
	}
	nc.notifications = append(nc.notifications, n)
	nc.enforceMax()
	nc.syncA11y()

	// Announce the new notification.
	text := strings.TrimSpace(n.Title)
	if body := strings.TrimSpace(n.Body); body != "" {
		text = text + ": " + body
	}
	if text != "" {
		if announcer := nc.services.Announcer(); announcer != nil {
			priority := accessibility.PriorityPolite
			if n.Level == NotificationError {
				priority = accessibility.PriorityAssertive
			}
			announcer.Announce(text, priority)
		}
	}
}

// AddSimple is a convenience helper that creates a notification with an auto-generated ID.
func (nc *NotificationCenter) AddSimple(level NotificationLevel, title, body string) {
	if nc == nil {
		return
	}
	id := fmt.Sprintf("notif-%d", len(nc.notifications)+1)
	nc.Add(Notification{
		ID:    id,
		Title: title,
		Body:  body,
		Level: level,
		Time:  time.Now(),
	})
}

// Dismiss removes a notification by ID.
func (nc *NotificationCenter) Dismiss(id string) {
	if nc == nil {
		return
	}
	for i, n := range nc.notifications {
		if n.ID == id {
			nc.notifications = append(nc.notifications[:i], nc.notifications[i+1:]...)
			if nc.selected >= len(nc.notifications) && nc.selected > 0 {
				nc.selected = len(nc.notifications) - 1
			}
			if nc.onDismiss != nil {
				nc.onDismiss(id)
			}
			nc.syncA11y()
			return
		}
	}
}

// DismissAll clears all notifications.
func (nc *NotificationCenter) DismissAll() {
	if nc == nil {
		return
	}
	nc.notifications = nil
	nc.selected = 0
	nc.syncA11y()
}

// MarkRead marks a notification as read by ID.
func (nc *NotificationCenter) MarkRead(id string) {
	if nc == nil {
		return
	}
	for i := range nc.notifications {
		if nc.notifications[i].ID == id {
			nc.notifications[i].Read = true
			nc.syncA11y()
			return
		}
	}
}

// MarkAllRead marks all notifications as read.
func (nc *NotificationCenter) MarkAllRead() {
	if nc == nil {
		return
	}
	for i := range nc.notifications {
		nc.notifications[i].Read = true
	}
	nc.syncA11y()
}

// UnreadCount returns the number of unread notifications.
func (nc *NotificationCenter) UnreadCount() int {
	if nc == nil {
		return 0
	}
	count := 0
	for _, n := range nc.notifications {
		if !n.Read {
			count++
		}
	}
	return count
}

// SetExpanded opens or closes the notification panel.
func (nc *NotificationCenter) SetExpanded(expanded bool) {
	if nc == nil {
		return
	}
	nc.expanded = expanded
	nc.syncA11y()
}

// Expanded returns whether the panel is currently open.
func (nc *NotificationCenter) Expanded() bool {
	if nc == nil {
		return false
	}
	return nc.expanded
}

// SetMaxItems sets the maximum number of notifications to retain.
// Older notifications are dropped when the cap is exceeded.
func (nc *NotificationCenter) SetMaxItems(max int) {
	if nc == nil {
		return
	}
	if max < 1 {
		max = 1
	}
	nc.maxItems = max
	nc.enforceMax()
}

// SetOnDismiss registers a callback invoked when a notification is dismissed.
func (nc *NotificationCenter) SetOnDismiss(fn func(id string)) {
	if nc == nil {
		return
	}
	nc.onDismiss = fn
}

// SetLabel sets the accessibility label.
func (nc *NotificationCenter) SetLabel(label string) {
	if nc == nil {
		return
	}
	nc.label = label
	nc.syncA11y()
}

// SetStyle sets the base rendering style.
func (nc *NotificationCenter) SetStyle(style backend.Style) {
	if nc == nil {
		return
	}
	nc.style = style
}

// SetSelectedStyle sets the style for the selected notification row.
func (nc *NotificationCenter) SetSelectedStyle(style backend.Style) {
	if nc == nil {
		return
	}
	nc.selectedStyle = style
}

// SetHeaderStyle sets the style for the header line.
func (nc *NotificationCenter) SetHeaderStyle(style backend.Style) {
	if nc == nil {
		return
	}
	nc.headerStyle = style
}

// Notifications returns a copy of all notifications.
func (nc *NotificationCenter) Notifications() []Notification {
	if nc == nil {
		return nil
	}
	out := make([]Notification, len(nc.notifications))
	copy(out, nc.notifications)
	return out
}

// Selected returns the currently selected index.
func (nc *NotificationCenter) Selected() int {
	if nc == nil {
		return 0
	}
	return nc.selected
}

// Measure returns the desired size.
func (nc *NotificationCenter) Measure(constraints runtime.Constraints) runtime.Size {
	return nc.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		if !nc.expanded {
			// Collapsed: badge with unread count, e.g. "[3]"
			badge := nc.badgeText()
			width := textWidth(badge)
			if width < 1 {
				width = 1
			}
			return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
		}
		// Expanded: header + notification list.
		height := 1 + len(nc.notifications) // header + one line per notification
		if height < 2 {
			height = 2
		}
		return contentConstraints.Constrain(runtime.Size{
			Width:  contentConstraints.MaxWidth,
			Height: height,
		})
	})
}

// Render draws the notification center.
func (nc *NotificationCenter) Render(ctx runtime.RenderContext) {
	if nc == nil {
		return
	}
	nc.syncA11y()
	outer := nc.bounds
	content := nc.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}

	baseStyle := resolveBaseStyle(ctx, nc, nc.style, false)

	if !nc.expanded {
		// Collapsed: render badge.
		badge := nc.badgeText()
		ctx.Buffer.Fill(content, ' ', baseStyle)
		writePadded(ctx.Buffer, content.X, content.Y, content.Width, truncateString(badge, content.Width), baseStyle)
		return
	}

	// Expanded: fill background.
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}

	// Header line.
	header := fmt.Sprintf("Notifications (%d)", len(nc.notifications))
	headerStyle := mergeBackendStyles(baseStyle, nc.headerStyle)
	writePadded(ctx.Buffer, content.X, content.Y, content.Width, truncateString(header, content.Width), headerStyle)

	// Notification list.
	listY := content.Y + 1
	for i, n := range nc.notifications {
		row := listY + i
		if row >= content.Y+content.Height {
			break
		}

		rowStyle := baseStyle
		if i == nc.selected {
			rowStyle = mergeBackendStyles(baseStyle, nc.selectedStyle)
		}

		line := nc.formatNotificationLine(n, content.Width)
		writePadded(ctx.Buffer, content.X, row, content.Width, line, rowStyle)
	}
}

// HandleMessage handles keyboard input for navigation and actions.
func (nc *NotificationCenter) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if nc == nil || !nc.focused {
		return runtime.Unhandled()
	}

	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyEnter:
		if !nc.expanded {
			nc.expanded = true
			nc.syncA11y()
			return runtime.Handled()
		}
		// Activate action on selected notification.
		if nc.selected >= 0 && nc.selected < len(nc.notifications) {
			n := nc.notifications[nc.selected]
			if len(n.Actions) > 0 && n.Actions[0].OnClick != nil {
				n.Actions[0].OnClick()
				return runtime.Handled()
			}
		}
		return runtime.Handled()

	case terminal.KeyUp:
		if nc.expanded && len(nc.notifications) > 0 {
			nc.selected--
			if nc.selected < 0 {
				nc.selected = 0
			}
			return runtime.Handled()
		}

	case terminal.KeyDown:
		if nc.expanded && len(nc.notifications) > 0 {
			nc.selected++
			if nc.selected >= len(nc.notifications) {
				nc.selected = len(nc.notifications) - 1
			}
			return runtime.Handled()
		}

	case terminal.KeyDelete, terminal.KeyBackspace:
		if nc.expanded && nc.selected >= 0 && nc.selected < len(nc.notifications) {
			id := nc.notifications[nc.selected].ID
			nc.Dismiss(id)
			return runtime.Handled()
		}

	case terminal.KeyEscape:
		if nc.expanded {
			nc.expanded = false
			nc.syncA11y()
			return runtime.Handled()
		}
		return runtime.WithCommand(runtime.Cancel{})

	case terminal.KeyRune:
		switch key.Rune {
		case 'r':
			if nc.expanded && nc.selected >= 0 && nc.selected < len(nc.notifications) {
				nc.notifications[nc.selected].Read = true
				nc.syncA11y()
				return runtime.Handled()
			}
		case 'R':
			if nc.expanded {
				nc.MarkAllRead()
				return runtime.Handled()
			}
		}
	}

	return runtime.Unhandled()
}

// enforceMax trims oldest notifications beyond maxItems.
func (nc *NotificationCenter) enforceMax() {
	if len(nc.notifications) > nc.maxItems {
		excess := len(nc.notifications) - nc.maxItems
		nc.notifications = nc.notifications[excess:]
		if nc.selected >= len(nc.notifications) {
			nc.selected = len(nc.notifications) - 1
		}
		if nc.selected < 0 {
			nc.selected = 0
		}
	}
}

// badgeText returns the collapsed badge text showing unread count.
func (nc *NotificationCenter) badgeText() string {
	unread := nc.UnreadCount()
	if unread == 0 {
		return "[0]"
	}
	return fmt.Sprintf("[%d]", unread)
}

// formatNotificationLine formats one notification for the expanded list.
func (nc *NotificationCenter) formatNotificationLine(n Notification, maxWidth int) string {
	icon := notificationLevelIcon(n.Level)
	readMark := " "
	if !n.Read {
		readMark = "*"
	}
	timeStr := n.Time.Format("15:04")
	prefix := icon + readMark + timeStr + " "
	remaining := maxWidth - textWidth(prefix)
	if remaining <= 0 {
		return truncateString(prefix, maxWidth)
	}
	title := truncateString(n.Title, remaining)
	return prefix + title
}

// syncA11y updates accessibility attributes.
func (nc *NotificationCenter) syncA11y() {
	if nc == nil {
		return
	}
	if nc.Base.Role == "" {
		nc.Base.Role = accessibility.RoleLog
	}
	label := strings.TrimSpace(nc.label)
	if label == "" {
		label = "Notification Center"
	}
	nc.Base.Label = label

	total := len(nc.notifications)
	unread := nc.UnreadCount()
	nc.Base.Description = fmt.Sprintf("%d notifications, %d unread", total, unread)

	if total > 0 && nc.selected >= 0 && nc.selected < total {
		sel := nc.notifications[nc.selected]
		text := strings.TrimSpace(sel.Title)
		if body := strings.TrimSpace(sel.Body); body != "" {
			text = text + ": " + body
		}
		nc.Base.Value = &accessibility.ValueInfo{Text: text}
	} else {
		nc.Base.Value = nil
	}
}

var _ runtime.Widget = (*NotificationCenter)(nil)
var _ runtime.Focusable = (*NotificationCenter)(nil)
var _ runtime.Bindable = (*NotificationCenter)(nil)
var _ runtime.Unbindable = (*NotificationCenter)(nil)
