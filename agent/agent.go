// Package agent provides AI-friendly interaction with FluffyUI applications.
// It enables automated testing, AI agents, and scripted interactions by exposing
// a semantic API over the widget tree rather than raw terminal I/O.
package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// Common errors returned by Agent methods.
var (
	ErrWidgetNotFound = errors.New("widget not found")
	ErrWidgetDisabled = errors.New("widget is disabled")
	ErrNotFocusable   = errors.New("widget is not focusable")
	ErrNotInteractive = errors.New("widget is not interactive")
	ErrTimeout        = errors.New("operation timed out")
	ErrNoApp          = errors.New("no app configured")
)

// Agent provides AI-friendly interaction with a FluffyUI application.
// It wraps a simulation backend and exposes semantic operations over
// the widget tree.
type Agent struct {
	mu          sync.Mutex
	app         *runtime.App
	sim         *sim.Backend
	postKey     PostKeyFunc
	screen      *runtime.Screen
	tickRate    time.Duration
	autoAttach  bool
	includeText bool
}

// PostKeyFunc is a callback for injecting key events into a custom event loop.
// Used when the application has its own message loop instead of runtime.App.
type PostKeyFunc func(msg runtime.KeyMsg) error

// Config configures an Agent.
type Config struct {
	// App is the FluffyUI application to control.
	// When provided, the agent auto-attaches to the app's screen once available.
	App *runtime.App

	// Sim is the simulation backend. If nil and App is not set, one will be created.
	Sim *sim.Backend

	// PostKey is an optional callback for posting key events.
	// When set, key events are sent through this function instead of the sim backend.
	// This is useful for applications with custom event loops.
	PostKey PostKeyFunc

	// DisableAutoAttach skips automatic App.Screen() attachment.
	// Useful when the caller wants to manage SetScreen explicitly.
	DisableAutoAttach bool

	// Width and Height set the terminal dimensions (default 80x24).
	Width, Height int

	// TickRate is how long to wait between operations for UI to settle.
	// Default is 50ms.
	TickRate time.Duration

	// IncludeText controls whether snapshots include raw screen text.
	// Default is false.
	IncludeText bool
}

// New creates a new Agent with the given configuration.
func New(cfg Config) *Agent {
	width, height := cfg.Width, cfg.Height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	s := cfg.Sim
	if s == nil && cfg.App == nil {
		s = sim.New(width, height)
	}

	tickRate := cfg.TickRate
	if tickRate <= 0 {
		tickRate = 50 * time.Millisecond
	}

	autoAttach := !cfg.DisableAutoAttach
	includeText := cfg.IncludeText

	var screen *runtime.Screen
	if cfg.App != nil && autoAttach {
		screen = cfg.App.Screen()
	}

	return &Agent{
		app:         cfg.App,
		sim:         s,
		postKey:     cfg.PostKey,
		screen:      screen,
		tickRate:    tickRate,
		autoAttach:  autoAttach,
		includeText: includeText,
	}
}

// Backend returns the underlying simulation backend.
func (a *Agent) Backend() *sim.Backend {
	if a == nil {
		return nil
	}
	return a.sim
}

// SetScreen overrides the screen reference for widget tree access.
// Most callers can rely on the agent auto-attaching to app.Screen().
func (a *Agent) SetScreen(screen *runtime.Screen) {
	if a == nil {
		return
	}
	a.mu.Lock()
	a.screen = screen
	a.mu.Unlock()
}

// Screen returns the current screen.
func (a *Agent) Screen() *runtime.Screen {
	if a == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.ensureScreenLocked()
}

func (a *Agent) ensureScreenLocked() *runtime.Screen {
	if !a.autoAttach {
		return a.screen
	}
	if a.screen == nil && a.app != nil {
		if screen := a.app.Screen(); screen != nil {
			a.screen = screen
		}
	}
	return a.screen
}

// Tick waits for the UI to process pending events.
func (a *Agent) Tick() {
	if a == nil {
		return
	}
	time.Sleep(a.tickRate)
}

// Snapshot returns a structured representation of the current UI state.
func (a *Agent) Snapshot() Snapshot {
	snap, _ := a.SnapshotWithContext(context.Background(), SnapshotOptions{
		IncludeText: a.includeText,
	})
	return snap
}

// SnapshotWithContext captures a snapshot on the app's event loop when available.
func (a *Agent) SnapshotWithContext(ctx context.Context, opts SnapshotOptions) (Snapshot, error) {
	if a == nil {
		return Snapshot{}, ErrNoApp
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if a.app != nil {
		var snap Snapshot
		err := a.app.Call(ctx, func(app *runtime.App) error {
			a.mu.Lock()
			defer a.mu.Unlock()
			if a.autoAttach && a.screen == nil {
				a.screen = app.Screen()
			}
			snap = a.snapshotLocked(opts.IncludeText)
			return nil
		})
		if err != nil {
			return Snapshot{}, err
		}
		return snap, nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	return a.snapshotLocked(opts.IncludeText), nil
}

func (a *Agent) snapshotLocked(includeText bool) Snapshot {
	snap := Snapshot{
		Timestamp: time.Now(),
	}

	a.ensureScreenLocked()
	if includeText {
		snap.Text = a.captureTextLocked()
	}

	if a.sim != nil {
		snap.Width, _ = a.sim.Size()
		_, snap.Height = a.sim.Size()
	}

	if a.screen == nil {
		return snap
	}

	snap.Width, snap.Height = a.screen.Size()
	snap.LayerCount = a.screen.LayerCount()

	// Walk all layers from base to top for complete widget tree
	for i := 0; i < a.screen.LayerCount(); i++ {
		layer := a.screen.Layer(i)
		if layer != nil && layer.Root != nil {
			explicitCounts := make(map[string]int)
			info := a.walkWidgets(layer.Root, i, []int{0}, explicitCounts)
			snap.Widgets = append(snap.Widgets, info)
		}
	}

	// Find focused widget, preferring the topmost layer.
	for i := len(snap.Widgets) - 1; i >= 0; i-- {
		if focused := findFocusedInfo(&snap.Widgets[i]); focused != nil {
			snap.FocusedID = focused.ID
			snap.Focused = focused
			break
		}
	}

	return snap
}

// walkWidgets recursively collects widget info from the tree.
func (a *Agent) walkWidgets(w runtime.Widget, layer int, path []int, explicitCounts map[string]int) WidgetInfo {
	if w == nil {
		return WidgetInfo{}
	}

	info := a.extractWidgetInfo(w, buildWidgetID(w, layer, path, explicitCounts, true))

	// Check for children
	if cp, ok := w.(runtime.ChildProvider); ok {
		children := cp.ChildWidgets()
		for i, child := range children {
			if child == nil {
				continue
			}
			childPath := make([]int, len(path)+1)
			copy(childPath, path)
			childPath[len(path)] = i
			info.Children = append(info.Children, a.walkWidgets(child, layer, childPath, explicitCounts))
		}
	}

	return info
}

// extractWidgetInfo builds WidgetInfo from a widget.
func (a *Agent) extractWidgetInfo(w runtime.Widget, id string) WidgetInfo {
	info := WidgetInfo{
		ID: id,
	}

	// Get bounds
	if bp, ok := w.(runtime.BoundsProvider); ok {
		info.Bounds = bp.Bounds()
	}

	// Get accessibility info
	if acc, ok := w.(accessibility.Accessible); ok {
		info.Role = acc.AccessibleRole()
		info.Label = acc.AccessibleLabel()
		info.Description = acc.AccessibleDescription()
		info.State = acc.AccessibleState()
		if val := acc.AccessibleValue(); val != nil {
			info.Value = val.Text
			info.ValueInfo = val
		}
	}

	// Infer textbox role/value from focusable text widgets when accessibility is missing.
	if info.Role == "" {
		if f, ok := w.(runtime.Focusable); ok && f.CanFocus() {
			if textWidget, ok := w.(interface{ Text() string }); ok {
				info.Role = accessibility.RoleTextbox
				info.Value = textWidget.Text()
			}
		}
	}

	if info.Role == "" {
		if textWidget, ok := w.(interface{ Text() string }); ok {
			if text := strings.TrimSpace(textWidget.Text()); text != "" {
				info.Role = accessibility.RoleText
				if info.Label == "" {
					info.Label = text
				}
			}
		}
	}

	if info.Role == "" {
		info.Role = accessibility.RoleGroup
	}

	if info.Label == "" {
		info.Label = defaultWidgetLabel(w)
	}

	if info.Value == "" {
		if textWidget, ok := w.(interface{ Text() string }); ok {
			info.Value = textWidget.Text()
		}
	}

	// Check focusable
	if f, ok := w.(runtime.Focusable); ok {
		info.Focusable = f.CanFocus()
		info.Focused = f.IsFocused()
	}

	// Determine available actions based on role
	info.Actions = actionsForRole(info.Role, info.State)
	if !info.Focusable {
		info.Actions = removeAction(info.Actions, "focus")
	}

	return info
}

func defaultWidgetLabel(w runtime.Widget) string {
	if w == nil {
		return "Widget"
	}
	typ := reflect.TypeOf(w)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	name := strings.TrimSpace(typ.Name())
	if name == "" {
		return "Widget"
	}
	return name
}

func removeAction(actions []string, action string) []string {
	if len(actions) == 0 {
		return actions
	}
	trimmed := strings.TrimSpace(action)
	if trimmed == "" {
		return actions
	}
	out := actions[:0]
	for _, entry := range actions {
		if entry == trimmed {
			continue
		}
		out = append(out, entry)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// buildWidgetID generates a deterministic identifier for a widget.
func buildWidgetID(w runtime.Widget, layer int, path []int, explicitCounts map[string]int, logCollisions bool) string {
	if w == nil {
		return ""
	}
	widgetType := widgetTypeName(w)
	explicit := widgetExplicitID(w)
	pathID := formatWidgetPath(path)
	if explicit != "" {
		pathID = "*"
		count := explicitCounts[explicit] + 1
		explicitCounts[explicit] = count
		suffix := ""
		if count > 1 {
			suffix = fmt.Sprintf("#%d", count)
			if logCollisions {
				log.Printf("agent: widget id collision %q on layer %d", explicit, layer)
			}
		}
		return fmt.Sprintf("layer%d:%s:%s:%s%s", layer, widgetType, pathID, explicit, suffix)
	}
	return fmt.Sprintf("layer%d:%s:%s", layer, widgetType, pathID)
}

// actionsForRole returns available actions based on widget role and state.
func actionsForRole(role accessibility.Role, state accessibility.StateSet) []string {
	if state.Disabled {
		return nil
	}

	switch role {
	case accessibility.RoleButton:
		return []string{"activate", "focus"}
	case accessibility.RoleCheckbox, accessibility.RoleRadio:
		return []string{"toggle", "focus"}
	case accessibility.RoleTextbox:
		return []string{"type", "clear", "focus"}
	case accessibility.RoleList, accessibility.RoleTree:
		return []string{"select", "focus", "scroll"}
	case accessibility.RoleMenuItem:
		return []string{"activate"}
	case accessibility.RoleTab:
		return []string{"activate", "focus"}
	default:
		return []string{"focus"}
	}
}

// FindByLabel finds the first widget with a matching label (case-insensitive substring).
func (a *Agent) FindByLabel(label string) *WidgetInfo {
	snap := a.Snapshot()
	return findByLabelIn(snap.Widgets, label)
}

func findByLabelIn(widgets []WidgetInfo, label string) *WidgetInfo {
	label = strings.ToLower(label)
	for i := range widgets {
		w := &widgets[i]
		if strings.Contains(strings.ToLower(w.Label), label) {
			return w
		}
		if found := findByLabelIn(w.Children, label); found != nil {
			return found
		}
	}
	return nil
}

func widgetExplicitID(w runtime.Widget) string {
	if w == nil {
		return ""
	}
	type idGetter interface {
		ID() string
	}
	if getter, ok := w.(idGetter); ok {
		return strings.TrimSpace(getter.ID())
	}
	return ""
}

func widgetTypeName(w runtime.Widget) string {
	if w == nil {
		return "widget"
	}
	typ := reflect.TypeOf(w)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	name := strings.TrimSpace(typ.Name())
	if name == "" {
		return "widget"
	}
	return strings.ToLower(name)
}

func formatWidgetPath(path []int) string {
	if len(path) == 0 {
		return "0"
	}
	var b strings.Builder
	for i, entry := range path {
		if i > 0 {
			b.WriteByte('.')
		}
		b.WriteString(strconv.Itoa(entry))
	}
	return b.String()
}

func findFocusedInfo(info *WidgetInfo) *WidgetInfo {
	if info == nil {
		return nil
	}
	if info.Focused {
		return info
	}
	for i := range info.Children {
		if found := findFocusedInfo(&info.Children[i]); found != nil {
			return found
		}
	}
	return nil
}

// FindByRole finds all widgets with the given role.
func (a *Agent) FindByRole(role accessibility.Role) []WidgetInfo {
	snap := a.Snapshot()
	var results []WidgetInfo
	findByRoleIn(snap.Widgets, role, &results)
	return results
}

// FindByType is an alias for FindByRole.
func (a *Agent) FindByType(role accessibility.Role) []WidgetInfo {
	return a.FindByRole(role)
}

func findByRoleIn(widgets []WidgetInfo, role accessibility.Role, out *[]WidgetInfo) {
	for _, w := range widgets {
		if w.Role == role {
			*out = append(*out, w)
		}
		findByRoleIn(w.Children, role, out)
	}
}

// FindByID finds a widget by its ID.
func (a *Agent) FindByID(id string) *WidgetInfo {
	snap := a.Snapshot()
	return findByIDIn(snap.Widgets, id)
}

func findByIDIn(widgets []WidgetInfo, id string) *WidgetInfo {
	for i := range widgets {
		w := &widgets[i]
		if w.ID == id {
			return w
		}
		if found := findByIDIn(w.Children, id); found != nil {
			return found
		}
	}
	return nil
}

// GetFocused returns the currently focused widget.
func (a *Agent) GetFocused() *WidgetInfo {
	snap := a.Snapshot()
	return snap.Focused
}

// IsFocused checks if a widget with the given label is focused.
func (a *Agent) IsFocused(label string) bool {
	w := a.FindByLabel(label)
	return w != nil && w.Focused
}

// IsEnabled checks if a widget with the given label is enabled.
func (a *Agent) IsEnabled(label string) bool {
	w := a.FindByLabel(label)
	return w != nil && !w.State.Disabled
}

// IsChecked checks if a checkbox/radio with the given label is checked.
func (a *Agent) IsChecked(label string) bool {
	w := a.FindByLabel(label)
	if w == nil || w.State.Checked == nil {
		return false
	}
	return *w.State.Checked
}

// GetValue returns the value of an input widget.
func (a *Agent) GetValue(label string) (string, error) {
	w := a.FindByLabel(label)
	if w == nil {
		return "", ErrWidgetNotFound
	}
	return w.Value, nil
}

// SnapshotJSON returns the current snapshot serialized to JSON.
func (a *Agent) SnapshotJSON() ([]byte, error) {
	snap, err := a.SnapshotWithContext(context.Background(), SnapshotOptions{
		IncludeText: a.includeText,
	})
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(snap, "", "  ")
}

// FocusWidget focuses the widget with the given label.
func (a *Agent) FocusWidget(label string) error {
	return a.Focus(label)
}

// Focus moves focus to the widget with the given label.
func (a *Agent) Focus(label string) error {
	info := a.FindByLabel(label)
	if info == nil {
		return ErrWidgetNotFound
	}
	if info.State.Disabled {
		return ErrWidgetDisabled
	}
	return a.focusByID(info.ID)
}

// ActivateWidget activates the widget with the given label.
func (a *Agent) ActivateWidget(label string) error {
	return a.Activate(label)
}

// Activate focuses and activates the widget with the given label.
func (a *Agent) Activate(label string) error {
	info := a.FindByLabel(label)
	if info == nil {
		return ErrWidgetNotFound
	}
	if info.State.Disabled {
		return ErrWidgetDisabled
	}
	if err := a.focusByID(info.ID); err != nil {
		return err
	}
	if err := a.sendKey(terminal.KeyEnter, 0); err != nil {
		return err
	}
	a.Tick()
	return nil
}

// TypeInto focuses the widget and types the given text.
func (a *Agent) TypeInto(label, text string) error {
	return a.Type(label, text)
}

// Type focuses the widget and types the given text.
func (a *Agent) Type(label, text string) error {
	info := a.FindByLabel(label)
	if info == nil {
		return ErrWidgetNotFound
	}
	if info.State.Disabled {
		return ErrWidgetDisabled
	}
	if err := a.focusByID(info.ID); err != nil {
		return err
	}
	if err := a.sendText(text); err != nil {
		return err
	}
	a.Tick()
	return nil
}

// Select focuses the widget and selects the option by label.
func (a *Agent) Select(label, option string) error {
	info := a.FindByLabel(label)
	if info == nil {
		return ErrWidgetNotFound
	}
	if info.State.Disabled {
		return ErrWidgetDisabled
	}

	w, acc, err := a.focusWidgetByID(info.ID)
	if err != nil {
		return err
	}
	if acc == nil {
		return ErrNotInteractive
	}

	current := accessibleChoice(acc)
	if strings.EqualFold(current, option) {
		return nil
	}

	seen := map[string]bool{current: true}
	for i := 0; i < 100; i++ {
		if err := a.sendKey(terminal.KeyDown, 0); err != nil {
			return err
		}
		a.Tick()
		current = accessibleChoice(acc)
		if strings.EqualFold(current, option) {
			_ = w
			return nil
		}
		if seen[current] {
			break
		}
		seen[current] = true
	}
	return ErrWidgetNotFound
}

func accessibleChoice(acc accessibility.Accessible) string {
	if acc == nil {
		return ""
	}
	if val := acc.AccessibleValue(); val != nil && strings.TrimSpace(val.Text) != "" {
		return val.Text
	}
	return acc.AccessibleLabel()
}

// SendKeyMsg injects a raw key message into the app.
func (a *Agent) SendKeyMsg(msg runtime.KeyMsg) error {
	if a == nil {
		return ErrNoApp
	}
	a.mu.Lock()
	simBackend := a.sim
	app := a.app
	postKey := a.postKey
	a.mu.Unlock()

	// Use custom PostKey callback if provided (for apps with custom event loops)
	if postKey != nil {
		return postKey(msg)
	}
	if simBackend != nil {
		return simBackend.PostEvent(terminal.KeyEvent{
			Key:   msg.Key,
			Rune:  msg.Rune,
			Alt:   msg.Alt,
			Ctrl:  msg.Ctrl,
			Shift: msg.Shift,
		})
	}
	if app != nil {
		app.Post(msg)
		return nil
	}
	return ErrNoApp
}

// SendKey injects a key into the app.
func (a *Agent) SendKey(key terminal.Key) error {
	if err := a.SendKeyMsg(runtime.KeyMsg{Key: key}); err != nil {
		return err
	}
	a.Tick()
	return nil
}

// SendKeyRune injects a key with rune payload.
func (a *Agent) SendKeyRune(key terminal.Key, r rune) error {
	if err := a.SendKeyMsg(runtime.KeyMsg{Key: key, Rune: r}); err != nil {
		return err
	}
	a.Tick()
	return nil
}

// SendKeyString injects a string as a sequence of key events.
func (a *Agent) SendKeyString(text string) error {
	if err := a.sendText(text); err != nil {
		return err
	}
	a.Tick()
	return nil
}

// SendMouse injects a mouse event into the app.
func (a *Agent) SendMouse(msg runtime.MouseMsg) error {
	if a == nil {
		return ErrNoApp
	}
	a.mu.Lock()
	simBackend := a.sim
	app := a.app
	a.mu.Unlock()

	if simBackend != nil {
		return simBackend.PostEvent(terminal.MouseEvent{
			X:      msg.X,
			Y:      msg.Y,
			Button: terminal.MouseButton(msg.Button),
			Action: terminal.MouseAction(msg.Action),
			Alt:    msg.Alt,
			Ctrl:   msg.Ctrl,
			Shift:  msg.Shift,
		})
	}
	if app != nil {
		app.Post(msg)
		return nil
	}
	return ErrNoApp
}

// SendPaste injects a paste event into the app.
func (a *Agent) SendPaste(text string) error {
	if a == nil {
		return ErrNoApp
	}
	a.mu.Lock()
	simBackend := a.sim
	app := a.app
	a.mu.Unlock()

	if simBackend != nil {
		return simBackend.PostEvent(terminal.PasteEvent{Text: text})
	}
	if app != nil {
		app.Post(runtime.PasteMsg{Text: text})
		return nil
	}
	return ErrNoApp
}

// SendResize injects a resize event into the app.
func (a *Agent) SendResize(width, height int) error {
	if a == nil {
		return ErrNoApp
	}
	a.mu.Lock()
	simBackend := a.sim
	app := a.app
	a.mu.Unlock()

	if simBackend != nil {
		simBackend.InjectResize(width, height)
		return nil
	}
	if app != nil {
		app.Post(runtime.ResizeMsg{Width: width, Height: height})
		return nil
	}
	return ErrNoApp
}

// WaitForText waits until text appears on screen or timeout occurs.
func (a *Agent) WaitForText(text string, timeout time.Duration) error {
	if a == nil {
		return ErrNoApp
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if a.ContainsText(text) {
			return nil
		}
		a.Tick()
	}
	return ErrTimeout
}

// WaitForWidget waits until a widget with the given label is present.
func (a *Agent) WaitForWidget(label string, timeout time.Duration) error {
	if a == nil {
		return ErrNoApp
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if a.FindByLabel(label) != nil {
			return nil
		}
		a.Tick()
	}
	return ErrTimeout
}

// ListWidgets returns widgets that match the given role.
func (a *Agent) ListWidgets(role accessibility.Role) []WidgetInfo {
	return a.FindByRole(role)
}

// ContainsText checks if the given text appears on screen.
func (a *Agent) ContainsText(text string) bool {
	if a == nil {
		return false
	}
	return strings.Contains(a.captureText(), text)
}

// FindText returns the position of text on screen, or (-1, -1) if not found.
func (a *Agent) FindText(text string) (x, y int) {
	if a == nil {
		return -1, -1
	}
	return findTextIn(a.captureText(), text)
}

func findTextIn(content, text string) (x, y int) {
	lines := strings.Split(content, "\n")
	for row, line := range lines {
		if col := strings.Index(line, text); col >= 0 {
			return col, row
		}
	}
	return -1, -1
}

// CaptureText returns the raw text content of the screen.
func (a *Agent) CaptureText() string {
	if a == nil {
		return ""
	}
	return a.captureText()
}

func (a *Agent) sendKey(key terminal.Key, r rune) error {
	return a.SendKeyMsg(runtime.KeyMsg{Key: key, Rune: r})
}

func (a *Agent) sendText(text string) error {
	if a == nil {
		return ErrNoApp
	}
	for _, r := range text {
		switch r {
		case '\n':
			if err := a.sendKey(terminal.KeyEnter, 0); err != nil {
				return err
			}
		case '\t':
			if err := a.sendKey(terminal.KeyTab, 0); err != nil {
				return err
			}
		default:
			if err := a.sendKey(terminal.KeyRune, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *Agent) captureText() string {
	if a == nil {
		return ""
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.captureTextLocked()
}

func (a *Agent) captureTextLocked() string {
	a.ensureScreenLocked()
	if a.sim != nil {
		return a.sim.Capture()
	}
	if a.app != nil {
		return a.app.SnapshotText()
	}
	if a.screen != nil {
		if buf := a.screen.Buffer(); buf != nil {
			return buf.SnapshotText()
		}
	}
	return ""
}

func (a *Agent) focusByID(id string) error {
	_, _, err := a.focusWidgetByID(id)
	return err
}

func (a *Agent) focusWidgetByID(id string) (runtime.Widget, accessibility.Accessible, error) {
	if a == nil {
		return nil, nil, ErrNoApp
	}
	if a.app != nil {
		var (
			w   runtime.Widget
			acc accessibility.Accessible
		)
		err := a.app.Call(context.Background(), func(app *runtime.App) error {
			a.mu.Lock()
			defer a.mu.Unlock()
			if a.autoAttach && a.screen == nil {
				a.screen = app.Screen()
			}
			var err error
			w, acc, err = a.focusWidgetByIDLocked(id)
			return err
		})
		if err != nil {
			return nil, nil, err
		}
		return w, acc, nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	return a.focusWidgetByIDLocked(id)
}

func (a *Agent) focusWidgetByIDLocked(id string) (runtime.Widget, accessibility.Accessible, error) {
	screen := a.ensureScreenLocked()
	if screen == nil {
		return nil, nil, ErrNoApp
	}

	// Search all layers for the widget (top to bottom for overlays first)
	var w runtime.Widget
	var foundLayer *runtime.Layer
	for i := screen.LayerCount() - 1; i >= 0; i-- {
		layer := screen.Layer(i)
		if layer == nil || layer.Root == nil {
			continue
		}
		explicitCounts := make(map[string]int)
		if found := findWidgetByID(layer.Root, id, i, []int{0}, explicitCounts); found != nil {
			w = found
			foundLayer = layer
			break
		}
	}

	if w == nil {
		return nil, nil, ErrWidgetNotFound
	}
	focusable, ok := w.(runtime.Focusable)
	if !ok || !focusable.CanFocus() {
		return w, accessibleFromWidget(w), ErrNotFocusable
	}

	// Use the focus scope from the layer where we found the widget
	scope := foundLayer.FocusScope
	if scope == nil {
		return w, accessibleFromWidget(w), ErrNotFocusable
	}

	if scope.SetFocus(focusable) || scope.Current() == focusable {
		return w, accessibleFromWidget(w), nil
	}

	scope.Reset()
	runtime.RegisterFocusables(scope, foundLayer.Root)
	if scope.SetFocus(focusable) || scope.Current() == focusable {
		return w, accessibleFromWidget(w), nil
	}
	return w, accessibleFromWidget(w), ErrNotFocusable
}

func accessibleFromWidget(w runtime.Widget) accessibility.Accessible {
	if w == nil {
		return nil
	}
	if acc, ok := w.(accessibility.Accessible); ok {
		return acc
	}
	return nil
}

func findWidgetByID(w runtime.Widget, id string, layer int, path []int, explicitCounts map[string]int) runtime.Widget {
	if w == nil {
		return nil
	}
	if buildWidgetID(w, layer, path, explicitCounts, false) == id {
		return w
	}
	if cp, ok := w.(runtime.ChildProvider); ok {
		children := cp.ChildWidgets()
		for i, child := range children {
			if child == nil {
				continue
			}
			childPath := make([]int, len(path)+1)
			copy(childPath, path)
			childPath[len(path)] = i
			if found := findWidgetByID(child, id, layer, childPath, explicitCounts); found != nil {
				return found
			}
		}
	}
	return nil
}
