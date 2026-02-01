package keybind

import (
	"strings"
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/clipboard"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/scroll"
	"github.com/odvcencio/fluffyui/terminal"
)

type testWidget struct {
	role  accessibility.Role
	label string
}

func (t *testWidget) Measure(runtime.Constraints) runtime.Size { return runtime.Size{} }
func (t *testWidget) Layout(runtime.Rect)                      {}
func (t *testWidget) Render(runtime.RenderContext)             {}
func (t *testWidget) HandleMessage(runtime.Message) runtime.HandleResult {
	return runtime.Unhandled()
}
func (t *testWidget) AccessibleRole() accessibility.Role { return t.role }
func (t *testWidget) AccessibleLabel() string            { return t.label }
func (t *testWidget) AccessibleDescription() string      { return "" }
func (t *testWidget) AccessibleState() accessibility.StateSet {
	return accessibility.StateSet{}
}
func (t *testWidget) AccessibleValue() *accessibility.ValueInfo { return nil }

type clipboardWidget struct {
	testWidget
	pasted []string
}

func (c *clipboardWidget) ClipboardCopy() (string, bool) { return "copy", true }
func (c *clipboardWidget) ClipboardCut() (string, bool)  { return "cut", true }
func (c *clipboardWidget) ClipboardPaste(text string) bool {
	c.pasted = append(c.pasted, text)
	return true
}

type scrollWidget struct {
	testWidget
	scrollDx int
	scrollDy int
	pageBy   int
	atStart  bool
	atEnd    bool
}

func (s *scrollWidget) ScrollBy(dx, dy int) {
	s.scrollDx += dx
	s.scrollDy += dy
}
func (s *scrollWidget) ScrollTo(x, y int) {}
func (s *scrollWidget) PageBy(pages int)  { s.pageBy += pages }
func (s *scrollWidget) ScrollToStart()    { s.atStart = true }
func (s *scrollWidget) ScrollToEnd()      { s.atEnd = true }

type handlerWidget struct {
	testWidget
	HandlerBase
}

func TestConditionsAndCombinators(t *testing.T) {
	ctx := Context{}
	if WhenFocused()(ctx) {
		t.Fatalf("expected not focused")
	}
	if !WhenFocusedNotClipboardTarget()(ctx) {
		t.Fatalf("expected not clipboard target when unfocused")
	}

	clip := &clipboardWidget{}
	clip.role = accessibility.RoleButton
	ctx.Focused = clip
	ctx.FocusedWidget = clip

	if !WhenFocused()(ctx) {
		t.Fatalf("expected focused")
	}
	if !WhenFocusedClipboardTarget()(ctx) {
		t.Fatalf("expected clipboard target")
	}
	if WhenFocusedNotClipboardTarget()(ctx) {
		t.Fatalf("expected clipboard target to be false")
	}

	if !WhenFocusedRole(accessibility.RoleButton)(ctx) {
		t.Fatalf("expected role match")
	}
	if WhenFocusedRole(accessibility.RoleTextbox)(ctx) {
		t.Fatalf("expected role mismatch")
	}

	if WhenFocusedWidget(nil)(ctx) {
		t.Fatalf("expected false with nil predicate")
	}
	if !WhenFocusedWidget(func(w runtime.Widget) bool { return w == clip })(ctx) {
		t.Fatalf("expected focused widget predicate to match")
	}
	if WhenFocusedAccessible(nil)(ctx) {
		t.Fatalf("expected false with nil accessible predicate")
	}
	if !WhenFocusedAccessible(func(a accessibility.Accessible) bool { return a == clip })(ctx) {
		t.Fatalf("expected focused accessible predicate to match")
	}

	all := All(WhenFocused(), nil)
	if !all(ctx) {
		t.Fatalf("expected All to pass with focused")
	}
	any := Any(nil, WhenFocused())
	if !any(ctx) {
		t.Fatalf("expected Any to pass with focused")
	}
	if !Not(nil)(ctx) {
		t.Fatalf("expected Not(nil) to return true")
	}
	if Not(WhenFocused())(ctx) {
		t.Fatalf("expected Not to invert condition")
	}
}

func TestKeyFormattingAndParsing(t *testing.T) {
	key, err := ParseKeySequence("ctrl+c")
	if err != nil {
		t.Fatalf("ParseKeySequence: %v", err)
	}
	if got := FormatKeySequence(key); got != "Ctrl+C" {
		t.Fatalf("FormatKeySequence = %q", got)
	}
	if got := key.String(); got != "Ctrl+C" {
		t.Fatalf("Key.String = %q", got)
	}
	if got := key.Sequence[0].String(); got != "Ctrl+C" {
		t.Fatalf("KeyPress.String = %q", got)
	}
	if !key.Matches(key.Sequence) {
		t.Fatalf("expected Matches to succeed")
	}

	press := KeyPressFromKeyMsg(runtime.KeyMsg{Key: terminal.KeyCtrlC})
	if !press.Ctrl {
		t.Fatalf("expected ctrl key to set Ctrl flag")
	}
	termPress := KeyPressFromTerminal(terminal.KeyEvent{Key: terminal.KeyRune, Rune: 'g'})
	if termPress.Key != terminal.KeyRune || termPress.Rune != 'g' {
		t.Fatalf("unexpected terminal key press")
	}

	formatted := FormatKeySequences([]Key{key})
	if !strings.Contains(formatted, "Ctrl+C") {
		t.Fatalf("unexpected formatted sequence: %q", formatted)
	}
}

func TestKeymapAndRouter(t *testing.T) {
	registry := NewRegistry()
	called := false
	registry.Register(Command{ID: "action", Handler: func(Context) { called = true }})

	binding := Binding{Key: MustParseKeySequence("g g"), Command: "action"}
	keymap := &Keymap{Name: "test", Bindings: []Binding{binding}}

	router := NewKeyRouter(registry, nil, nil)
	ctx := Context{Keymap: keymap}

	if !router.HandleKey(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'g'}, ctx) {
		t.Fatalf("expected prefix match")
	}
	if called {
		t.Fatalf("did not expect command yet")
	}
	if !router.HandleKey(runtime.KeyMsg{Key: terminal.KeyRune, Rune: 'g'}, ctx) {
		t.Fatalf("expected command execution")
	}
	if !called {
		t.Fatalf("expected command to run")
	}

	router.Reset()
}

func TestRegistryAndShortcuts(t *testing.T) {
	registry := NewRegistry()
	registry.Register(Command{ID: "one", Handler: func(Context) {}})
	registry.RegisterAll(Command{ID: "two", Handler: func(Context) {}})

	if _, ok := registry.Get("one"); !ok {
		t.Fatalf("expected command one")
	}
	if len(registry.List()) < 2 {
		t.Fatalf("expected list of commands")
	}

	called := false
	registry.Register(Command{ID: "exec", Handler: func(Context) { called = true }})
	if !registry.Execute("exec", Context{}) {
		t.Fatalf("expected execute to succeed")
	}
	if !called {
		t.Fatalf("expected handler to run")
	}

	shortcuts := CommandShortcuts(DefaultKeymap())
	if len(shortcuts) == 0 {
		t.Fatalf("expected shortcuts")
	}
}

func TestModeManagerAndStack(t *testing.T) {
	manager := NewModeManager()
	keymap := &Keymap{Name: "primary"}
	manager.Register("main", keymap)

	if manager.Current() != keymap {
		t.Fatalf("expected current keymap")
	}
	if manager.CurrentName() != "main" {
		t.Fatalf("expected current name")
	}
	manager.Push("alternate")
	if manager.CurrentName() != "alternate" {
		t.Fatalf("expected push to update current")
	}
	manager.Pop()
	if manager.CurrentName() != "main" {
		t.Fatalf("expected pop to restore")
	}
	manager.Set("main")

	stack := &KeymapStack{}
	stack.Push(keymap)
	if stack.Current() != keymap {
		t.Fatalf("expected current keymap")
	}
	if len(stack.All()) != 1 {
		t.Fatalf("expected stack copy")
	}
	if stack.Pop() != keymap {
		t.Fatalf("expected pop")
	}
}

func TestRuntimeHandler(t *testing.T) {
	registry := NewRegistry()
	called := false
	registry.Register(Command{ID: "run", Handler: func(Context) { called = true }})

	keymap := &Keymap{Bindings: []Binding{{Key: MustParseKeySequence("enter"), Command: "run"}}}
	focused := &handlerWidget{}
	focused.HandlerBase.Map = keymap
	focused.role = accessibility.RoleButton

	router := NewKeyRouter(registry, nil, nil)
	handler := &RuntimeHandler{Router: router}
	app := runtime.NewApp(runtime.AppConfig{})

	handled := handler.HandleKey(app, runtime.KeyMsg{Key: terminal.KeyEnter}, focused)
	if !handled || !called {
		t.Fatalf("expected runtime handler to execute")
	}
}

func TestStandardScrollAndClipboardCommands(t *testing.T) {
	registry := NewRegistry()
	RegisterStandardCommands(registry)
	RegisterScrollCommands(registry)
	RegisterClipboardCommands(registry)

	app := runtime.NewApp(runtime.AppConfig{Clipboard: &clipboard.MemoryClipboard{}})

	// Standard commands execute without errors.
	_ = registry.Execute("app.quit", Context{App: app})
	_ = registry.Execute("app.refresh", Context{App: app})

	// Focus commands route to app handler.
	called := false
	appWithHandler := runtime.NewApp(runtime.AppConfig{CommandHandler: func(cmd runtime.Command) bool {
		called = true
		return true
	}})
	_ = registry.Execute("focus.next", Context{App: appWithHandler})
	if !called {
		t.Fatalf("expected command handler to run")
	}

	// Scroll commands invoke controller methods.
	scroller := &scrollWidget{}
	ctx := Context{Focused: scroller, App: app}
	_ = registry.Execute("scroll.up", ctx)
	_ = registry.Execute("scroll.pageDown", ctx)
	_ = registry.Execute("scroll.home", ctx)
	_ = registry.Execute("scroll.end", ctx)
	if scroller.scrollDy == 0 || scroller.pageBy == 0 || !scroller.atStart || !scroller.atEnd {
		t.Fatalf("expected scroll commands to update state")
	}

	// Clipboard commands update clipboard and target.
	clip := &clipboardWidget{}
	ctx = Context{Focused: clip, App: app}
	_ = registry.Execute(clipboard.CommandCopy, ctx)
	_ = registry.Execute(clipboard.CommandCut, ctx)
	_ = registry.Execute(clipboard.CommandPaste, ctx)
	if len(clip.pasted) == 0 {
		t.Fatalf("expected paste to be called")
	}
	if val, _ := app.Services().Clipboard().Read(); val == "" {
		t.Fatalf("expected clipboard to be populated")
	}
}

var _ runtime.Widget = (*testWidget)(nil)
var _ accessibility.Accessible = (*testWidget)(nil)
var _ runtime.Widget = (*clipboardWidget)(nil)
var _ clipboard.Target = (*clipboardWidget)(nil)
var _ runtime.Widget = (*scrollWidget)(nil)
var _ scroll.Controller = (*scrollWidget)(nil)
var _ runtime.Widget = (*handlerWidget)(nil)
var _ Handler = (*handlerWidget)(nil)
var _ accessibility.Accessible = (*handlerWidget)(nil)
