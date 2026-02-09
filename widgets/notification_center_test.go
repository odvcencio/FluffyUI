package widgets

import (
	"strings"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	fluffytest "github.com/odvcencio/fluffyui/testing"
)

func TestNotificationCenter_NewNotificationCenter(t *testing.T) {
	nc := NewNotificationCenter()
	if nc == nil {
		t.Fatal("expected non-nil NotificationCenter")
	}
	if nc.Expanded() {
		t.Fatal("expected collapsed by default")
	}
	if nc.UnreadCount() != 0 {
		t.Fatalf("expected 0 unread, got %d", nc.UnreadCount())
	}
	if nc.AccessibleRole() != "log" {
		t.Fatalf("expected role log, got %q", nc.AccessibleRole())
	}
	if nc.AccessibleLive() != "polite" {
		t.Fatalf("expected live polite, got %q", nc.AccessibleLive())
	}
	if !nc.CanFocus() {
		t.Fatal("expected NotificationCenter to be focusable")
	}
}

func TestNotificationCenter_Add(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Add(Notification{
		ID:    "n1",
		Title: "Hello",
		Body:  "World",
		Level: NotificationInfo,
		Time:  time.Now(),
	})
	if len(nc.Notifications()) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(nc.Notifications()))
	}
	n := nc.Notifications()[0]
	if n.ID != "n1" {
		t.Fatalf("expected ID n1, got %q", n.ID)
	}
	if n.Title != "Hello" {
		t.Fatalf("expected title Hello, got %q", n.Title)
	}
	if n.Read {
		t.Fatal("expected notification to be unread")
	}
}

func TestNotificationCenter_AddSimple(t *testing.T) {
	nc := NewNotificationCenter()
	nc.AddSimple(NotificationSuccess, "Deployed", "Version 1.0")
	nc.AddSimple(NotificationWarning, "Low disk", "")

	notifs := nc.Notifications()
	if len(notifs) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(notifs))
	}
	if notifs[0].Level != NotificationSuccess {
		t.Fatalf("expected success level, got %v", notifs[0].Level)
	}
	if notifs[1].Level != NotificationWarning {
		t.Fatalf("expected warning level, got %v", notifs[1].Level)
	}
	if notifs[0].ID == "" {
		t.Fatal("expected auto-generated ID")
	}
}

func TestNotificationCenter_AddSetsTimeIfZero(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Add(Notification{ID: "t1", Title: "Test"})
	notifs := nc.Notifications()
	if notifs[0].Time.IsZero() {
		t.Fatal("expected Time to be populated when zero")
	}
}

func TestNotificationCenter_Dismiss(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Add(Notification{ID: "a", Title: "A"})
	nc.Add(Notification{ID: "b", Title: "B"})
	nc.Add(Notification{ID: "c", Title: "C"})

	var dismissed string
	nc.SetOnDismiss(func(id string) {
		dismissed = id
	})

	nc.Dismiss("b")
	notifs := nc.Notifications()
	if len(notifs) != 2 {
		t.Fatalf("expected 2 notifications after dismiss, got %d", len(notifs))
	}
	if dismissed != "b" {
		t.Fatalf("expected dismissed callback for 'b', got %q", dismissed)
	}
	for _, n := range notifs {
		if n.ID == "b" {
			t.Fatal("notification 'b' should have been removed")
		}
	}
}

func TestNotificationCenter_DismissAll(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Add(Notification{ID: "a", Title: "A"})
	nc.Add(Notification{ID: "b", Title: "B"})
	nc.DismissAll()

	if len(nc.Notifications()) != 0 {
		t.Fatalf("expected 0 notifications after DismissAll, got %d", len(nc.Notifications()))
	}
}

func TestNotificationCenter_MarkRead(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Add(Notification{ID: "a", Title: "A"})
	nc.Add(Notification{ID: "b", Title: "B"})

	if nc.UnreadCount() != 2 {
		t.Fatalf("expected 2 unread, got %d", nc.UnreadCount())
	}

	nc.MarkRead("a")
	if nc.UnreadCount() != 1 {
		t.Fatalf("expected 1 unread after marking 'a', got %d", nc.UnreadCount())
	}

	// Verify the correct one was marked.
	notifs := nc.Notifications()
	if !notifs[0].Read {
		t.Fatal("expected notification 'a' to be read")
	}
	if notifs[1].Read {
		t.Fatal("expected notification 'b' to still be unread")
	}
}

func TestNotificationCenter_MarkAllRead(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Add(Notification{ID: "a", Title: "A"})
	nc.Add(Notification{ID: "b", Title: "B"})
	nc.Add(Notification{ID: "c", Title: "C"})

	nc.MarkAllRead()
	if nc.UnreadCount() != 0 {
		t.Fatalf("expected 0 unread after MarkAllRead, got %d", nc.UnreadCount())
	}
}

func TestNotificationCenter_UnreadCount(t *testing.T) {
	nc := NewNotificationCenter()
	if nc.UnreadCount() != 0 {
		t.Fatal("expected 0 unread for empty center")
	}
	nc.Add(Notification{ID: "a", Title: "A"})
	nc.Add(Notification{ID: "b", Title: "B", Read: true})
	if nc.UnreadCount() != 1 {
		t.Fatalf("expected 1 unread (b is pre-read), got %d", nc.UnreadCount())
	}
}

func TestNotificationCenter_ExpandCollapse(t *testing.T) {
	nc := NewNotificationCenter()
	if nc.Expanded() {
		t.Fatal("expected collapsed initially")
	}
	nc.SetExpanded(true)
	if !nc.Expanded() {
		t.Fatal("expected expanded after SetExpanded(true)")
	}
	nc.SetExpanded(false)
	if nc.Expanded() {
		t.Fatal("expected collapsed after SetExpanded(false)")
	}
}

func TestNotificationCenter_MaxItemsCap(t *testing.T) {
	nc := NewNotificationCenter()
	nc.SetMaxItems(3)
	for i := 0; i < 5; i++ {
		nc.Add(Notification{
			ID:    strings.Repeat("n", i+1),
			Title: strings.Repeat("N", i+1),
		})
	}
	notifs := nc.Notifications()
	if len(notifs) != 3 {
		t.Fatalf("expected 3 notifications after cap, got %d", len(notifs))
	}
	// Oldest should have been dropped.
	if notifs[0].ID == "n" {
		t.Fatal("oldest notification should have been trimmed")
	}
}

func TestNotificationCenter_MaxItemsMinimum(t *testing.T) {
	nc := NewNotificationCenter()
	nc.SetMaxItems(0) // should be clamped to 1
	nc.Add(Notification{ID: "a", Title: "A"})
	nc.Add(Notification{ID: "b", Title: "B"})
	if len(nc.Notifications()) != 1 {
		t.Fatalf("expected 1 notification with maxItems=1, got %d", len(nc.Notifications()))
	}
}

func TestNotificationCenter_KeyboardNavigation(t *testing.T) {
	nc := NewNotificationCenter()
	nc.SetExpanded(true)
	nc.Focus()

	nc.Add(Notification{ID: "a", Title: "A"})
	nc.Add(Notification{ID: "b", Title: "B"})
	nc.Add(Notification{ID: "c", Title: "C"})

	if nc.Selected() != 0 {
		t.Fatalf("expected selected=0, got %d", nc.Selected())
	}

	// Down
	res := nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if !res.Handled {
		t.Fatal("expected Down to be handled")
	}
	if nc.Selected() != 1 {
		t.Fatalf("expected selected=1 after Down, got %d", nc.Selected())
	}

	// Down again
	nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if nc.Selected() != 2 {
		t.Fatalf("expected selected=2, got %d", nc.Selected())
	}

	// Down at end should clamp
	nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if nc.Selected() != 2 {
		t.Fatalf("expected selected=2 (clamped), got %d", nc.Selected())
	}

	// Up
	res = nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if !res.Handled {
		t.Fatal("expected Up to be handled")
	}
	if nc.Selected() != 1 {
		t.Fatalf("expected selected=1 after Up, got %d", nc.Selected())
	}

	// Up to 0
	nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if nc.Selected() != 0 {
		t.Fatalf("expected selected=0, got %d", nc.Selected())
	}

	// Up at top should clamp
	nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyUp})
	if nc.Selected() != 0 {
		t.Fatalf("expected selected=0 (clamped), got %d", nc.Selected())
	}
}

func TestNotificationCenter_KeyboardEnterToggleExpand(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Focus()

	// When collapsed, Enter expands.
	res := nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !res.Handled {
		t.Fatal("expected Enter to be handled")
	}
	if !nc.Expanded() {
		t.Fatal("expected expanded after Enter on collapsed")
	}
}

func TestNotificationCenter_KeyboardEnterActivatesAction(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Focus()
	nc.SetExpanded(true)

	fired := false
	nc.Add(Notification{
		ID:    "a",
		Title: "A",
		Actions: []NotificationAction{
			{Label: "Open", OnClick: func() { fired = true }},
		},
	})

	res := nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if !res.Handled {
		t.Fatal("expected Enter to be handled")
	}
	if !fired {
		t.Fatal("expected action OnClick to fire")
	}
}

func TestNotificationCenter_KeyboardDeleteDismiss(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Focus()
	nc.SetExpanded(true)
	nc.Add(Notification{ID: "a", Title: "A"})
	nc.Add(Notification{ID: "b", Title: "B"})

	// Select second notification.
	nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})

	// Delete it.
	res := nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDelete})
	if !res.Handled {
		t.Fatal("expected Delete to be handled")
	}
	notifs := nc.Notifications()
	if len(notifs) != 1 {
		t.Fatalf("expected 1 notification after delete, got %d", len(notifs))
	}
	if notifs[0].ID != "a" {
		t.Fatalf("expected remaining notification to be 'a', got %q", notifs[0].ID)
	}
}

func TestNotificationCenter_KeyboardBackspaceDismiss(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Focus()
	nc.SetExpanded(true)
	nc.Add(Notification{ID: "x", Title: "X"})

	res := nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyBackspace})
	if !res.Handled {
		t.Fatal("expected Backspace to be handled")
	}
	if len(nc.Notifications()) != 0 {
		t.Fatal("expected notification to be dismissed")
	}
}

func TestNotificationCenter_KeyboardEscapeCloses(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Focus()
	nc.SetExpanded(true)

	res := nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if !res.Handled {
		t.Fatal("expected Escape to be handled")
	}
	if nc.Expanded() {
		t.Fatal("expected panel to close on Escape")
	}
}

func TestNotificationCenter_KeyboardEscapeWhenCollapsed(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Focus()

	res := nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEscape})
	if !res.Handled {
		t.Fatal("expected Escape when collapsed to emit Cancel command")
	}
	if len(res.Commands) == 0 {
		t.Fatal("expected Cancel command from Escape when collapsed")
	}
}

func TestNotificationCenter_KeyboardMarkRead(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Focus()
	nc.SetExpanded(true)
	nc.Add(Notification{ID: "a", Title: "A"})
	nc.Add(Notification{ID: "b", Title: "B"})

	// Mark selected (index 0) as read.
	res := nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'r'})
	if !res.Handled {
		t.Fatal("expected 'r' to be handled")
	}
	notifs := nc.Notifications()
	if !notifs[0].Read {
		t.Fatal("expected first notification to be marked read")
	}
	if notifs[1].Read {
		t.Fatal("expected second notification to still be unread")
	}
}

func TestNotificationCenter_KeyboardMarkAllRead(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Focus()
	nc.SetExpanded(true)
	nc.Add(Notification{ID: "a", Title: "A"})
	nc.Add(Notification{ID: "b", Title: "B"})

	res := nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'R'})
	if !res.Handled {
		t.Fatal("expected 'R' to be handled")
	}
	if nc.UnreadCount() != 0 {
		t.Fatalf("expected 0 unread after 'R', got %d", nc.UnreadCount())
	}
}

func TestNotificationCenter_UnfocusedUnhandled(t *testing.T) {
	nc := NewNotificationCenter()
	// Not focused.
	res := nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyEnter})
	if res.Handled {
		t.Fatal("expected unhandled when not focused")
	}
}

func TestNotificationCenter_MouseUnhandled(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Focus()
	res := nc.HandleMessage(runtime.MouseMsg{X: 0, Y: 0})
	if res.Handled {
		t.Fatal("expected mouse messages to be unhandled")
	}
}

func TestNotificationCenter_RenderCollapsed(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Add(Notification{ID: "a", Title: "A"})
	nc.Add(Notification{ID: "b", Title: "B"})

	output := fluffytest.RenderToString(nc, 20, 1)
	if !strings.Contains(output, "[2]") {
		t.Fatalf("expected collapsed badge to show [2], got %q", output)
	}
}

func TestNotificationCenter_RenderCollapsedZero(t *testing.T) {
	nc := NewNotificationCenter()
	output := fluffytest.RenderToString(nc, 20, 1)
	if !strings.Contains(output, "[0]") {
		t.Fatalf("expected collapsed badge to show [0], got %q", output)
	}
}

func TestNotificationCenter_RenderExpanded(t *testing.T) {
	nc := NewNotificationCenter()
	nc.SetExpanded(true)
	nc.Add(Notification{
		ID:    "n1",
		Title: "Deploy success",
		Level: NotificationSuccess,
		Time:  time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC),
	})
	nc.Add(Notification{
		ID:    "n2",
		Title: "Disk warning",
		Level: NotificationWarning,
		Time:  time.Date(2025, 1, 15, 14, 31, 0, 0, time.UTC),
	})

	output := fluffytest.RenderToString(nc, 60, 5)
	if !strings.Contains(output, "Notifications (2)") {
		t.Fatalf("expected header 'Notifications (2)', got %q", output)
	}
	if !strings.Contains(output, "Deploy success") {
		t.Fatalf("expected 'Deploy success' in output, got %q", output)
	}
	if !strings.Contains(output, "[v]") {
		t.Fatalf("expected success icon [v] in output, got %q", output)
	}
	if !strings.Contains(output, "[!]") {
		t.Fatalf("expected warning icon [!] in output, got %q", output)
	}
}

func TestNotificationCenter_RenderLevelIcons(t *testing.T) {
	nc := NewNotificationCenter()
	nc.SetExpanded(true)

	nc.Add(Notification{ID: "i", Title: "Info", Level: NotificationInfo, Time: time.Now()})
	nc.Add(Notification{ID: "s", Title: "Success", Level: NotificationSuccess, Time: time.Now()})
	nc.Add(Notification{ID: "w", Title: "Warning", Level: NotificationWarning, Time: time.Now()})
	nc.Add(Notification{ID: "e", Title: "Error", Level: NotificationError, Time: time.Now()})

	output := fluffytest.RenderToString(nc, 60, 6)
	for _, icon := range []string{"[i]", "[v]", "[!]", "[x]"} {
		if !strings.Contains(output, icon) {
			t.Fatalf("expected icon %q in output, got %q", icon, output)
		}
	}
}

func TestNotificationCenter_RenderUnreadMarker(t *testing.T) {
	nc := NewNotificationCenter()
	nc.SetExpanded(true)
	nc.Add(Notification{ID: "a", Title: "Unread", Time: time.Now()})
	nc.Add(Notification{ID: "b", Title: "Read", Time: time.Now(), Read: true})

	output := fluffytest.RenderToString(nc, 60, 4)
	// Unread notifications have a '*' marker.
	if !strings.Contains(output, "*") {
		t.Fatalf("expected '*' marker for unread notification, got %q", output)
	}
}

func TestNotificationCenter_A11yRole(t *testing.T) {
	nc := NewNotificationCenter()
	if nc.AccessibleRole() != "log" {
		t.Fatalf("expected role log, got %q", nc.AccessibleRole())
	}
}

func TestNotificationCenter_A11yDescription(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Add(Notification{ID: "a", Title: "A"})
	nc.Add(Notification{ID: "b", Title: "B", Read: true})

	desc := nc.AccessibleDescription()
	if !strings.Contains(desc, "2 notifications") {
		t.Fatalf("expected description to contain '2 notifications', got %q", desc)
	}
	if !strings.Contains(desc, "1 unread") {
		t.Fatalf("expected description to contain '1 unread', got %q", desc)
	}
}

func TestNotificationCenter_A11yValue(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Add(Notification{ID: "a", Title: "First", Body: "Details"})
	nc.Add(Notification{ID: "b", Title: "Second"})

	val := nc.AccessibleValue()
	if val == nil {
		t.Fatal("expected non-nil value")
	}
	if !strings.Contains(val.Text, "First") {
		t.Fatalf("expected value to contain selected notification title 'First', got %q", val.Text)
	}
	if !strings.Contains(val.Text, "Details") {
		t.Fatalf("expected value to contain body 'Details', got %q", val.Text)
	}
}

func TestNotificationCenter_DismissAdjustsSelected(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Focus()
	nc.SetExpanded(true)
	nc.Add(Notification{ID: "a", Title: "A"})
	nc.Add(Notification{ID: "b", Title: "B"})

	// Select second (index 1).
	nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if nc.Selected() != 1 {
		t.Fatalf("expected selected=1, got %d", nc.Selected())
	}

	// Dismiss last, selected should adjust.
	nc.Dismiss("b")
	if nc.Selected() != 0 {
		t.Fatalf("expected selected=0 after dismissing last, got %d", nc.Selected())
	}
}

func TestNotificationCenter_NilSafety(t *testing.T) {
	var nc *NotificationCenter

	// These should not panic.
	nc.Add(Notification{ID: "a", Title: "A"})
	nc.AddSimple(NotificationInfo, "title", "body")
	nc.Dismiss("a")
	nc.DismissAll()
	nc.MarkRead("a")
	nc.MarkAllRead()
	nc.SetExpanded(true)
	nc.SetMaxItems(10)
	nc.SetOnDismiss(func(string) {})
	nc.SetLabel("test")
	nc.SetStyle(backend.DefaultStyle())
	nc.SetSelectedStyle(backend.DefaultStyle())
	nc.SetHeaderStyle(backend.DefaultStyle())
	nc.Bind(runtime.Services{})
	nc.Unbind()

	if nc.UnreadCount() != 0 {
		t.Fatal("expected 0 from nil UnreadCount")
	}
	if nc.Expanded() {
		t.Fatal("expected false from nil Expanded")
	}
	if nc.Notifications() != nil {
		t.Fatal("expected nil from nil Notifications")
	}
	if nc.Selected() != 0 {
		t.Fatal("expected 0 from nil Selected")
	}
}

func TestNotificationCenter_NotificationLevelString(t *testing.T) {
	tests := []struct {
		level NotificationLevel
		want  string
	}{
		{NotificationInfo, "info"},
		{NotificationSuccess, "success"},
		{NotificationWarning, "warning"},
		{NotificationError, "error"},
		{NotificationLevel(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("NotificationLevel(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestNotificationCenter_SelectedNotUpdatedWhenCollapsed(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Focus()
	// Collapsed, navigation should not be handled.
	nc.Add(Notification{ID: "a", Title: "A"})
	nc.Add(Notification{ID: "b", Title: "B"})

	res := nc.HandleMessage(runtime.KeyMsg{Key: terminal.KeyDown})
	if res.Handled {
		t.Fatal("expected Down to be unhandled when collapsed")
	}
	if nc.Selected() != 0 {
		t.Fatal("expected selected to remain 0 when collapsed")
	}
}

func TestNotificationCenter_DismissNonExistent(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Add(Notification{ID: "a", Title: "A"})

	// Should not panic or change anything.
	nc.Dismiss("nonexistent")
	if len(nc.Notifications()) != 1 {
		t.Fatal("expected notification count unchanged")
	}
}

func TestNotificationCenter_MarkReadNonExistent(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Add(Notification{ID: "a", Title: "A"})

	// Should not panic.
	nc.MarkRead("nonexistent")
	if nc.UnreadCount() != 1 {
		t.Fatal("expected unread count unchanged")
	}
}

func TestNotificationCenter_StyleType(t *testing.T) {
	nc := NewNotificationCenter()
	if nc.StyleType() != "NotificationCenter" {
		t.Fatalf("expected StyleType NotificationCenter, got %q", nc.StyleType())
	}
}

func TestNotificationCenter_BindUnbind(t *testing.T) {
	nc := NewNotificationCenter()
	nc.Bind(runtime.Services{})
	nc.Unbind()
	// Should not panic.
}
