package agent

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/widgets"
)

type testCheckbox struct {
	mu       sync.RWMutex
	bounds   runtime.Rect
	focused  bool
	label    string
	checked  bool
	disabled bool
}

func (t *testCheckbox) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.Constrain(runtime.Size{Width: len(t.label) + 4, Height: 1})
}

func (t *testCheckbox) Layout(bounds runtime.Rect) {
	t.mu.Lock()
	t.bounds = bounds
	t.mu.Unlock()
}

func (t *testCheckbox) Render(ctx runtime.RenderContext) {
	if ctx.Buffer == nil {
		return
	}
	t.mu.RLock()
	mark := " "
	if t.checked {
		mark = "x"
	}
	bounds := t.bounds
	label := t.label
	t.mu.RUnlock()
	ctx.Buffer.SetString(bounds.X, bounds.Y, "["+mark+"] "+label, backend.DefaultStyle())
}

func (t *testCheckbox) HandleMessage(msg runtime.Message) runtime.HandleResult {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.disabled || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	if key.Key == terminal.KeyEnter || (key.Key == terminal.KeyRune && key.Rune == ' ') {
		t.checked = !t.checked
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (t *testCheckbox) Bounds() runtime.Rect {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.bounds
}

func (t *testCheckbox) CanFocus() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return !t.disabled
}
func (t *testCheckbox) Focus() {
	t.mu.Lock()
	t.focused = true
	t.mu.Unlock()
}
func (t *testCheckbox) Blur() {
	t.mu.Lock()
	t.focused = false
	t.mu.Unlock()
}
func (t *testCheckbox) IsFocused() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.focused
}

func (t *testCheckbox) AccessibleRole() accessibility.Role { return accessibility.RoleCheckbox }
func (t *testCheckbox) AccessibleLabel() string            { return t.label }
func (t *testCheckbox) AccessibleDescription() string      { return "" }
func (t *testCheckbox) AccessibleState() accessibility.StateSet {
	t.mu.RLock()
	disabled := t.disabled
	checked := t.checked
	t.mu.RUnlock()
	return accessibility.StateSet{Disabled: disabled, Checked: &checked}
}
func (t *testCheckbox) AccessibleValue() *accessibility.ValueInfo { return nil }

type testList struct {
	mu       sync.RWMutex
	bounds   runtime.Rect
	focused  bool
	label    string
	options  []string
	index    int
	disabled bool
}

func (t *testList) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.Constrain(runtime.Size{Width: 12, Height: 1})
}

func (t *testList) Layout(bounds runtime.Rect) {
	t.mu.Lock()
	t.bounds = bounds
	t.mu.Unlock()
}

func (t *testList) Render(ctx runtime.RenderContext) {
	if ctx.Buffer == nil {
		return
	}
	t.mu.RLock()
	text := t.label + ": " + t.currentLocked()
	bounds := t.bounds
	t.mu.RUnlock()
	ctx.Buffer.SetString(bounds.X, bounds.Y, text, backend.DefaultStyle())
}

func (t *testList) HandleMessage(msg runtime.Message) runtime.HandleResult {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.disabled || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyDown:
		t.advanceLocked(1)
		return runtime.Handled()
	case terminal.KeyUp:
		t.advanceLocked(-1)
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (t *testList) advanceLocked(delta int) {
	if len(t.options) == 0 {
		return
	}
	t.index = (t.index + delta) % len(t.options)
	if t.index < 0 {
		t.index += len(t.options)
	}
}

func (t *testList) currentLocked() string {
	if len(t.options) == 0 || t.index < 0 || t.index >= len(t.options) {
		return ""
	}
	return t.options[t.index]
}

func (t *testList) current() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.currentLocked()
}

func (t *testList) Bounds() runtime.Rect {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.bounds
}

func (t *testList) CanFocus() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return !t.disabled
}
func (t *testList) Focus() {
	t.mu.Lock()
	t.focused = true
	t.mu.Unlock()
}
func (t *testList) Blur() {
	t.mu.Lock()
	t.focused = false
	t.mu.Unlock()
}
func (t *testList) IsFocused() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.focused
}

func (t *testList) AccessibleRole() accessibility.Role { return accessibility.RoleList }
func (t *testList) AccessibleLabel() string            { return t.label }
func (t *testList) AccessibleDescription() string      { return "" }
func (t *testList) AccessibleState() accessibility.StateSet {
	t.mu.RLock()
	disabled := t.disabled
	t.mu.RUnlock()
	return accessibility.StateSet{Disabled: disabled}
}
func (t *testList) AccessibleValue() *accessibility.ValueInfo {
	return &accessibility.ValueInfo{Text: t.current()}
}

type recordWidget struct {
	mu        sync.Mutex
	bounds    runtime.Rect
	lastKey   runtime.KeyMsg
	typed     string
	lastMouse runtime.MouseMsg
	lastPaste string
}

func (r *recordWidget) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.Constrain(runtime.Size{Width: 20, Height: 2})
}

func (r *recordWidget) Layout(bounds runtime.Rect) { r.bounds = bounds }

func (r *recordWidget) Render(ctx runtime.RenderContext) {
	if ctx.Buffer == nil {
		return
	}
	ctx.Buffer.SetString(r.bounds.X, r.bounds.Y, "Hello", backend.DefaultStyle())
}

func (r *recordWidget) HandleMessage(msg runtime.Message) runtime.HandleResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch m := msg.(type) {
	case runtime.KeyMsg:
		r.lastKey = m
		switch m.Key {
		case terminal.KeyRune:
			if m.Rune != 0 {
				r.typed += string(m.Rune)
			}
		case terminal.KeyEnter:
			r.typed += "\n"
		case terminal.KeyTab:
			r.typed += "\t"
		}
		return runtime.Handled()
	case runtime.MouseMsg:
		r.lastMouse = m
		return runtime.Handled()
	case runtime.PasteMsg:
		r.lastPaste = m.Text
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (r *recordWidget) Bounds() runtime.Rect { return r.bounds }

func (r *recordWidget) snapshot() (runtime.KeyMsg, string, runtime.MouseMsg, string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lastKey, r.typed, r.lastMouse, r.lastPaste
}

func (r *recordWidget) resetTyped() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.typed = ""
}

func TestAgentFinderState(t *testing.T) {
	input := &testInput{label: "Name"}
	button := &testButton{label: "Submit"}
	checkbox := &testCheckbox{label: "Agree", checked: true}
	root := runtime.VBox(runtime.Fixed(input), runtime.Fixed(button), runtime.Fixed(checkbox)).WithGap(1)

	be := sim.New(40, 10)
	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              root,
		Update:            runtime.DefaultUpdate,
		FocusRegistration: runtime.FocusRegistrationAuto,
		TickRate:          time.Second / 60,
	})

	agt := New(Config{App: app, TickRate: time.Millisecond})
	runAppForTest(t, app)

	if err := agt.WaitForWidget("Name", time.Second); err != nil {
		t.Fatalf("wait for widget: %v", err)
	}

	if agt.Screen() == nil {
		t.Fatal("expected screen to be available")
	}

	agt.SetScreen(app.Screen())
	if agt.Screen() == nil {
		t.Fatal("expected screen after SetScreen")
	}

	if err := agt.FocusWidget("Name"); err != nil {
		t.Fatalf("focus widget: %v", err)
	}

	if focused := agt.GetFocused(); focused == nil || focused.Label != "Name" {
		t.Fatalf("focused = %+v", focused)
	}

	if !agt.IsFocused("Name") {
		t.Fatalf("expected Name to be focused")
	}

	if !agt.IsEnabled("Name") {
		t.Fatalf("expected Name to be enabled")
	}

	if !agt.IsChecked("Agree") {
		t.Fatalf("expected Agree to be checked")
	}

	buttons := agt.FindByRole(accessibility.RoleButton)
	if len(buttons) != 1 || buttons[0].Label != "Submit" {
		t.Fatalf("unexpected buttons: %+v", buttons)
	}

	typed := agt.FindByType(accessibility.RoleButton)
	if len(typed) != 1 || typed[0].Label != "Submit" {
		t.Fatalf("unexpected type results: %+v", typed)
	}

	listed := agt.ListWidgets(accessibility.RoleButton)
	if len(listed) != 1 || listed[0].Label != "Submit" {
		t.Fatalf("unexpected listed results: %+v", listed)
	}

	nameInfo := agt.FindByLabel("Name")
	if nameInfo == nil {
		t.Fatal("expected Name widget info")
	}

	byID := agt.FindByID(nameInfo.ID)
	if byID == nil || byID.Label != "Name" {
		t.Fatalf("find by id failed: %+v", byID)
	}

	if err := agt.WithWidgetByID(nil, nameInfo.ID, func(w runtime.Widget, _ accessibility.Accessible) error {
		if w != input {
			return errors.New("unexpected widget")
		}
		return nil
	}); err != nil {
		t.Fatalf("with widget by id: %v", err)
	}

	if err := agt.WithWidgetByID(nil, " ", func(runtime.Widget, accessibility.Accessible) error {
		return nil
	}); !errors.Is(err, ErrWidgetNotFound) {
		t.Fatalf("expected ErrWidgetNotFound, got %v", err)
	}

	if err := agt.ClearFocus(); err != nil {
		t.Fatalf("clear focus: %v", err)
	}
	agt.Tick()
	if agt.IsFocused("Name") {
		t.Fatalf("expected focus to be cleared")
	}

	if err := agt.FocusByID(nameInfo.ID); err != nil {
		t.Fatalf("focus by id: %v", err)
	}
	if !agt.IsFocused("Name") {
		t.Fatalf("expected Name to be focused after FocusByID")
	}
}

func TestAgentActionsAndWaits(t *testing.T) {
	input := &testInput{label: "Name"}
	button := &testButton{label: "Submit"}
	choices := &testList{label: "Choice", options: []string{"A", "B", "C"}}
	root := runtime.VBox(runtime.Fixed(input), runtime.Fixed(button), runtime.Fixed(choices)).WithGap(1)

	be := sim.New(60, 12)
	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              root,
		Update:            runtime.DefaultUpdate,
		FocusRegistration: runtime.FocusRegistrationAuto,
		TickRate:          time.Second / 60,
	})

	agt := New(Config{App: app, TickRate: time.Millisecond})
	runAppForTest(t, app)

	if err := agt.TypeInto("Name", "Ada"); err != nil {
		t.Fatalf("type into: %v", err)
	}
	if input.Text() != "Ada" {
		t.Fatalf("value = %q", input.Text())
	}
	if err := agt.WaitForValue("Name", "Ada", time.Second); err != nil {
		t.Fatalf("wait for value: %v", err)
	}

	if err := agt.ActivateWidget("Submit"); err != nil {
		t.Fatalf("activate widget: %v", err)
	}
	if !button.Clicked() {
		t.Fatal("expected submit to be clicked")
	}

	if err := agt.Select("Choice", "C"); err != nil {
		t.Fatalf("select choice: %v", err)
	}
	if choices.current() != "C" {
		t.Fatalf("choice = %q", choices.current())
	}

	choiceInfo := agt.FindByLabel("Choice")
	if choiceInfo == nil {
		t.Fatal("expected choice info")
	}
	if err := agt.SelectByID(choiceInfo.ID, "B"); err != nil {
		t.Fatalf("select by id: %v", err)
	}
	if choices.current() != "B" {
		t.Fatalf("choice = %q", choices.current())
	}

	if err := agt.WaitForWidget("Choice", time.Second); err != nil {
		t.Fatalf("wait for widget: %v", err)
	}
	if err := agt.WaitForText("Submit", time.Second); err != nil {
		t.Fatalf("wait for text: %v", err)
	}

	if err := agt.Focus("Name"); err != nil {
		t.Fatalf("focus: %v", err)
	}
	if err := agt.WaitForFocus("Name", time.Second); err != nil {
		t.Fatalf("wait for focus: %v", err)
	}

	if err := agt.WaitForEnabled("Name", time.Second); err != nil {
		t.Fatalf("wait for enabled: %v", err)
	}

	if err := agt.WaitForWidgetGone("Missing", 50*time.Millisecond); err != nil {
		t.Fatalf("wait for widget gone: %v", err)
	}
	if err := agt.WaitForTextGone("MissingText", 50*time.Millisecond); err != nil {
		t.Fatalf("wait for text gone: %v", err)
	}

	if err := agt.WaitForIdle(200 * time.Millisecond); err != nil {
		t.Fatalf("wait for idle: %v", err)
	}
}

func TestAgentSendersAndCapture(t *testing.T) {
	recorder := &recordWidget{}
	be := sim.New(20, 6)
	app := runtime.NewApp(runtime.AppConfig{
		Backend:           be,
		Root:              recorder,
		Update:            runtime.DefaultUpdate,
		FocusRegistration: runtime.FocusRegistrationAuto,
		TickRate:          time.Second / 60,
	})

	agt := New(Config{App: app, Sim: be, TickRate: time.Millisecond})
	runAppForTest(t, app)

	if err := agt.WaitForText("Hello", time.Second); err != nil {
		t.Fatalf("wait for text: %v", err)
	}

	if !agt.ContainsText("Hello") {
		t.Fatal("expected Hello in text")
	}

	if x, y := agt.FindText("Hello"); x != 0 || y != 0 {
		t.Fatalf("FindText = (%d,%d)", x, y)
	}

	captured := agt.CaptureText()
	if !strings.Contains(captured, "Hello") {
		t.Fatalf("capture text missing Hello: %q", captured)
	}

	if x, y := findTextIn("one\ntwo\nthree", "two"); x != 0 || y != 1 {
		t.Fatalf("findTextIn = (%d,%d)", x, y)
	}

	region := agt.CaptureRegion(0, 0, 5, 1)
	if region != "Hello" {
		t.Fatalf("region = %q", region)
	}

	cell, ok := agt.CellAt(0, 0)
	if !ok || cell.Rune != 'H' {
		t.Fatalf("cell = %+v ok=%v", cell, ok)
	}

	width, height := agt.Dimensions()
	if width <= 0 || height <= 0 {
		t.Fatalf("dimensions = %dx%d", width, height)
	}

	newWidth := width + 10
	newHeight := height + 2
	if err := agt.SendResize(newWidth, newHeight); err != nil {
		t.Fatalf("send resize: %v", err)
	}
	agt.Tick()
	width, height = agt.Dimensions()
	if width != newWidth || height != newHeight {
		t.Fatalf("dimensions after resize = %dx%d", width, height)
	}

	if err := agt.SendKey(terminal.KeyEnter); err != nil {
		t.Fatalf("send key: %v", err)
	}
	if lastKey, _, _, _ := recorder.snapshot(); lastKey.Key != terminal.KeyEnter {
		t.Fatalf("last key = %v", lastKey.Key)
	}

	if err := agt.SendKeyRune(terminal.KeyRune, 'x'); err != nil {
		t.Fatalf("send key rune: %v", err)
	}
	if lastKey, _, _, _ := recorder.snapshot(); lastKey.Rune != 'x' {
		t.Fatalf("last rune = %q", lastKey.Rune)
	}

	recorder.resetTyped()
	if err := agt.SendKeyString("ab\tc\n"); err != nil {
		t.Fatalf("send key string: %v", err)
	}
	if _, typed, _, _ := recorder.snapshot(); typed != "ab\tc\n" {
		t.Fatalf("typed = %q", typed)
	}

	if err := agt.SendMouse(runtime.MouseMsg{X: 0, Y: 0, Button: runtime.MouseLeft, Action: runtime.MousePress}); err != nil {
		t.Fatalf("send mouse: %v", err)
	}
	agt.Tick()
	if _, _, lastMouse, _ := recorder.snapshot(); lastMouse.Button != runtime.MouseLeft {
		t.Fatalf("last mouse button = %v", lastMouse.Button)
	}

	if err := agt.SendPaste("pasted"); err != nil {
		t.Fatalf("send paste: %v", err)
	}
}

func TestAgentSnapshotsAndActionsHelpers(t *testing.T) {
	actions := []string{"focus", "activate", "focus"}
	trimmed := removeAction(actions, "focus")
	if len(trimmed) != 1 || trimmed[0] != "activate" {
		t.Fatalf("removeAction = %#v", trimmed)
	}

	empty := removeAction([]string{"focus"}, "focus")
	if empty != nil {
		t.Fatalf("expected nil actions")
	}

	snap := Snapshot{
		Width:      10,
		Height:     5,
		LayerCount: 1,
		FocusedID:  "id-1",
		Text:       "hello",
		Widgets: []WidgetInfo{
			{
				ID:        "id-1",
				Role:      accessibility.RoleButton,
				Label:     "Ok",
				Bounds:    runtime.Rect{X: 0, Y: 0, Width: 2, Height: 1},
				Actions:   []string{"activate"},
				Focusable: true,
				Focused:   true,
			},
		},
	}

	if !snapshotsEqual(snap, snap) {
		t.Fatal("expected snapshots to be equal")
	}

	changed := snap
	changed.Text = "bye"
	if snapshotsEqual(snap, changed) {
		t.Fatal("expected snapshots to differ")
	}

	alt := []WidgetInfo{{ID: "id-1", Label: "Ok"}}
	if widgetsEqual(snap.Widgets, alt) {
		t.Fatal("expected widgets to differ")
	}
}

func TestAgentBackendAndSimPaste(t *testing.T) {
	be := sim.New(10, 5)
	agt := New(Config{Sim: be})
	if agt.Backend() != be {
		t.Fatal("expected backend to match sim")
	}
	if err := agt.SendPaste("noop"); err != nil {
		t.Fatalf("send paste: %v", err)
	}
}

func TestAppHookIntegration(t *testing.T) {
	be := sim.New(20, 6)
	app := runtime.NewApp(runtime.AppConfig{Backend: be, Root: widgets.NewLabel("App")})
	runAppForTest(t, app)

	opts := DefaultEnhancedServerOptions()
	opts.Addr = "unix:" + filepath.Join(t.TempDir(), "agent.sock")
	opts.App = app

	integration, err := NewAppIntegration(app, opts)
	if err != nil {
		t.Fatalf("new app integration: %v", err)
	}

	if err := integration.Start(); err != nil {
		t.Fatalf("start integration: %v", err)
	}
	t.Cleanup(func() {
		_ = integration.Stop()
	})

	if integration.Hook == nil {
		t.Fatal("expected hook")
	}

	rendered := 0
	integration.Hook.OnRender(func() { rendered++ })
	integration.Hook.OnMessage(func(runtime.Message) {})

	integration.Hook.handleRender(runtime.RenderStats{Frame: 3})
	if rendered != 1 {
		t.Fatalf("rendered = %d", rendered)
	}

	integration.Hook.Uninstall()
	if len(integration.Hook.onRender) != 0 || len(integration.Hook.onMessage) != 0 {
		t.Fatal("expected hook callbacks to be cleared")
	}

	if err := (&AppIntegration{}).Start(); err != nil {
		t.Fatalf("expected nil start error, got %v", err)
	}
	if err := (&AppIntegration{}).Stop(); err != nil {
		t.Fatalf("expected nil stop error, got %v", err)
	}
}
