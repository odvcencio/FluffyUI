package mcp

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"
	"testing"
	"time"

	mcpclient "github.com/odvcencio/fluffyui/third_party/mcp-go/client"
	mcptr "github.com/odvcencio/fluffyui/third_party/mcp-go/client/transport"
	mcp "github.com/odvcencio/fluffyui/third_party/mcp-go/mcp"
)

type mockTransport struct {
	mu                  sync.Mutex
	lastRequest         mcptr.JSONRPCRequest
	notificationHandler func(mcp.JSONRPCNotification)
	toolErrors          map[string]string
	toolIsError         map[string]bool
	resourceResults     map[string]mcp.ReadResourceResult
}

func (m *mockTransport) Start(context.Context) error { return nil }

func (m *mockTransport) SendRequest(_ context.Context, request mcptr.JSONRPCRequest) (*mcptr.JSONRPCResponse, error) {
	m.mu.Lock()
	m.lastRequest = request
	m.mu.Unlock()

	switch request.Method {
	case "tools/call":
		params, _ := request.Params.(mcp.CallToolParams)
		name := params.Name
		envelope := map[string]any{
			"_schema": SchemaVersion,
			"_tool":   name,
		}
		if data, ok := toolDataFor(name); ok {
			envelope["data"] = data
		}
		if errMsg, ok := m.toolErrors[name]; ok {
			envelope["error"] = errMsg
		}
		result := mcp.CallToolResult{
			Content:           []mcp.Content{mcp.TextContent{Type: "text", Text: "{}"}},
			StructuredContent: envelope,
		}
		if m.toolIsError[name] {
			result.IsError = true
		}
		raw, _ := json.Marshal(result)
		return &mcptr.JSONRPCResponse{JSONRPC: mcp.JSONRPC_VERSION, ID: request.ID, Result: raw}, nil
	case "resources/read":
		params, _ := request.Params.(mcp.ReadResourceParams)
		res, ok := m.resourceResults[params.URI]
		if !ok || len(res.Contents) == 0 {
			res = mcp.ReadResourceResult{Contents: []mcp.ResourceContents{
				mcp.TextResourceContents{URI: params.URI, MIMEType: "text/plain", Text: "default"},
			}}
		}
		raw, _ := json.Marshal(res)
		return &mcptr.JSONRPCResponse{JSONRPC: mcp.JSONRPC_VERSION, ID: request.ID, Result: raw}, nil
	default:
		raw, _ := json.Marshal(map[string]any{})
		return &mcptr.JSONRPCResponse{JSONRPC: mcp.JSONRPC_VERSION, ID: request.ID, Result: raw}, nil
	}
}

func (m *mockTransport) SendNotification(_ context.Context, notification mcp.JSONRPCNotification) error {
	if m.notificationHandler != nil {
		m.notificationHandler(notification)
	}
	return nil
}

func (m *mockTransport) SetNotificationHandler(handler func(mcp.JSONRPCNotification)) {
	m.notificationHandler = handler
}

func (m *mockTransport) Close() error { return nil }
func (m *mockTransport) GetSessionId() string {
	return "mock"
}

func newMockClient() (*Client, *mockTransport) {
	transport := &mockTransport{}
	inner := mcpclient.NewClient(transport, mcpclient.WithSession())
	client := &Client{inner: inner, timeout: 25 * time.Millisecond}
	inner.OnNotification(client.handleNotification)
	return client, transport
}

func toolDataFor(name string) (any, bool) {
	switch name {
	case "snapshot":
		return Snapshot{Dimensions: Dimensions{Width: 80, Height: 24}, LayerCount: 1}, true
	case "snapshot_text":
		return "snapshot", true
	case "snapshot_region":
		return "region", true
	case "get_dimensions":
		return Dimensions{Width: 100, Height: 40}, true
	case "get_layer_count":
		return 2, true
	case "get_cell":
		return CellInfo{Char: "X"}, true
	case "get_bounds":
		return Rect{X: 1, Y: 2, Width: 3, Height: 4}, true
	case "get_state":
		return StateSet{Focused: true}, true
	case "get_actions":
		return []string{"activate"}, true
	case "get_label":
		return "Label", true
	case "get_role":
		return "button", true
	case "get_value":
		return "42", true
	case "get_description":
		return "desc", true
	case "get_selection":
		return "selection", true
	case "get_selection_bounds":
		return SelectionBounds{Start: 1, End: 3}, true
	case "get_cursor_position":
		return CursorPosition{X: 2, Y: 3}, true
	case "get_cursor_offset":
		return 7, true
	case "get_capabilities":
		return Capabilities{AllowText: true, Transport: "mock"}, true
	case "get_app_info":
		return AppInfo{Width: 80, Height: 24, LayerCount: 1}, true
	case "diff_snapshots":
		return Diff{TextChanged: true}, true
	case "find_by_label", "find_by_id", "find_at_position", "find_focused", "get_parent", "get_next_focusable", "get_prev_focusable":
		return &WidgetInfo{ID: "w1", Label: "Widget"}, true
	case "find_by_role", "find_by_value", "find_by_state", "find_all", "find_focusable", "find_actionable", "get_children", "get_siblings", "get_descendants", "get_ancestors":
		return []WidgetInfo{{ID: "w2"}}, true
	case "activate", "focus", "type_into", "clear", "select_option", "select_index", "toggle", "check", "uncheck", "expand", "collapse", "scroll_to", "scroll_by", "scroll_to_top", "scroll_to_bottom", "click_widget":
		return ActionResult{Status: "ok"}, true
	case "clipboard_read":
		return "clip", true
	case "clipboard_read_primary":
		return "clip-primary", true
	}

	if isBoolTool(name) {
		return true, true
	}

	return nil, false
}

func isBoolTool(name string) bool {
	switch name {
	case "is_focused", "is_enabled", "is_visible", "is_checked", "is_expanded", "is_selected", "has_focus",
		"blur", "press_key", "press_keys", "press_chord", "press_rune", "type_string", "press_enter", "press_escape", "press_tab",
		"press_shift_tab", "press_space", "press_backspace", "press_delete", "press_up", "press_down", "press_left", "press_right",
		"press_home", "press_end", "press_page_up", "press_page_down", "press_f1", "press_f2", "press_f3", "press_f4",
		"press_f5", "press_f6", "press_f7", "press_f8", "press_f9", "press_f10", "press_f11", "press_f12",
		"mouse_click", "mouse_double_click", "mouse_right_click", "mouse_press", "mouse_release", "mouse_move", "mouse_drag",
		"mouse_scroll_up", "mouse_scroll_down", "clipboard_write", "clipboard_clear", "clipboard_has_text", "clipboard_write_primary",
		"select_all", "select_range", "select_word", "select_line", "select_none", "has_selection", "copy", "cut", "paste",
		"paste_text", "set_cursor_position", "set_cursor_offset", "cursor_to_start", "cursor_to_end", "cursor_word_left",
		"cursor_word_right", "tick", "wait_ms", "wait_for_widget", "wait_for_widget_gone", "wait_for_text", "wait_for_text_gone",
		"wait_for_focus", "wait_for_value", "wait_for_enabled", "wait_for_idle", "resize", "resize_width", "resize_height",
		"widgets_changed", "text_changed", "ping":
		return true
	}
	return false
}

func TestClientToolHelpers(t *testing.T) {
	client, _ := newMockClient()

	snap, err := client.Snapshot()
	if err != nil {
		t.Fatalf("snapshot error: %v", err)
	}
	if snap.Dimensions.Width != 80 {
		t.Fatalf("unexpected snapshot dimensions")
	}

	if _, err := client.SnapshotWithText(); err != nil {
		t.Fatalf("snapshot with text error: %v", err)
	}
	if _, err := client.SnapshotText(); err != nil {
		t.Fatalf("snapshot text error: %v", err)
	}
	if _, err := client.SnapshotRegion(0, 0, 2, 2); err != nil {
		t.Fatalf("snapshot region error: %v", err)
	}

	dims, err := client.GetDimensions()
	if err != nil || dims.Width != 100 {
		t.Fatalf("dimensions error: %v", err)
	}
	if _, err := client.GetLayerCount(); err != nil {
		t.Fatalf("layer count error: %v", err)
	}
	if _, err := client.GetCell(1, 2); err != nil {
		t.Fatalf("cell error: %v", err)
	}

	if _, err := client.FindByLabel("Label"); err != nil {
		t.Fatalf("find by label error: %v", err)
	}
	if _, err := client.FindByLabelAt("Label", 1); err != nil {
		t.Fatalf("find by label at error: %v", err)
	}
	if _, err := client.FindByLabelInLayer("Label", 0); err != nil {
		t.Fatalf("find by label in layer error: %v", err)
	}
	if _, err := client.FindByLabelAtLayer("Label", 1, 0); err != nil {
		t.Fatalf("find by label at layer error: %v", err)
	}
	if _, err := client.FindByRole("button"); err != nil {
		t.Fatalf("find by role error: %v", err)
	}
	if _, err := client.FindByID("id"); err != nil {
		t.Fatalf("find by id error: %v", err)
	}
	if _, err := client.FindByValue("value"); err != nil {
		t.Fatalf("find by value error: %v", err)
	}
	if _, err := client.FindByState(StateSet{}); err != nil {
		t.Fatalf("find by state error: %v", err)
	}
	if _, err := client.FindAtPosition(1, 1); err != nil {
		t.Fatalf("find at position error: %v", err)
	}
	if _, err := client.FindFocused(); err != nil {
		t.Fatalf("find focused error: %v", err)
	}
	if _, err := client.FindAll(); err != nil {
		t.Fatalf("find all error: %v", err)
	}
	if _, err := client.FindFocusable(); err != nil {
		t.Fatalf("find focusable error: %v", err)
	}
	if _, err := client.FindActionable(); err != nil {
		t.Fatalf("find actionable error: %v", err)
	}

	if _, err := client.GetChildren("id"); err != nil {
		t.Fatalf("get children error: %v", err)
	}
	if _, err := client.GetParent("id"); err != nil {
		t.Fatalf("get parent error: %v", err)
	}
	if _, err := client.GetSiblings("id"); err != nil {
		t.Fatalf("get siblings error: %v", err)
	}
	if _, err := client.GetDescendants("id"); err != nil {
		t.Fatalf("get descendants error: %v", err)
	}
	if _, err := client.GetAncestors("id"); err != nil {
		t.Fatalf("get ancestors error: %v", err)
	}
	if _, err := client.NextFocusable(); err != nil {
		t.Fatalf("next focusable error: %v", err)
	}
	if _, err := client.NextFocusableFrom("id"); err != nil {
		t.Fatalf("next focusable from error: %v", err)
	}
	if _, err := client.PrevFocusable(); err != nil {
		t.Fatalf("prev focusable error: %v", err)
	}
	if _, err := client.PrevFocusableFrom("id"); err != nil {
		t.Fatalf("prev focusable from error: %v", err)
	}

	if _, err := client.GetLabel("id"); err != nil {
		t.Fatalf("get label error: %v", err)
	}
	if _, err := client.GetRole("id"); err != nil {
		t.Fatalf("get role error: %v", err)
	}
	if _, err := client.GetValue("id"); err != nil {
		t.Fatalf("get value error: %v", err)
	}
	if _, err := client.GetDescription("id"); err != nil {
		t.Fatalf("get description error: %v", err)
	}
	if _, err := client.GetBounds("id"); err != nil {
		t.Fatalf("get bounds error: %v", err)
	}
	if _, err := client.GetState("id"); err != nil {
		t.Fatalf("get state error: %v", err)
	}
	if _, err := client.GetActions("id"); err != nil {
		t.Fatalf("get actions error: %v", err)
	}
	if _, err := client.IsFocused("id"); err != nil {
		t.Fatalf("is focused error: %v", err)
	}
	if _, err := client.IsEnabled("id"); err != nil {
		t.Fatalf("is enabled error: %v", err)
	}
	if _, err := client.IsVisible("id"); err != nil {
		t.Fatalf("is visible error: %v", err)
	}
	if _, err := client.IsChecked("id"); err != nil {
		t.Fatalf("is checked error: %v", err)
	}
	if _, err := client.IsExpanded("id"); err != nil {
		t.Fatalf("is expanded error: %v", err)
	}
	if _, err := client.IsSelected("id"); err != nil {
		t.Fatalf("is selected error: %v", err)
	}
	if _, err := client.HasFocus(); err != nil {
		t.Fatalf("has focus error: %v", err)
	}

	if _, err := client.Activate("label"); err != nil {
		t.Fatalf("activate error: %v", err)
	}
	if _, err := client.ActivateAt("label", 1); err != nil {
		t.Fatalf("activate at error: %v", err)
	}
	if _, err := client.ActivateLayer("label", 0); err != nil {
		t.Fatalf("activate layer error: %v", err)
	}
	if _, err := client.ActivateID("id"); err != nil {
		t.Fatalf("activate id error: %v", err)
	}
	if _, err := client.Focus("label"); err != nil {
		t.Fatalf("focus error: %v", err)
	}
	if _, err := client.FocusAt("label", 1); err != nil {
		t.Fatalf("focus at error: %v", err)
	}
	if _, err := client.FocusLayer("label", 0); err != nil {
		t.Fatalf("focus layer error: %v", err)
	}
	if _, err := client.FocusID("id"); err != nil {
		t.Fatalf("focus id error: %v", err)
	}
	if _, err := client.Blur(); err != nil {
		t.Fatalf("blur error: %v", err)
	}
	if _, err := client.TypeInto("label", "text"); err != nil {
		t.Fatalf("type into error: %v", err)
	}
	if _, err := client.TypeIntoAt("label", 1, "text"); err != nil {
		t.Fatalf("type into at error: %v", err)
	}
	if _, err := client.TypeIntoLayer("label", 0, "text"); err != nil {
		t.Fatalf("type into layer error: %v", err)
	}
	if _, err := client.TypeIntoID("id", "text"); err != nil {
		t.Fatalf("type into id error: %v", err)
	}
	if _, err := client.Clear("label"); err != nil {
		t.Fatalf("clear error: %v", err)
	}
	if _, err := client.ClearID("id"); err != nil {
		t.Fatalf("clear id error: %v", err)
	}
	if _, err := client.SelectOption("label", "option"); err != nil {
		t.Fatalf("select option error: %v", err)
	}
	if _, err := client.SelectOptionID("id", "option"); err != nil {
		t.Fatalf("select option id error: %v", err)
	}
	if _, err := client.SelectIndex("label", 1); err != nil {
		t.Fatalf("select index error: %v", err)
	}
	if _, err := client.SelectIndexID("id", 1); err != nil {
		t.Fatalf("select index id error: %v", err)
	}
	if _, err := client.Toggle("label"); err != nil {
		t.Fatalf("toggle error: %v", err)
	}
	if _, err := client.ToggleID("id"); err != nil {
		t.Fatalf("toggle id error: %v", err)
	}
	if _, err := client.Check("label"); err != nil {
		t.Fatalf("check error: %v", err)
	}
	if _, err := client.CheckID("id"); err != nil {
		t.Fatalf("check id error: %v", err)
	}
	if _, err := client.Uncheck("label"); err != nil {
		t.Fatalf("uncheck error: %v", err)
	}
	if _, err := client.UncheckID("id"); err != nil {
		t.Fatalf("uncheck id error: %v", err)
	}
	if _, err := client.Expand("label"); err != nil {
		t.Fatalf("expand error: %v", err)
	}
	if _, err := client.ExpandID("id"); err != nil {
		t.Fatalf("expand id error: %v", err)
	}
	if _, err := client.Collapse("label"); err != nil {
		t.Fatalf("collapse error: %v", err)
	}
	if _, err := client.CollapseID("id"); err != nil {
		t.Fatalf("collapse id error: %v", err)
	}
	if _, err := client.ScrollTo("label"); err != nil {
		t.Fatalf("scroll to error: %v", err)
	}
	if _, err := client.ScrollToID("id"); err != nil {
		t.Fatalf("scroll to id error: %v", err)
	}
	if _, err := client.ScrollBy("label", 1); err != nil {
		t.Fatalf("scroll by error: %v", err)
	}
	if _, err := client.ScrollByID("id", 1); err != nil {
		t.Fatalf("scroll by id error: %v", err)
	}
	if _, err := client.ScrollToTop("label"); err != nil {
		t.Fatalf("scroll to top error: %v", err)
	}
	if _, err := client.ScrollToTopID("id"); err != nil {
		t.Fatalf("scroll to top id error: %v", err)
	}
	if _, err := client.ScrollToBottom("label"); err != nil {
		t.Fatalf("scroll to bottom error: %v", err)
	}
	if _, err := client.ScrollToBottomID("id"); err != nil {
		t.Fatalf("scroll to bottom id error: %v", err)
	}

	if _, err := client.PressKey("k"); err != nil {
		t.Fatalf("press key error: %v", err)
	}
	if _, err := client.PressKeys("ctrl+k"); err != nil {
		t.Fatalf("press keys error: %v", err)
	}
	if _, err := client.PressChord("ctrl+shift+k"); err != nil {
		t.Fatalf("press chord error: %v", err)
	}
	if _, err := client.PressRune('x'); err != nil {
		t.Fatalf("press rune error: %v", err)
	}
	if _, err := client.TypeString("text"); err != nil {
		t.Fatalf("type string error: %v", err)
	}
	if _, err := client.PressEnter(); err != nil {
		t.Fatalf("press enter error: %v", err)
	}
	if _, err := client.PressEscape(); err != nil {
		t.Fatalf("press escape error: %v", err)
	}
	if _, err := client.PressTab(); err != nil {
		t.Fatalf("press tab error: %v", err)
	}
	if _, err := client.PressShiftTab(); err != nil {
		t.Fatalf("press shift tab error: %v", err)
	}
	if _, err := client.PressSpace(); err != nil {
		t.Fatalf("press space error: %v", err)
	}
	if _, err := client.PressBackspace(); err != nil {
		t.Fatalf("press backspace error: %v", err)
	}
	if _, err := client.PressDelete(); err != nil {
		t.Fatalf("press delete error: %v", err)
	}
	if _, err := client.PressUp(); err != nil {
		t.Fatalf("press up error: %v", err)
	}
	if _, err := client.PressDown(); err != nil {
		t.Fatalf("press down error: %v", err)
	}
	if _, err := client.PressLeft(); err != nil {
		t.Fatalf("press left error: %v", err)
	}
	if _, err := client.PressRight(); err != nil {
		t.Fatalf("press right error: %v", err)
	}
	if _, err := client.PressHome(); err != nil {
		t.Fatalf("press home error: %v", err)
	}
	if _, err := client.PressEnd(); err != nil {
		t.Fatalf("press end error: %v", err)
	}
	if _, err := client.PressPageUp(); err != nil {
		t.Fatalf("press page up error: %v", err)
	}
	if _, err := client.PressPageDown(); err != nil {
		t.Fatalf("press page down error: %v", err)
	}
	if _, err := client.PressF1(); err != nil {
		t.Fatalf("press f1 error: %v", err)
	}
	if _, err := client.PressF2(); err != nil {
		t.Fatalf("press f2 error: %v", err)
	}
	if _, err := client.PressF3(); err != nil {
		t.Fatalf("press f3 error: %v", err)
	}
	if _, err := client.PressF4(); err != nil {
		t.Fatalf("press f4 error: %v", err)
	}
	if _, err := client.PressF5(); err != nil {
		t.Fatalf("press f5 error: %v", err)
	}
	if _, err := client.PressF6(); err != nil {
		t.Fatalf("press f6 error: %v", err)
	}
	if _, err := client.PressF7(); err != nil {
		t.Fatalf("press f7 error: %v", err)
	}
	if _, err := client.PressF8(); err != nil {
		t.Fatalf("press f8 error: %v", err)
	}
	if _, err := client.PressF9(); err != nil {
		t.Fatalf("press f9 error: %v", err)
	}
	if _, err := client.PressF10(); err != nil {
		t.Fatalf("press f10 error: %v", err)
	}
	if _, err := client.PressF11(); err != nil {
		t.Fatalf("press f11 error: %v", err)
	}
	if _, err := client.PressF12(); err != nil {
		t.Fatalf("press f12 error: %v", err)
	}

	if _, err := client.MouseClick(1, 1); err != nil {
		t.Fatalf("mouse click error: %v", err)
	}
	if _, err := client.MouseDoubleClick(1, 1); err != nil {
		t.Fatalf("mouse double click error: %v", err)
	}
	if _, err := client.MouseRightClick(1, 1); err != nil {
		t.Fatalf("mouse right click error: %v", err)
	}
	if _, err := client.MousePress(1, 1, "left"); err != nil {
		t.Fatalf("mouse press error: %v", err)
	}
	if _, err := client.MouseRelease(1, 1, "left"); err != nil {
		t.Fatalf("mouse release error: %v", err)
	}
	if _, err := client.MouseMove(1, 1); err != nil {
		t.Fatalf("mouse move error: %v", err)
	}
	if _, err := client.MouseDrag(0, 0, 1, 1); err != nil {
		t.Fatalf("mouse drag error: %v", err)
	}
	if _, err := client.MouseScrollUp(1, 1, 1); err != nil {
		t.Fatalf("mouse scroll up error: %v", err)
	}
	if _, err := client.MouseScrollDown(1, 1, 1); err != nil {
		t.Fatalf("mouse scroll down error: %v", err)
	}
	if _, err := client.ClickWidget("label"); err != nil {
		t.Fatalf("click widget error: %v", err)
	}

	if _, err := client.ClipboardRead(); err != nil {
		t.Fatalf("clipboard read error: %v", err)
	}
	if _, err := client.ClipboardWrite("text"); err != nil {
		t.Fatalf("clipboard write error: %v", err)
	}
	if _, err := client.ClipboardClear(); err != nil {
		t.Fatalf("clipboard clear error: %v", err)
	}
	if _, err := client.ClipboardHasText(); err != nil {
		t.Fatalf("clipboard has text error: %v", err)
	}
	if _, err := client.ClipboardReadPrimary(); err != nil {
		t.Fatalf("clipboard read primary error: %v", err)
	}
	if _, err := client.ClipboardWritePrimary("text"); err != nil {
		t.Fatalf("clipboard write primary error: %v", err)
	}

	if _, err := client.SelectAll(); err != nil {
		t.Fatalf("select all error: %v", err)
	}
	if _, err := client.SelectRange(0, 1); err != nil {
		t.Fatalf("select range error: %v", err)
	}
	if _, err := client.SelectWord(); err != nil {
		t.Fatalf("select word error: %v", err)
	}
	if _, err := client.SelectLine(); err != nil {
		t.Fatalf("select line error: %v", err)
	}
	if _, err := client.SelectNone(); err != nil {
		t.Fatalf("select none error: %v", err)
	}
	if _, err := client.GetSelection(); err != nil {
		t.Fatalf("get selection error: %v", err)
	}
	if _, err := client.GetSelectionBounds(); err != nil {
		t.Fatalf("get selection bounds error: %v", err)
	}
	if _, err := client.HasSelection(); err != nil {
		t.Fatalf("has selection error: %v", err)
	}
	if _, err := client.Copy(); err != nil {
		t.Fatalf("copy error: %v", err)
	}
	if _, err := client.Cut(); err != nil {
		t.Fatalf("cut error: %v", err)
	}
	if _, err := client.Paste(); err != nil {
		t.Fatalf("paste error: %v", err)
	}
	if _, err := client.PasteText("text"); err != nil {
		t.Fatalf("paste text error: %v", err)
	}

	if _, err := client.GetCursorPosition(); err != nil {
		t.Fatalf("get cursor position error: %v", err)
	}
	if _, err := client.SetCursorPosition(1, 1); err != nil {
		t.Fatalf("set cursor position error: %v", err)
	}
	if _, err := client.GetCursorOffset(); err != nil {
		t.Fatalf("get cursor offset error: %v", err)
	}
	if _, err := client.SetCursorOffset(2); err != nil {
		t.Fatalf("set cursor offset error: %v", err)
	}
	if _, err := client.CursorToStart(); err != nil {
		t.Fatalf("cursor to start error: %v", err)
	}
	if _, err := client.CursorToEnd(); err != nil {
		t.Fatalf("cursor to end error: %v", err)
	}
	if _, err := client.CursorWordLeft(); err != nil {
		t.Fatalf("cursor word left error: %v", err)
	}
	if _, err := client.CursorWordRight(); err != nil {
		t.Fatalf("cursor word right error: %v", err)
	}

	if _, err := client.Tick(); err != nil {
		t.Fatalf("tick error: %v", err)
	}
	if _, err := client.WaitMS(1); err != nil {
		t.Fatalf("wait ms error: %v", err)
	}
	if _, err := client.WaitForWidget("label", time.Millisecond); err != nil {
		t.Fatalf("wait for widget error: %v", err)
	}
	if _, err := client.WaitForWidgetGone("label", time.Millisecond); err != nil {
		t.Fatalf("wait for widget gone error: %v", err)
	}
	if _, err := client.WaitForText("text", time.Millisecond); err != nil {
		t.Fatalf("wait for text error: %v", err)
	}
	if _, err := client.WaitForTextGone("text", time.Millisecond); err != nil {
		t.Fatalf("wait for text gone error: %v", err)
	}
	if _, err := client.WaitForFocus("label", time.Millisecond); err != nil {
		t.Fatalf("wait for focus error: %v", err)
	}
	if _, err := client.WaitForValue("label", "value", time.Millisecond); err != nil {
		t.Fatalf("wait for value error: %v", err)
	}
	if _, err := client.WaitForEnabled("label", time.Millisecond); err != nil {
		t.Fatalf("wait for enabled error: %v", err)
	}
	if _, err := client.WaitForIdle(time.Millisecond); err != nil {
		t.Fatalf("wait for idle error: %v", err)
	}

	if _, err := client.Resize(80, 24); err != nil {
		t.Fatalf("resize error: %v", err)
	}
	if _, err := client.ResizeWidth(80); err != nil {
		t.Fatalf("resize width error: %v", err)
	}
	if _, err := client.ResizeHeight(24); err != nil {
		t.Fatalf("resize height error: %v", err)
	}

	if _, err := client.DiffSnapshots(Snapshot{}, Snapshot{}); err != nil {
		t.Fatalf("diff snapshots error: %v", err)
	}
	if _, err := client.WidgetsChanged(Snapshot{}); err != nil {
		t.Fatalf("widgets changed error: %v", err)
	}
	if _, err := client.TextChanged(Snapshot{}); err != nil {
		t.Fatalf("text changed error: %v", err)
	}

	if _, err := client.GetCapabilities(); err != nil {
		t.Fatalf("get capabilities error: %v", err)
	}
	if _, err := client.GetAppInfo(); err != nil {
		t.Fatalf("get app info error: %v", err)
	}
	if _, err := client.Ping(); err != nil {
		t.Fatalf("ping error: %v", err)
	}
}

func TestClientToolErrors(t *testing.T) {
	client, transport := newMockClient()
	transport.toolErrors = map[string]string{"ping": "boom"}
	if _, err := client.Ping(); err == nil {
		t.Fatalf("expected tool error")
	}

	transport.toolErrors = nil
	transport.toolIsError = map[string]bool{"ping": true}
	if _, err := client.Ping(); err == nil {
		t.Fatalf("expected tool isError")
	}
}

func TestClientReadResourceValue(t *testing.T) {
	client, transport := newMockClient()
	transport.resourceResults = map[string]mcp.ReadResourceResult{
		"fluffy://text": {
			Contents: []mcp.ResourceContents{mcp.TextResourceContents{URI: "fluffy://text", MIMEType: "text/plain", Text: "hello"}},
		},
		"fluffy://json": {
			Contents: []mcp.ResourceContents{mcp.TextResourceContents{URI: "fluffy://json", MIMEType: "application/json", Text: "{\"id\":\"w1\"}"}},
		},
	}

	val, err := client.readResourceValue("fluffy://text", reflectTypeString())
	if err != nil || val.String() != "hello" {
		t.Fatalf("unexpected resource text: %v", err)
	}

	val, err = client.readResourceValue("fluffy://json", reflectTypeWidgetPtr())
	if err != nil {
		t.Fatalf("unexpected resource json error: %v", err)
	}
	info, ok := val.Interface().(*WidgetInfo)
	if !ok || info == nil || info.ID != "w1" {
		t.Fatalf("unexpected resource json value")
	}
}

func reflectTypeString() reflect.Type {
	return reflect.TypeOf("")
}

func reflectTypeWidgetPtr() reflect.Type {
	return reflect.TypeOf(&WidgetInfo{})
}
