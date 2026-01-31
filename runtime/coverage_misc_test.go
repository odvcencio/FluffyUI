package runtime

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
)

type boundsWidget struct {
	bounds    Rect
	children  []Widget
	focusable bool
	focused   bool
}

func (b *boundsWidget) Measure(constraints Constraints) Size {
	return constraints.MaxSize()
}

func (b *boundsWidget) Layout(bounds Rect) {
	b.bounds = bounds
	for _, child := range b.children {
		if child != nil {
			child.Layout(bounds)
		}
	}
}

func (b *boundsWidget) Render(ctx RenderContext) {}

func (b *boundsWidget) HandleMessage(msg Message) HandleResult { return Unhandled() }

func (b *boundsWidget) Bounds() Rect { return b.bounds }

func (b *boundsWidget) ChildWidgets() []Widget { return b.children }

func (b *boundsWidget) HitSelf() bool { return true }

func (b *boundsWidget) CanFocus() bool { return b.focusable }

func (b *boundsWidget) Focus() { b.focused = true }

func (b *boundsWidget) Blur() { b.focused = false }

func (b *boundsWidget) IsFocused() bool { return b.focused }

type persistWidget struct {
	key      string
	state    string
	children []Widget
}

func (p *persistWidget) Measure(constraints Constraints) Size { return Size{} }

func (p *persistWidget) Layout(bounds Rect) {}

func (p *persistWidget) Render(ctx RenderContext) {}

func (p *persistWidget) HandleMessage(msg Message) HandleResult { return Unhandled() }

func (p *persistWidget) ChildWidgets() []Widget { return p.children }

func (p *persistWidget) Key() string { return p.key }

func (p *persistWidget) MarshalState() ([]byte, error) {
	return json.Marshal(p.state)
}

func (p *persistWidget) UnmarshalState(data []byte) error {
	return json.Unmarshal(data, &p.state)
}

func TestScreenHitGridAndFocus(t *testing.T) {
	child := &boundsWidget{focusable: true}
	root := &boundsWidget{children: []Widget{child}}
	announcer := &accessibility.SimpleAnnouncer{}
	focusStyle := &accessibility.FocusStyle{Indicator: ">", Style: backend.DefaultStyle()}

	app := NewApp(AppConfig{Announcer: announcer, FocusStyle: focusStyle})
	screen := NewScreen(10, 5)
	screen.SetServices(app.Services())
	screen.SetRoot(root)
	screen.SetAutoRegisterFocus(true)
	screen.RefreshFocusables()

	if screen.Layer(0) == nil {
		t.Fatalf("expected base layer")
	}
	if screen.BaseLayer() == nil {
		t.Fatalf("expected base layer")
	}
	if screen.BaseFocusScope() == nil {
		t.Fatalf("expected base focus scope")
	}

	if scope := screen.FocusScope(); scope != nil {
		scope.FocusFirst()
		scope.Reset()
	}

	screen.Render()
	if screen.WidgetAt(0, 0) == nil {
		t.Fatalf("expected widget at position")
	}
}

func TestHitGridAndPersist(t *testing.T) {
	grid := NewHitGrid(4, 4)
	widget := &boundsWidget{bounds: Rect{X: 1, Y: 1, Width: 2, Height: 2}}
	grid.Add(widget, widget.bounds)
	if grid.WidgetAt(1, 1) != widget {
		t.Fatalf("expected widget at location")
	}

	child := &persistWidget{key: "child", state: "b"}
	root := &persistWidget{key: "root", state: "a", children: []Widget{child}}
	snap, err := CaptureState(root)
	if err != nil {
		t.Fatalf("capture error: %v", err)
	}
	child.state = ""
	if err := ApplyState(root, snap); err != nil {
		t.Fatalf("apply error: %v", err)
	}
	if child.state != "b" {
		t.Fatalf("expected state restored")
	}

	path := filepath.Join(t.TempDir(), "snap.json")
	if err := SaveSnapshot(path, snap); err != nil {
		t.Fatalf("save snapshot: %v", err)
	}
	if _, err := LoadSnapshot(path); err != nil {
		t.Fatalf("load snapshot: %v", err)
	}
}

func TestBufferExtras(t *testing.T) {
	buf := NewBuffer(4, 4)
	buf.SetContent(1, 1, 'x', nil, backend.DefaultStyle())
	img := backend.Image{Width: 1, Height: 1, CellWidth: 1, CellHeight: 1, Pixels: []byte{0, 0, 0, 0}}
	buf.SetImage(0, 0, img)
	if len(buf.ImageOps()) == 0 {
		t.Fatalf("expected image ops")
	}
	buf.ClearImageOps()
	buf.DrawDoubleBox(Rect{X: 0, Y: 0, Width: 4, Height: 4}, backend.DefaultStyle())
}

func TestMessageMarkers(t *testing.T) {
	KeyMsg{}.isMessage()
	ResizeMsg{}.isMessage()
	MouseMsg{}.isMessage()
	PasteMsg{}.isMessage()
	TickMsg{}.isMessage()
	FocusChangedMsg{}.isMessage()
	QueueFlushMsg{}.isMessage()
	InvalidateMsg{}.isMessage()
	CustomMsg{}.isMessage()
	callMsg{}.isMessage()
}

func TestCommandMarkers(t *testing.T) {
	Quit{}.Command()
	Refresh{}.Command()
	SendMsg{}.Command()
	Submit{}.Command()
	Cancel{}.Command()
	Effect{}.Command()
	FileSelected{}.Command()
	FocusNext{}.Command()
	FocusPrev{}.Command()
	PushOverlay{}.Command()
	PopOverlay{}.Command()
	PaletteSelected{}.Command()
	_ = Send(InvalidateMsg{})
}

func TestMCPEnvParsing(t *testing.T) {
	defer os.Unsetenv("FLUFFY_MCP")
	defer os.Unsetenv("FLUFFY_MCP_ALLOW_TEXT")
	defer os.Unsetenv("FLUFFY_MCP_ALLOW_CLIPBOARD")
	defer os.Unsetenv("FLUFFY_MCP_TOKEN")
	defer os.Unsetenv("FLUFFY_MCP_RATE_LIMIT")
	defer os.Unsetenv("FLUFFY_MCP_BURST_LIMIT")
	defer os.Unsetenv("FLUFFY_MCP_MAX_SESSIONS")
	defer os.Unsetenv("FLUFFY_MCP_MAX_PENDING_EVENTS")
	defer os.Unsetenv("FLUFFY_MCP_SLOW_CLIENT_POLICY")
	defer os.Unsetenv("FLUFFY_MCP_STRICT_LABELS")
	defer os.Unsetenv("FLUFFY_MCP_SESSION_TIMEOUT")
	defer os.Unsetenv("XDG_RUNTIME_DIR")

	os.Setenv("FLUFFY_MCP", "unix:///tmp/fluffy.sock")
	os.Setenv("FLUFFY_MCP_ALLOW_TEXT", "true")
	os.Setenv("FLUFFY_MCP_ALLOW_CLIPBOARD", "true")
	os.Setenv("FLUFFY_MCP_TOKEN", "token")
	os.Setenv("FLUFFY_MCP_RATE_LIMIT", "5")
	os.Setenv("FLUFFY_MCP_BURST_LIMIT", "10")
	os.Setenv("FLUFFY_MCP_MAX_SESSIONS", "2")
	os.Setenv("FLUFFY_MCP_MAX_PENDING_EVENTS", "50")
	os.Setenv("FLUFFY_MCP_SLOW_CLIENT_POLICY", "drop_oldest")
	os.Setenv("FLUFFY_MCP_STRICT_LABELS", "true")
	os.Setenv("FLUFFY_MCP_SESSION_TIMEOUT", "1s")

	opts, ok, err := mcpOptionsFromEnv()
	if err != nil || !ok {
		t.Fatalf("expected options from env")
	}
	if opts.Transport != "unix" {
		t.Fatalf("expected unix transport")
	}

	_, _, _ = envDuration("FLUFFY_MCP_SESSION_TIMEOUT")
	_ = envBool("FLUFFY_MCP_ALLOW_TEXT")
	_ = envInt("FLUFFY_MCP_RATE_LIMIT")
	_, _ = normalizeMCPOptions(MCPOptions{Transport: "stdio"})
	_ = defaultTransport()
	_ = defaultUnixSocketPath()

	// Invalid env value path
	os.Setenv("FLUFFY_MCP", "unix://")
	_, _, _ = mcpOptionsFromEnv()
}

func TestFocusHelpersAndFlex(t *testing.T) {
	child := &boundsWidget{focusable: true}
	root := &boundsWidget{children: []Widget{child}}
	scope := NewFocusScope()
	RegisterFocusables(scope, root)
	if scope.Count() == 0 {
		t.Fatalf("expected focusables registered")
	}
	if scope.FocusFirst() {
		scope.ClearFocus()
	}

	flexChild := FlexChild{Widget: &boundsWidget{focusable: false}}
	flex := &Flex{Children: []FlexChild{flexChild}}
	flex.Layout(Rect{X: 0, Y: 0, Width: 5, Height: 1})
	ctx := RenderContext{Buffer: NewBuffer(5, 1), Bounds: Rect{X: 0, Y: 0, Width: 5, Height: 1}}
	flex.Render(ctx)
	_ = flex.Bounds()
	_ = flex.ChildWidgets()
	_ = flex.PathSegment(flexChild.Widget)
}

func TestDebugHelpers(t *testing.T) {
	buf := &bytes.Buffer{}
	reporter := &ErrorReporter{Writer: buf, RootProvider: func() Widget { return nil }}
	reporter.ReportWidgetError(nil, nil, nil)
	_ = typeName(nil)
	_ = formatMessage(nil)
	_ = formatBox([]string{"a"})
	_ = padRightASCII("a", 3)
	_ = widgetDisplayName(nil)
}

var _ Widget = (*boundsWidget)(nil)
var _ BoundsProvider = (*boundsWidget)(nil)
var _ ChildProvider = (*boundsWidget)(nil)
var _ HitSelfProvider = (*boundsWidget)(nil)
var _ Focusable = (*boundsWidget)(nil)
var _ Persistable = (*persistWidget)(nil)
var _ ChildProvider = (*persistWidget)(nil)
var _ Keyed = (*persistWidget)(nil)
