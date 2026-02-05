package agent

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/backend/sim"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

type testInput struct {
	mu       sync.RWMutex
	bounds   runtime.Rect
	focused  bool
	label    string
	value    string
	disabled bool
}

func (t *testInput) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.Constrain(runtime.Size{Width: 12, Height: 1})
}

func (t *testInput) Layout(bounds runtime.Rect) {
	t.mu.Lock()
	t.bounds = bounds
	t.mu.Unlock()
}

func (t *testInput) Render(ctx runtime.RenderContext) {
	if ctx.Buffer == nil {
		return
	}
	t.mu.RLock()
	text := t.value
	bounds := t.bounds
	t.mu.RUnlock()
	if text == "" {
		text = " "
	}
	ctx.Buffer.SetString(bounds.X, bounds.Y, text, backend.DefaultStyle())
}

func (t *testInput) HandleMessage(msg runtime.Message) runtime.HandleResult {
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
	case terminal.KeyBackspace:
		if len(t.value) > 0 {
			t.value = t.value[:len(t.value)-1]
		}
		return runtime.Handled()
	case terminal.KeyRune:
		if key.Rune != 0 {
			t.value += string(key.Rune)
			return runtime.Handled()
		}
	}
	return runtime.Unhandled()
}

func (t *testInput) Bounds() runtime.Rect {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.bounds
}

func (t *testInput) CanFocus() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return !t.disabled
}
func (t *testInput) Focus() {
	t.mu.Lock()
	t.focused = true
	t.mu.Unlock()
}
func (t *testInput) Blur() {
	t.mu.Lock()
	t.focused = false
	t.mu.Unlock()
}
func (t *testInput) IsFocused() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.focused
}

func (t *testInput) AccessibleRole() accessibility.Role { return accessibility.RoleTextbox }
func (t *testInput) AccessibleLabel() string            { return t.label }
func (t *testInput) AccessibleDescription() string      { return "" }
func (t *testInput) AccessibleState() accessibility.StateSet {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return accessibility.StateSet{Disabled: t.disabled}
}
func (t *testInput) AccessibleValue() *accessibility.ValueInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return &accessibility.ValueInfo{Text: t.value}
}
func (t *testInput) Text() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.value
}

type testButton struct {
	mu       sync.RWMutex
	bounds   runtime.Rect
	focused  bool
	label    string
	disabled bool
	clicked  bool
}

func (t *testButton) Measure(constraints runtime.Constraints) runtime.Size {
	return constraints.Constrain(runtime.Size{Width: len(t.label) + 2, Height: 1})
}

func (t *testButton) Layout(bounds runtime.Rect) {
	t.mu.Lock()
	t.bounds = bounds
	t.mu.Unlock()
}

func (t *testButton) Render(ctx runtime.RenderContext) {
	if ctx.Buffer == nil {
		return
	}
	t.mu.RLock()
	bounds := t.bounds
	label := t.label
	t.mu.RUnlock()
	ctx.Buffer.SetString(bounds.X, bounds.Y, "["+label+"]", backend.DefaultStyle())
}

func (t *testButton) HandleMessage(msg runtime.Message) runtime.HandleResult {
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
		t.clicked = true
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (t *testButton) Bounds() runtime.Rect {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.bounds
}

func (t *testButton) CanFocus() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return !t.disabled
}
func (t *testButton) Focus() {
	t.mu.Lock()
	t.focused = true
	t.mu.Unlock()
}
func (t *testButton) Blur() {
	t.mu.Lock()
	t.focused = false
	t.mu.Unlock()
}
func (t *testButton) IsFocused() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.focused
}

func (t *testButton) AccessibleRole() accessibility.Role { return accessibility.RoleButton }
func (t *testButton) AccessibleLabel() string            { return t.label }
func (t *testButton) AccessibleDescription() string      { return "" }
func (t *testButton) AccessibleState() accessibility.StateSet {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return accessibility.StateSet{Disabled: t.disabled}
}
func (t *testButton) AccessibleValue() *accessibility.ValueInfo { return nil }

func (t *testButton) Clicked() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.clicked
}

func TestAgentSnapshotAndActions(t *testing.T) {
	input := &testInput{label: "Name"}
	button := &testButton{label: "Submit"}
	root := runtime.VBox(runtime.Fixed(input), runtime.Fixed(button)).WithGap(1)

	simBackend := sim.New(40, 10)
	app := runtime.NewApp(runtime.AppConfig{
		Backend:           simBackend,
		Root:              root,
		Update:            runtime.DefaultUpdate,
		FocusRegistration: runtime.FocusRegistrationAuto,
		TickRate:          time.Second / 60,
	})

	agt := New(Config{App: app})

	runAppForTest(t, app)

	if err := agt.WaitForWidget("Name", time.Second); err != nil {
		t.Fatalf("wait for widget: %v", err)
	}

	if err := agt.Focus("Name"); err != nil {
		t.Fatalf("focus name: %v", err)
	}

	if err := agt.Type("Name", "Alice"); err != nil {
		t.Fatalf("type name: %v", err)
	}

	value, err := agt.GetValue("Name")
	if err != nil {
		t.Fatalf("get value: %v", err)
	}
	if value != "Alice" {
		t.Fatalf("value = %q, want %q", value, "Alice")
	}

	if err := agt.Activate("Submit"); err != nil {
		t.Fatalf("activate submit: %v", err)
	}
	if !button.Clicked() {
		t.Fatal("expected submit to be activated")
	}

	raw, err := agt.SnapshotJSON()
	if err != nil {
		t.Fatalf("snapshot json: %v", err)
	}
	if !strings.Contains(string(raw), "\"widgets\"") {
		t.Fatalf("snapshot json missing widgets: %s", string(raw))
	}
}
