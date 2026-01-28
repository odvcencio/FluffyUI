package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/agent"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/clipboard"
	"github.com/odvcencio/fluffyui/keybind"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/scroll"
	"github.com/odvcencio/fluffyui/terminal"
	"github.com/odvcencio/fluffyui/widgets"
)

func registerTools(s *Server) {
	addTool(s, mcp.NewTool("snapshot",
		mcp.WithDescription("Capture current UI snapshot."),
		mcp.WithInputSchema[snapshotArgs](),
	), s.handleSnapshot)
	addTool(s, mcp.NewTool("snapshot_text",
		mcp.WithDescription("Capture raw screen text."),
	), s.handleSnapshotText)
	addTool(s, mcp.NewTool("snapshot_region",
		mcp.WithDescription("Capture raw text from a screen region."),
		mcp.WithInputSchema[regionArgs](),
	), s.handleSnapshotRegion)
	addTool(s, mcp.NewTool("get_dimensions",
		mcp.WithDescription("Get screen dimensions."),
	), s.handleGetDimensions)
	addTool(s, mcp.NewTool("get_layer_count",
		mcp.WithDescription("Get modal layer count."),
	), s.handleGetLayerCount)
	addTool(s, mcp.NewTool("get_cell",
		mcp.WithDescription("Get the screen cell at coordinates."),
		mcp.WithInputSchema[cellArgs](),
	), s.handleGetCell)

	addTool(s, mcp.NewTool("find_by_label",
		mcp.WithDescription("Find a widget by label substring."),
		mcp.WithInputSchema[labelQueryArgs](),
	), s.handleFindByLabel)
	addTool(s, mcp.NewTool("find_by_role",
		mcp.WithDescription("Find widgets by role."),
		mcp.WithInputSchema[roleArgs](),
	), s.handleFindByRole)
	addTool(s, mcp.NewTool("find_by_id",
		mcp.WithDescription("Find a widget by ID."),
		mcp.WithInputSchema[idArgs](),
	), s.handleFindByID)
	addTool(s, mcp.NewTool("find_by_value",
		mcp.WithDescription("Find widgets by value substring."),
		mcp.WithInputSchema[valueArgs](),
	), s.handleFindByValue)
	addTool(s, mcp.NewTool("find_by_state",
		mcp.WithDescription("Find widgets by state match."),
		mcp.WithInputSchema[stateQueryArgs](),
	), s.handleFindByState)
	addTool(s, mcp.NewTool("find_at_position",
		mcp.WithDescription("Find widget at screen coordinates."),
		mcp.WithInputSchema[cellArgs](),
	), s.handleFindAtPosition)
	addTool(s, mcp.NewTool("find_focused",
		mcp.WithDescription("Find currently focused widget."),
	), s.handleFindFocused)
	addTool(s, mcp.NewTool("find_all",
		mcp.WithDescription("List all widgets."),
	), s.handleFindAll)
	addTool(s, mcp.NewTool("find_focusable",
		mcp.WithDescription("List focusable widgets."),
	), s.handleFindFocusable)
	addTool(s, mcp.NewTool("find_actionable",
		mcp.WithDescription("List actionable widgets."),
	), s.handleFindActionable)

	addTool(s, mcp.NewTool("get_children",
		mcp.WithDescription("Get direct children of a widget."),
		mcp.WithInputSchema[idArgs](),
	), s.handleGetChildren)
	addTool(s, mcp.NewTool("get_parent",
		mcp.WithDescription("Get parent of a widget."),
		mcp.WithInputSchema[idArgs](),
	), s.handleGetParent)
	addTool(s, mcp.NewTool("get_siblings",
		mcp.WithDescription("Get siblings of a widget."),
		mcp.WithInputSchema[idArgs](),
	), s.handleGetSiblings)
	addTool(s, mcp.NewTool("get_descendants",
		mcp.WithDescription("Get descendants of a widget."),
		mcp.WithInputSchema[idArgs](),
	), s.handleGetDescendants)
	addTool(s, mcp.NewTool("get_ancestors",
		mcp.WithDescription("Get ancestors of a widget."),
		mcp.WithInputSchema[idArgs](),
	), s.handleGetAncestors)
	addTool(s, mcp.NewTool("get_next_focusable",
		mcp.WithDescription("Get next focusable widget."),
		mcp.WithInputSchema[nextPrevArgs](),
	), s.handleGetNextFocusable)
	addTool(s, mcp.NewTool("get_prev_focusable",
		mcp.WithDescription("Get previous focusable widget."),
		mcp.WithInputSchema[nextPrevArgs](),
	), s.handleGetPrevFocusable)

	addTool(s, mcp.NewTool("get_label",
		mcp.WithDescription("Get widget label."),
		mcp.WithInputSchema[idArgs](),
	), s.handleGetLabel)
	addTool(s, mcp.NewTool("get_role",
		mcp.WithDescription("Get widget role."),
		mcp.WithInputSchema[idArgs](),
	), s.handleGetRole)
	addTool(s, mcp.NewTool("get_value",
		mcp.WithDescription("Get widget value."),
		mcp.WithInputSchema[idArgs](),
	), s.handleGetValue)
	addTool(s, mcp.NewTool("get_description",
		mcp.WithDescription("Get widget description."),
		mcp.WithInputSchema[idArgs](),
	), s.handleGetDescription)
	addTool(s, mcp.NewTool("get_bounds",
		mcp.WithDescription("Get widget bounds."),
		mcp.WithInputSchema[idArgs](),
	), s.handleGetBounds)
	addTool(s, mcp.NewTool("get_state",
		mcp.WithDescription("Get widget state."),
		mcp.WithInputSchema[idArgs](),
	), s.handleGetState)
	addTool(s, mcp.NewTool("get_actions",
		mcp.WithDescription("Get widget actions."),
		mcp.WithInputSchema[idArgs](),
	), s.handleGetActions)
	addTool(s, mcp.NewTool("is_focused",
		mcp.WithDescription("Check widget focus."),
		mcp.WithInputSchema[idArgs](),
	), s.handleIsFocused)
	addTool(s, mcp.NewTool("is_enabled",
		mcp.WithDescription("Check widget enabled state."),
		mcp.WithInputSchema[idArgs](),
	), s.handleIsEnabled)
	addTool(s, mcp.NewTool("is_visible",
		mcp.WithDescription("Check widget visibility."),
		mcp.WithInputSchema[idArgs](),
	), s.handleIsVisible)
	addTool(s, mcp.NewTool("is_checked",
		mcp.WithDescription("Check widget checked state."),
		mcp.WithInputSchema[idArgs](),
	), s.handleIsChecked)
	addTool(s, mcp.NewTool("is_expanded",
		mcp.WithDescription("Check widget expanded state."),
		mcp.WithInputSchema[idArgs](),
	), s.handleIsExpanded)
	addTool(s, mcp.NewTool("is_selected",
		mcp.WithDescription("Check widget selected state."),
		mcp.WithInputSchema[idArgs](),
	), s.handleIsSelected)
	addTool(s, mcp.NewTool("has_focus",
		mcp.WithDescription("Check if any widget is focused."),
	), s.handleHasFocus)

	addTool(s, mcp.NewTool("activate",
		mcp.WithDescription("Activate a widget."),
		mcp.WithInputSchema[actionArgs](),
	), s.handleActivate)
	addTool(s, mcp.NewTool("focus",
		mcp.WithDescription("Focus a widget."),
		mcp.WithInputSchema[actionArgs](),
	), s.handleFocus)
	addTool(s, mcp.NewTool("blur",
		mcp.WithDescription("Clear focus."),
	), s.handleBlur)
	addTool(s, mcp.NewTool("type_into",
		mcp.WithDescription("Type into a widget."),
		mcp.WithInputSchema[typeArgs](),
	), s.handleTypeInto)
	addTool(s, mcp.NewTool("clear",
		mcp.WithDescription("Clear widget input."),
		mcp.WithInputSchema[actionArgs](),
	), s.handleClear)
	addTool(s, mcp.NewTool("select_option",
		mcp.WithDescription("Select option by label."),
		mcp.WithInputSchema[selectOptionArgs](),
	), s.handleSelectOption)
	addTool(s, mcp.NewTool("select_index",
		mcp.WithDescription("Select option by index."),
		mcp.WithInputSchema[selectIndexArgs](),
	), s.handleSelectIndex)
	addTool(s, mcp.NewTool("toggle",
		mcp.WithDescription("Toggle a widget."),
		mcp.WithInputSchema[actionArgs](),
	), s.handleToggle)
	addTool(s, mcp.NewTool("check",
		mcp.WithDescription("Check a widget."),
		mcp.WithInputSchema[actionArgs](),
	), s.handleCheck)
	addTool(s, mcp.NewTool("uncheck",
		mcp.WithDescription("Uncheck a widget."),
		mcp.WithInputSchema[actionArgs](),
	), s.handleUncheck)
	addTool(s, mcp.NewTool("expand",
		mcp.WithDescription("Expand a widget."),
		mcp.WithInputSchema[actionArgs](),
	), s.handleExpand)
	addTool(s, mcp.NewTool("collapse",
		mcp.WithDescription("Collapse a widget."),
		mcp.WithInputSchema[actionArgs](),
	), s.handleCollapse)
	addTool(s, mcp.NewTool("scroll_to",
		mcp.WithDescription("Scroll to a widget."),
		mcp.WithInputSchema[actionArgs](),
	), s.handleScrollTo)
	addTool(s, mcp.NewTool("scroll_by",
		mcp.WithDescription("Scroll a widget by delta."),
		mcp.WithInputSchema[scrollByArgs](),
	), s.handleScrollBy)
	addTool(s, mcp.NewTool("scroll_to_top",
		mcp.WithDescription("Scroll a widget to top."),
		mcp.WithInputSchema[actionArgs](),
	), s.handleScrollToTop)
	addTool(s, mcp.NewTool("scroll_to_bottom",
		mcp.WithDescription("Scroll a widget to bottom."),
		mcp.WithInputSchema[actionArgs](),
	), s.handleScrollToBottom)

	addTool(s, mcp.NewTool("press_key",
		mcp.WithDescription("Press a single key."),
		mcp.WithInputSchema[keyArgs](),
	), s.handlePressKey)
	addTool(s, mcp.NewTool("press_keys",
		mcp.WithDescription("Press a sequence of keys."),
		mcp.WithInputSchema[keysArgs](),
	), s.handlePressKeys)
	addTool(s, mcp.NewTool("press_chord",
		mcp.WithDescription("Press a key chord."),
		mcp.WithInputSchema[chordArgs](),
	), s.handlePressChord)
	addTool(s, mcp.NewTool("press_rune",
		mcp.WithDescription("Press a single rune."),
		mcp.WithInputSchema[runeArgs](),
	), s.handlePressRune)
	addTool(s, mcp.NewTool("type_string",
		mcp.WithDescription("Type a string."),
		mcp.WithInputSchema[textArgs](),
	), s.handleTypeString)
	addTool(s, mcp.NewTool("press_enter", mcp.WithDescription("Press Enter.")), s.handlePressEnter)
	addTool(s, mcp.NewTool("press_escape", mcp.WithDescription("Press Escape.")), s.handlePressEscape)
	addTool(s, mcp.NewTool("press_tab", mcp.WithDescription("Press Tab.")), s.handlePressTab)
	addTool(s, mcp.NewTool("press_shift_tab", mcp.WithDescription("Press Shift+Tab.")), s.handlePressShiftTab)
	addTool(s, mcp.NewTool("press_space", mcp.WithDescription("Press Space.")), s.handlePressSpace)
	addTool(s, mcp.NewTool("press_backspace", mcp.WithDescription("Press Backspace.")), s.handlePressBackspace)
	addTool(s, mcp.NewTool("press_delete", mcp.WithDescription("Press Delete.")), s.handlePressDelete)
	addTool(s, mcp.NewTool("press_up", mcp.WithDescription("Press Up.")), s.handlePressUp)
	addTool(s, mcp.NewTool("press_down", mcp.WithDescription("Press Down.")), s.handlePressDown)
	addTool(s, mcp.NewTool("press_left", mcp.WithDescription("Press Left.")), s.handlePressLeft)
	addTool(s, mcp.NewTool("press_right", mcp.WithDescription("Press Right.")), s.handlePressRight)
	addTool(s, mcp.NewTool("press_home", mcp.WithDescription("Press Home.")), s.handlePressHome)
	addTool(s, mcp.NewTool("press_end", mcp.WithDescription("Press End.")), s.handlePressEnd)
	addTool(s, mcp.NewTool("press_page_up", mcp.WithDescription("Press Page Up.")), s.handlePressPageUp)
	addTool(s, mcp.NewTool("press_page_down", mcp.WithDescription("Press Page Down.")), s.handlePressPageDown)
	addTool(s, mcp.NewTool("press_f1", mcp.WithDescription("Press F1.")), s.handlePressF1)
	addTool(s, mcp.NewTool("press_f2", mcp.WithDescription("Press F2.")), s.handlePressF2)
	addTool(s, mcp.NewTool("press_f3", mcp.WithDescription("Press F3.")), s.handlePressF3)
	addTool(s, mcp.NewTool("press_f4", mcp.WithDescription("Press F4.")), s.handlePressF4)
	addTool(s, mcp.NewTool("press_f5", mcp.WithDescription("Press F5.")), s.handlePressF5)
	addTool(s, mcp.NewTool("press_f6", mcp.WithDescription("Press F6.")), s.handlePressF6)
	addTool(s, mcp.NewTool("press_f7", mcp.WithDescription("Press F7.")), s.handlePressF7)
	addTool(s, mcp.NewTool("press_f8", mcp.WithDescription("Press F8.")), s.handlePressF8)
	addTool(s, mcp.NewTool("press_f9", mcp.WithDescription("Press F9.")), s.handlePressF9)
	addTool(s, mcp.NewTool("press_f10", mcp.WithDescription("Press F10.")), s.handlePressF10)
	addTool(s, mcp.NewTool("press_f11", mcp.WithDescription("Press F11.")), s.handlePressF11)
	addTool(s, mcp.NewTool("press_f12", mcp.WithDescription("Press F12.")), s.handlePressF12)

	addTool(s, mcp.NewTool("mouse_click",
		mcp.WithDescription("Click mouse button at coordinates."),
		mcp.WithInputSchema[mouseArgs](),
	), s.handleMouseClick)
	addTool(s, mcp.NewTool("mouse_double_click",
		mcp.WithDescription("Double click at coordinates."),
		mcp.WithInputSchema[mouseArgs](),
	), s.handleMouseDoubleClick)
	addTool(s, mcp.NewTool("mouse_right_click",
		mcp.WithDescription("Right click at coordinates."),
		mcp.WithInputSchema[mouseArgs](),
	), s.handleMouseRightClick)
	addTool(s, mcp.NewTool("mouse_press",
		mcp.WithDescription("Mouse button press."),
		mcp.WithInputSchema[mouseButtonArgs](),
	), s.handleMousePress)
	addTool(s, mcp.NewTool("mouse_release",
		mcp.WithDescription("Mouse button release."),
		mcp.WithInputSchema[mouseButtonArgs](),
	), s.handleMouseRelease)
	addTool(s, mcp.NewTool("mouse_move",
		mcp.WithDescription("Move mouse."),
		mcp.WithInputSchema[cellArgs](),
	), s.handleMouseMove)
	addTool(s, mcp.NewTool("mouse_drag",
		mcp.WithDescription("Drag mouse."),
		mcp.WithInputSchema[mouseDragArgs](),
	), s.handleMouseDrag)
	addTool(s, mcp.NewTool("mouse_scroll_up",
		mcp.WithDescription("Scroll up."),
		mcp.WithInputSchema[mouseScrollArgs](),
	), s.handleMouseScrollUp)
	addTool(s, mcp.NewTool("mouse_scroll_down",
		mcp.WithDescription("Scroll down."),
		mcp.WithInputSchema[mouseScrollArgs](),
	), s.handleMouseScrollDown)
	addTool(s, mcp.NewTool("click_widget",
		mcp.WithDescription("Click a widget by label."),
		mcp.WithInputSchema[actionArgs](),
	), s.handleClickWidget)

	addTool(s, mcp.NewTool("clipboard_read", mcp.WithDescription("Read clipboard.")), s.handleClipboardRead)
	addTool(s, mcp.NewTool("clipboard_write",
		mcp.WithDescription("Write clipboard."),
		mcp.WithInputSchema[textArgs](),
	), s.handleClipboardWrite)
	addTool(s, mcp.NewTool("clipboard_clear", mcp.WithDescription("Clear clipboard.")), s.handleClipboardClear)
	addTool(s, mcp.NewTool("clipboard_has_text", mcp.WithDescription("Check clipboard text.")), s.handleClipboardHasText)
	addTool(s, mcp.NewTool("clipboard_read_primary", mcp.WithDescription("Read primary clipboard.")), s.handleClipboardReadPrimary)
	addTool(s, mcp.NewTool("clipboard_write_primary",
		mcp.WithDescription("Write primary clipboard."),
		mcp.WithInputSchema[textArgs](),
	), s.handleClipboardWritePrimary)
	addTool(s, mcp.NewTool("select_all", mcp.WithDescription("Select all text.")), s.handleSelectAll)
	addTool(s, mcp.NewTool("select_range",
		mcp.WithDescription("Select text range."),
		mcp.WithInputSchema[selectRangeArgs](),
	), s.handleSelectRange)
	addTool(s, mcp.NewTool("select_word", mcp.WithDescription("Select word.")), s.handleSelectWord)
	addTool(s, mcp.NewTool("select_line", mcp.WithDescription("Select line.")), s.handleSelectLine)
	addTool(s, mcp.NewTool("select_none", mcp.WithDescription("Clear selection.")), s.handleSelectNone)
	addTool(s, mcp.NewTool("get_selection", mcp.WithDescription("Get selection.")), s.handleGetSelection)
	addTool(s, mcp.NewTool("get_selection_bounds", mcp.WithDescription("Get selection bounds.")), s.handleGetSelectionBounds)
	addTool(s, mcp.NewTool("has_selection", mcp.WithDescription("Check selection.")), s.handleHasSelection)
	addTool(s, mcp.NewTool("copy", mcp.WithDescription("Copy selection.")), s.handleCopy)
	addTool(s, mcp.NewTool("cut", mcp.WithDescription("Cut selection.")), s.handleCut)
	addTool(s, mcp.NewTool("paste", mcp.WithDescription("Paste clipboard.")), s.handlePaste)
	addTool(s, mcp.NewTool("paste_text",
		mcp.WithDescription("Paste text."),
		mcp.WithInputSchema[textArgs](),
	), s.handlePasteText)

	addTool(s, mcp.NewTool("get_cursor_position", mcp.WithDescription("Get cursor position.")), s.handleGetCursorPosition)
	addTool(s, mcp.NewTool("set_cursor_position",
		mcp.WithDescription("Set cursor position."),
		mcp.WithInputSchema[cursorPosArgs](),
	), s.handleSetCursorPosition)
	addTool(s, mcp.NewTool("get_cursor_offset", mcp.WithDescription("Get cursor offset.")), s.handleGetCursorOffset)
	addTool(s, mcp.NewTool("set_cursor_offset",
		mcp.WithDescription("Set cursor offset."),
		mcp.WithInputSchema[cursorOffsetArgs](),
	), s.handleSetCursorOffset)
	addTool(s, mcp.NewTool("cursor_to_start", mcp.WithDescription("Move cursor to start.")), s.handleCursorToStart)
	addTool(s, mcp.NewTool("cursor_to_end", mcp.WithDescription("Move cursor to end.")), s.handleCursorToEnd)
	addTool(s, mcp.NewTool("cursor_word_left", mcp.WithDescription("Move cursor word left.")), s.handleCursorWordLeft)
	addTool(s, mcp.NewTool("cursor_word_right", mcp.WithDescription("Move cursor word right.")), s.handleCursorWordRight)

	addTool(s, mcp.NewTool("tick", mcp.WithDescription("Wait a tick.")), s.handleTick)
	addTool(s, mcp.NewTool("wait_ms",
		mcp.WithDescription("Wait for milliseconds."),
		mcp.WithInputSchema[waitArgs](),
	), s.handleWaitMS)
	addTool(s, mcp.NewTool("wait_for_widget",
		mcp.WithDescription("Wait for widget to appear."),
		mcp.WithInputSchema[waitLabelArgs](),
	), s.handleWaitForWidget)
	addTool(s, mcp.NewTool("wait_for_widget_gone",
		mcp.WithDescription("Wait for widget to disappear."),
		mcp.WithInputSchema[waitLabelArgs](),
	), s.handleWaitForWidgetGone)
	addTool(s, mcp.NewTool("wait_for_text",
		mcp.WithDescription("Wait for text to appear."),
		mcp.WithInputSchema[waitTextArgs](),
	), s.handleWaitForText)
	addTool(s, mcp.NewTool("wait_for_text_gone",
		mcp.WithDescription("Wait for text to disappear."),
		mcp.WithInputSchema[waitTextArgs](),
	), s.handleWaitForTextGone)
	addTool(s, mcp.NewTool("wait_for_focus",
		mcp.WithDescription("Wait for focus on widget."),
		mcp.WithInputSchema[waitLabelArgs](),
	), s.handleWaitForFocus)
	addTool(s, mcp.NewTool("wait_for_value",
		mcp.WithDescription("Wait for widget value."),
		mcp.WithInputSchema[waitValueArgs](),
	), s.handleWaitForValue)
	addTool(s, mcp.NewTool("wait_for_enabled",
		mcp.WithDescription("Wait for widget enabled."),
		mcp.WithInputSchema[waitLabelArgs](),
	), s.handleWaitForEnabled)
	addTool(s, mcp.NewTool("wait_for_idle",
		mcp.WithDescription("Wait for idle."),
		mcp.WithInputSchema[waitIdleArgs](),
	), s.handleWaitForIdle)

	addTool(s, mcp.NewTool("resize",
		mcp.WithDescription("Resize screen."),
		mcp.WithInputSchema[resizeArgs](),
	), s.handleResize)
	addTool(s, mcp.NewTool("resize_width",
		mcp.WithDescription("Resize screen width."),
		mcp.WithInputSchema[resizeWidthArgs](),
	), s.handleResizeWidth)
	addTool(s, mcp.NewTool("resize_height",
		mcp.WithDescription("Resize screen height."),
		mcp.WithInputSchema[resizeHeightArgs](),
	), s.handleResizeHeight)

	addTool(s, mcp.NewTool("diff_snapshots",
		mcp.WithDescription("Diff two snapshots."),
		mcp.WithInputSchema[diffArgs](),
	), s.handleDiffSnapshots)
	addTool(s, mcp.NewTool("widgets_changed",
		mcp.WithDescription("Check widget tree changes."),
		mcp.WithInputSchema[snapshotArg](),
	), s.handleWidgetsChanged)
	addTool(s, mcp.NewTool("text_changed",
		mcp.WithDescription("Check text changes."),
		mcp.WithInputSchema[snapshotArg](),
	), s.handleTextChanged)

	addTool(s, mcp.NewTool("get_capabilities", mcp.WithDescription("Get server capabilities.")), s.handleGetCapabilities)
	addTool(s, mcp.NewTool("get_app_info", mcp.WithDescription("Get app info.")), s.handleGetAppInfo)
	addTool(s, mcp.NewTool("ping", mcp.WithDescription("Ping server.")), s.handlePing)
}

func addTool(s *Server, tool mcp.Tool, handler mcpserver.ToolHandlerFunc) {
	if s == nil || s.mcpServer == nil {
		return
	}
	s.mcpServer.AddTool(tool, handler)
}

type snapshotArgs struct {
	IncludeText bool `json:"include_text,omitempty"`
}

type regionArgs struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type cellArgs struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type labelQueryArgs struct {
	Label string `json:"label"`
	Index *int   `json:"index,omitempty"`
	Layer *int   `json:"layer,omitempty"`
}

type roleArgs struct {
	Role string `json:"role"`
}

type idArgs struct {
	ID string `json:"id"`
}

type valueArgs struct {
	Value string `json:"value"`
}

type stateQueryArgs struct {
	State StateSet `json:"state"`
}

type nextPrevArgs struct {
	ID string `json:"id,omitempty"`
}

type actionArgs struct {
	Label string `json:"label,omitempty"`
	ID    string `json:"id,omitempty"`
	Index *int   `json:"index,omitempty"`
	Layer *int   `json:"layer,omitempty"`
}

type typeArgs struct {
	Label string `json:"label,omitempty"`
	ID    string `json:"id,omitempty"`
	Text  string `json:"text"`
	Index *int   `json:"index,omitempty"`
	Layer *int   `json:"layer,omitempty"`
}

type selectOptionArgs struct {
	Label  string `json:"label,omitempty"`
	ID     string `json:"id,omitempty"`
	Option string `json:"option"`
}

type selectIndexArgs struct {
	Label string `json:"label,omitempty"`
	ID    string `json:"id,omitempty"`
	Index int    `json:"index"`
}

type scrollByArgs struct {
	Label string `json:"label,omitempty"`
	ID    string `json:"id,omitempty"`
	Delta int    `json:"delta"`
}

type keyArgs struct {
	Key string `json:"key"`
}

type keysArgs struct {
	Keys string `json:"keys"`
}

type chordArgs struct {
	Chord string `json:"chord"`
}

type runeArgs struct {
	Rune string `json:"rune"`
}

type textArgs struct {
	Text string `json:"text"`
}

type mouseArgs struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type mouseButtonArgs struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Button string `json:"button"`
}

type mouseDragArgs struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

type mouseScrollArgs struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Amount int `json:"amount,omitempty"`
}

type selectRangeArgs struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type cursorPosArgs struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type cursorOffsetArgs struct {
	Offset int `json:"offset"`
}

type waitArgs struct {
	MS int `json:"ms"`
}

type waitLabelArgs struct {
	Label     string `json:"label"`
	TimeoutMS int    `json:"timeout_ms,omitempty"`
}

type waitTextArgs struct {
	Text      string `json:"text"`
	TimeoutMS int    `json:"timeout_ms,omitempty"`
}

type waitValueArgs struct {
	Label     string `json:"label"`
	Value     string `json:"value"`
	TimeoutMS int    `json:"timeout_ms,omitempty"`
}

type waitIdleArgs struct {
	TimeoutMS int `json:"timeout_ms,omitempty"`
}

type resizeArgs struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type resizeWidthArgs struct {
	Width int `json:"width"`
}

type resizeHeightArgs struct {
	Height int `json:"height"`
}

type diffArgs struct {
	Before Snapshot `json:"before"`
	After  Snapshot `json:"after"`
}

type snapshotArg struct {
	Since Snapshot `json:"since"`
}

func (s *Server) handleSnapshot(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := snapshotArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	if args.IncludeText && !s.textAllowed() {
		return nil, textDeniedError("snapshot")
	}
	snap, err := s.agent.SnapshotWithContext(ctx, agent.SnapshotOptions{
		IncludeText: args.IncludeText,
	})
	if err != nil {
		return s.toolError("snapshot", err), nil
	}
	result := snapshotFromAgent(snap, args.IncludeText)
	return s.toolResult("snapshot", result), nil
}

func (s *Server) handleSnapshotText(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.textAllowed() {
		return nil, textDeniedError("snapshot_text")
	}
	return s.toolResult("snapshot_text", s.agent.CaptureText()), nil
}

func (s *Server) handleSnapshotRegion(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.textAllowed() {
		return nil, textDeniedError("snapshot_region")
	}
	args := regionArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	text := s.agent.CaptureRegion(args.X, args.Y, args.Width, args.Height)
	return s.toolResult("snapshot_region", text), nil
}

func (s *Server) handleGetDimensions(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	width, height := s.agent.Dimensions()
	return s.toolResult("get_dimensions", Dimensions{Width: width, Height: height}), nil
}

func (s *Server) handleGetLayerCount(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("get_layer_count", err), nil
	}
	return s.toolResult("get_layer_count", snap.LayerCount), nil
}

func (s *Server) handleGetCell(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := cellArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	cell, ok := s.agent.CellAt(args.X, args.Y)
	if !ok {
		return s.toolError("get_cell", errors.New("cell not found")), nil
	}
	info := cellInfoFromCell(cell, s.textAllowed())
	return s.toolResult("get_cell", info), nil
}

func (s *Server) handleFindByLabel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := labelQueryArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("find_by_label", err), nil
	}
	info, _, err := s.resolveLabel(snap, args.Label, args.Index, args.Layer, false)
	if err != nil {
		return s.toolError("find_by_label", err), nil
	}
	return s.toolResult("find_by_label", info), nil
}

func (s *Server) handleFindByRole(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := roleArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	role := strings.ToLower(strings.TrimSpace(args.Role))
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("find_by_role", err), nil
	}
	var results []WidgetInfo
	for _, widget := range snap.Widgets {
		if strings.ToLower(widget.Role) == role {
			results = append(results, widget)
		}
	}
	return s.toolResult("find_by_role", results), nil
}

func (s *Server) handleFindByID(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := idArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("find_by_id", err), nil
	}
	info := findWidgetByID(snap.Widgets, args.ID, s.opts.StrictLabelMatching)
	return s.toolResult("find_by_id", info), nil
}

func (s *Server) handleFindByValue(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := valueArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	value := strings.ToLower(strings.TrimSpace(args.Value))
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("find_by_value", err), nil
	}
	var results []WidgetInfo
	for _, widget := range snap.Widgets {
		if strings.Contains(strings.ToLower(widget.Value), value) {
			results = append(results, widget)
		}
	}
	return s.toolResult("find_by_value", results), nil
}

func (s *Server) handleFindByState(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := stateQueryArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("find_by_state", err), nil
	}
	var results []WidgetInfo
	for _, widget := range snap.Widgets {
		if stateMatches(widget.State, args.State) {
			results = append(results, widget)
		}
	}
	return s.toolResult("find_by_state", results), nil
}

func (s *Server) handleFindAtPosition(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := cellArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("find_at_position", err), nil
	}
	info := findWidgetAtPosition(snap, args.X, args.Y)
	return s.toolResult("find_at_position", info), nil
}

func (s *Server) handleFindFocused(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("find_focused", err), nil
	}
	info := findWidgetByID(snap.Widgets, snap.FocusedID, false)
	return s.toolResult("find_focused", info), nil
}

func (s *Server) handleFindAll(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("find_all", err), nil
	}
	return s.toolResult("find_all", snap.Widgets), nil
}

func (s *Server) handleFindFocusable(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("find_focusable", err), nil
	}
	var results []WidgetInfo
	for _, widget := range snap.Widgets {
		if hasAction(widget, "focus") {
			results = append(results, widget)
		}
	}
	return s.toolResult("find_focusable", results), nil
}

func (s *Server) handleFindActionable(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("find_actionable", err), nil
	}
	var results []WidgetInfo
	for _, widget := range snap.Widgets {
		if len(widget.Actions) > 0 {
			results = append(results, widget)
		}
	}
	return s.toolResult("find_actionable", results), nil
}

func (s *Server) handleGetChildren(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := idArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("get_children", err), nil
	}
	index := indexWidgets(snap.Widgets)
	parent, ok := index[args.ID]
	if !ok {
		return s.toolError("get_children", errors.New("widget not found")), nil
	}
	children := collectWidgets(index, parent.ChildrenIDs)
	return s.toolResult("get_children", children), nil
}

func (s *Server) handleGetParent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := idArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("get_parent", err), nil
	}
	index := indexWidgets(snap.Widgets)
	child, ok := index[args.ID]
	if !ok || child.ParentID == "" {
		return s.toolResult("get_parent", (*WidgetInfo)(nil)), nil
	}
	parent, ok := index[child.ParentID]
	if !ok {
		return s.toolResult("get_parent", (*WidgetInfo)(nil)), nil
	}
	return s.toolResult("get_parent", parent), nil
}

func (s *Server) handleGetSiblings(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := idArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("get_siblings", err), nil
	}
	index := indexWidgets(snap.Widgets)
	widget, ok := index[args.ID]
	if !ok {
		return s.toolError("get_siblings", errors.New("widget not found")), nil
	}
	var siblings []WidgetInfo
	for _, candidate := range snap.Widgets {
		if candidate.ParentID == widget.ParentID && candidate.ID != widget.ID {
			siblings = append(siblings, candidate)
		}
	}
	return s.toolResult("get_siblings", siblings), nil
}

func (s *Server) handleGetDescendants(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := idArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("get_descendants", err), nil
	}
	index := indexWidgets(snap.Widgets)
	widget, ok := index[args.ID]
	if !ok {
		return s.toolError("get_descendants", errors.New("widget not found")), nil
	}
	var descendants []WidgetInfo
	walkDescendants(index, widget.ChildrenIDs, &descendants)
	return s.toolResult("get_descendants", descendants), nil
}

func (s *Server) handleGetAncestors(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := idArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("get_ancestors", err), nil
	}
	index := indexWidgets(snap.Widgets)
	widget, ok := index[args.ID]
	if !ok {
		return s.toolError("get_ancestors", errors.New("widget not found")), nil
	}
	var ancestors []WidgetInfo
	current := widget
	for current.ParentID != "" {
		parent, ok := index[current.ParentID]
		if !ok {
			break
		}
		ancestors = append(ancestors, parent)
		current = parent
	}
	return s.toolResult("get_ancestors", ancestors), nil
}

func (s *Server) handleGetNextFocusable(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := nextPrevArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("get_next_focusable", err), nil
	}
	info := nextFocusable(snap, args.ID, true)
	return s.toolResult("get_next_focusable", info), nil
}

func (s *Server) handleGetPrevFocusable(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := nextPrevArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("get_prev_focusable", err), nil
	}
	info := nextFocusable(snap, args.ID, false)
	return s.toolResult("get_prev_focusable", info), nil
}

func (s *Server) handleGetLabel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.toolResult("get_label", s.widgetStringProperty(ctx, req, func(w WidgetInfo) string {
		return w.Label
	})), nil
}

func (s *Server) handleGetRole(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.toolResult("get_role", s.widgetStringProperty(ctx, req, func(w WidgetInfo) string {
		return w.Role
	})), nil
}

func (s *Server) handleGetValue(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.toolResult("get_value", s.widgetStringProperty(ctx, req, func(w WidgetInfo) string {
		return w.Value
	})), nil
}

func (s *Server) handleGetDescription(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.toolResult("get_description", s.widgetStringProperty(ctx, req, func(w WidgetInfo) string {
		return w.Description
	})), nil
}

func (s *Server) handleGetBounds(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget := s.widgetByID(ctx, req)
	if widget == nil {
		return s.toolError("get_bounds", errors.New("widget not found")), nil
	}
	return s.toolResult("get_bounds", widget.Bounds), nil
}

func (s *Server) handleGetState(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget := s.widgetByID(ctx, req)
	if widget == nil {
		return s.toolError("get_state", errors.New("widget not found")), nil
	}
	return s.toolResult("get_state", widget.State), nil
}

func (s *Server) handleGetActions(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget := s.widgetByID(ctx, req)
	if widget == nil {
		return s.toolError("get_actions", errors.New("widget not found")), nil
	}
	return s.toolResult("get_actions", widget.Actions), nil
}

func (s *Server) handleIsFocused(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget := s.widgetByID(ctx, req)
	if widget == nil {
		return s.toolError("is_focused", errors.New("widget not found")), nil
	}
	return s.toolResult("is_focused", widget.State.Focused), nil
}

func (s *Server) handleIsEnabled(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget := s.widgetByID(ctx, req)
	if widget == nil {
		return s.toolError("is_enabled", errors.New("widget not found")), nil
	}
	return s.toolResult("is_enabled", !widget.State.Disabled), nil
}

func (s *Server) handleIsVisible(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget := s.widgetByID(ctx, req)
	if widget == nil {
		return s.toolError("is_visible", errors.New("widget not found")), nil
	}
	dims := s.currentDimensions()
	visible := visibleBounds(widget.Bounds, dims)
	return s.toolResult("is_visible", visible), nil
}

func (s *Server) handleIsChecked(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget := s.widgetByID(ctx, req)
	if widget == nil {
		return s.toolError("is_checked", errors.New("widget not found")), nil
	}
	checked := widget.State.Checked != nil && *widget.State.Checked
	return s.toolResult("is_checked", checked), nil
}

func (s *Server) handleIsExpanded(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget := s.widgetByID(ctx, req)
	if widget == nil {
		return s.toolError("is_expanded", errors.New("widget not found")), nil
	}
	expanded := widget.State.Expanded != nil && *widget.State.Expanded
	return s.toolResult("is_expanded", expanded), nil
}

func (s *Server) handleIsSelected(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget := s.widgetByID(ctx, req)
	if widget == nil {
		return s.toolError("is_selected", errors.New("widget not found")), nil
	}
	return s.toolResult("is_selected", widget.State.Selected), nil
}

func (s *Server) handleHasFocus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("has_focus", err), nil
	}
	return s.toolResult("has_focus", snap.FocusedID != ""), nil
}

func (s *Server) handleActivate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := actionArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("activate", err), nil
	}
	target, result, err := s.resolveActionTarget(snap, args)
	if err != nil {
		return s.toolError("activate", err), nil
	}
	if target.ID == "" {
		return s.toolResult("activate", result), nil
	}
	if err := s.agent.FocusByID(target.ID); err != nil {
		return s.toolError("activate", err), nil
	}
	if err := s.agent.SendKey(terminal.KeyEnter); err != nil {
		return s.toolError("activate", err), nil
	}
	s.agent.Tick()
	result.WidgetID = target.ID
	result.ResolvedTo = target.ID
	return s.toolResult("activate", result), nil
}

func (s *Server) handleFocus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := actionArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("focus", err), nil
	}
	target, result, err := s.resolveActionTarget(snap, args)
	if err != nil {
		return s.toolError("focus", err), nil
	}
	if target.ID == "" {
		return s.toolResult("focus", result), nil
	}
	if err := s.agent.FocusByID(target.ID); err != nil {
		return s.toolError("focus", err), nil
	}
	s.agent.Tick()
	result.WidgetID = target.ID
	result.ResolvedTo = target.ID
	return s.toolResult("focus", result), nil
}

func (s *Server) handleBlur(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.agent.ClearFocus(); err != nil {
		return s.toolError("blur", err), nil
	}
	return s.toolResult("blur", true), nil
}

func (s *Server) handleTypeInto(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := typeArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("type_into", err), nil
	}
	target, result, err := s.resolveActionTarget(snap, actionArgs{
		Label: args.Label,
		ID:    args.ID,
		Index: args.Index,
		Layer: args.Layer,
	})
	if err != nil {
		return s.toolError("type_into", err), nil
	}
	if target.ID == "" {
		return s.toolResult("type_into", result), nil
	}
	if err := s.agent.FocusByID(target.ID); err != nil {
		return s.toolError("type_into", err), nil
	}
	if err := s.agent.SendKeyString(args.Text); err != nil {
		return s.toolError("type_into", err), nil
	}
	s.agent.Tick()
	result.WidgetID = target.ID
	result.ResolvedTo = target.ID
	return s.toolResult("type_into", result), nil
}

func (s *Server) handleClear(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := actionArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("clear", err), nil
	}
	target, result, err := s.resolveActionTarget(snap, args)
	if err != nil {
		return s.toolError("clear", err), nil
	}
	if target.ID == "" {
		return s.toolResult("clear", result), nil
	}
	if err := s.clearWidget(ctx, target.ID); err != nil {
		return s.toolError("clear", err), nil
	}
	result.WidgetID = target.ID
	result.ResolvedTo = target.ID
	return s.toolResult("clear", result), nil
}

func (s *Server) handleSelectOption(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := selectOptionArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("select_option", err), nil
	}
	var targetID string
	result := ActionResult{Status: "ok"}
	if strings.TrimSpace(args.ID) != "" {
		info := findWidgetByID(snap.Widgets, args.ID, false)
		if info == nil {
			return s.toolError("select_option", errors.New("widget not found")), nil
		}
		targetID = info.ID
		result.WidgetID = info.ID
		result.ResolvedTo = info.ID
	} else {
		if strings.TrimSpace(args.Label) == "" {
			return s.toolError("select_option", errors.New("label or id required")), nil
		}
		target, resolved, err := s.resolveLabel(snap, args.Label, nil, nil, s.opts.StrictLabelMatching)
		if err != nil {
			return s.toolError("select_option", err), nil
		}
		targetID = target.ID
		result = resolved
	}
	if targetID == "" {
		return s.toolError("select_option", errors.New("label or id required")), nil
	}
	if err := s.agent.SelectByID(targetID, args.Option); err != nil {
		return s.toolError("select_option", err), nil
	}
	result.WidgetID = targetID
	result.ResolvedTo = targetID
	return s.toolResult("select_option", result), nil
}

func (s *Server) handleSelectIndex(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := selectIndexArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("select_index", err), nil
	}
	var targetID string
	result := ActionResult{Status: "ok"}
	if strings.TrimSpace(args.ID) != "" {
		info := findWidgetByID(snap.Widgets, args.ID, false)
		if info == nil {
			return s.toolError("select_index", errors.New("widget not found")), nil
		}
		targetID = info.ID
		result.WidgetID = info.ID
		result.ResolvedTo = info.ID
	} else {
		if strings.TrimSpace(args.Label) == "" {
			return s.toolError("select_index", errors.New("label or id required")), nil
		}
		target, resolved, err := s.resolveLabel(snap, args.Label, nil, nil, s.opts.StrictLabelMatching)
		if err != nil {
			return s.toolError("select_index", err), nil
		}
		targetID = target.ID
		result = resolved
	}
	if targetID == "" {
		return s.toolError("select_index", errors.New("label or id required")), nil
	}
	err = s.agent.WithWidgetByID(ctx, targetID, func(w runtime.Widget, acc accessibility.Accessible) error {
		if selector, ok := w.(interface{ SetSelected(int) }); ok {
			selector.SetSelected(args.Index)
			return nil
		}
		return errors.New("widget does not support indexed selection")
	})
	if err != nil {
		return s.toolError("select_index", err), nil
	}
	result.WidgetID = targetID
	result.ResolvedTo = targetID
	return s.toolResult("select_index", result), nil
}

func (s *Server) handleToggle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCheckToggle(ctx, req, "toggle")
}

func (s *Server) handleCheck(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCheckToggle(ctx, req, "check")
}

func (s *Server) handleUncheck(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleCheckToggle(ctx, req, "uncheck")
}

func (s *Server) handleExpand(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleExpandCollapse(ctx, req, true)
}

func (s *Server) handleCollapse(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleExpandCollapse(ctx, req, false)
}

func (s *Server) handleScrollTo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleScrollAction(ctx, req, "scroll_to", func(ctrl scroll.Controller) {
		ctrl.ScrollTo(0, 0)
	})
}

func (s *Server) handleScrollBy(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := scrollByArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	return s.handleScrollActionWithArgs(ctx, "scroll_by", actionArgs{
		Label: args.Label,
		ID:    args.ID,
	}, func(ctrl scroll.Controller) {
		ctrl.ScrollBy(0, args.Delta)
	})
}

func (s *Server) handleScrollToTop(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleScrollAction(ctx, req, "scroll_to_top", func(ctrl scroll.Controller) {
		ctrl.ScrollToStart()
	})
}

func (s *Server) handleScrollToBottom(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleScrollAction(ctx, req, "scroll_to_bottom", func(ctrl scroll.Controller) {
		ctrl.ScrollToEnd()
	})
}

func (s *Server) handlePressKey(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := keyArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	if err := s.sendKeySequence(args.Key); err != nil {
		return s.toolError("press_key", err), nil
	}
	return s.toolResult("press_key", true), nil
}

func (s *Server) handlePressKeys(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := keysArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	if err := s.sendKeySequence(args.Keys); err != nil {
		return s.toolError("press_keys", err), nil
	}
	return s.toolResult("press_keys", true), nil
}

func (s *Server) handlePressChord(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := chordArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	if err := s.sendKeySequence(args.Chord); err != nil {
		return s.toolError("press_chord", err), nil
	}
	return s.toolResult("press_chord", true), nil
}

func (s *Server) handlePressRune(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := runeArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	runes := []rune(args.Rune)
	if len(runes) != 1 {
		return s.toolError("press_rune", errors.New("single rune required")), nil
	}
	if err := s.agent.SendKeyMsg(runtime.KeyMsg{Key: terminal.KeyRune, Rune: runes[0]}); err != nil {
		return s.toolError("press_rune", err), nil
	}
	s.agent.Tick()
	return s.toolResult("press_rune", true), nil
}

func (s *Server) handleTypeString(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := textArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	if err := s.agent.SendKeyString(args.Text); err != nil {
		return s.toolError("type_string", err), nil
	}
	return s.toolResult("type_string", true), nil
}

func (s *Server) handlePressEnter(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_enter", terminal.KeyEnter)
}

func (s *Server) handlePressEscape(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_escape", terminal.KeyEscape)
}

func (s *Server) handlePressTab(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_tab", terminal.KeyTab)
}

func (s *Server) handlePressShiftTab(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.agent.SendKeyMsg(runtime.KeyMsg{Key: terminal.KeyTab, Shift: true}); err != nil {
		return s.toolError("press_shift_tab", err), nil
	}
	s.agent.Tick()
	return s.toolResult("press_shift_tab", true), nil
}

func (s *Server) handlePressSpace(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_space", terminal.KeyRune, ' ')
}

func (s *Server) handlePressBackspace(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_backspace", terminal.KeyBackspace)
}

func (s *Server) handlePressDelete(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_delete", terminal.KeyDelete)
}

func (s *Server) handlePressUp(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_up", terminal.KeyUp)
}

func (s *Server) handlePressDown(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_down", terminal.KeyDown)
}

func (s *Server) handlePressLeft(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_left", terminal.KeyLeft)
}

func (s *Server) handlePressRight(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_right", terminal.KeyRight)
}

func (s *Server) handlePressHome(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_home", terminal.KeyHome)
}

func (s *Server) handlePressEnd(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_end", terminal.KeyEnd)
}

func (s *Server) handlePressPageUp(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_page_up", terminal.KeyPageUp)
}

func (s *Server) handlePressPageDown(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_page_down", terminal.KeyPageDown)
}

func (s *Server) handlePressF1(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_f1", terminal.KeyF1)
}

func (s *Server) handlePressF2(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_f2", terminal.KeyF2)
}

func (s *Server) handlePressF3(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_f3", terminal.KeyF3)
}

func (s *Server) handlePressF4(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_f4", terminal.KeyF4)
}

func (s *Server) handlePressF5(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_f5", terminal.KeyF5)
}

func (s *Server) handlePressF6(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_f6", terminal.KeyF6)
}

func (s *Server) handlePressF7(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_f7", terminal.KeyF7)
}

func (s *Server) handlePressF8(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_f8", terminal.KeyF8)
}

func (s *Server) handlePressF9(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_f9", terminal.KeyF9)
}

func (s *Server) handlePressF10(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_f10", terminal.KeyF10)
}

func (s *Server) handlePressF11(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_f11", terminal.KeyF11)
}

func (s *Server) handlePressF12(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.pressKeyTool("press_f12", terminal.KeyF12)
}

func (s *Server) handleMouseClick(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleMouseClickButton(req, runtime.MouseLeft, "mouse_click")
}

func (s *Server) handleMouseDoubleClick(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := mouseArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	for i := 0; i < 2; i++ {
		if err := s.clickMouse(args.X, args.Y, runtime.MouseLeft); err != nil {
			return s.toolError("mouse_double_click", err), nil
		}
	}
	return s.toolResult("mouse_double_click", true), nil
}

func (s *Server) handleMouseRightClick(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleMouseClickButton(req, runtime.MouseRight, "mouse_right_click")
}

func (s *Server) handleMousePress(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := mouseButtonArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	button, err := parseMouseButton(args.Button)
	if err != nil {
		return s.toolError("mouse_press", err), nil
	}
	msg := runtime.MouseMsg{X: args.X, Y: args.Y, Button: button, Action: runtime.MousePress}
	if err := s.agent.SendMouse(msg); err != nil {
		return s.toolError("mouse_press", err), nil
	}
	s.agent.Tick()
	return s.toolResult("mouse_press", true), nil
}

func (s *Server) handleMouseRelease(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := mouseButtonArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	button, err := parseMouseButton(args.Button)
	if err != nil {
		return s.toolError("mouse_release", err), nil
	}
	msg := runtime.MouseMsg{X: args.X, Y: args.Y, Button: button, Action: runtime.MouseRelease}
	if err := s.agent.SendMouse(msg); err != nil {
		return s.toolError("mouse_release", err), nil
	}
	s.agent.Tick()
	return s.toolResult("mouse_release", true), nil
}

func (s *Server) handleMouseMove(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := cellArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	msg := runtime.MouseMsg{X: args.X, Y: args.Y, Action: runtime.MouseMove}
	if err := s.agent.SendMouse(msg); err != nil {
		return s.toolError("mouse_move", err), nil
	}
	s.agent.Tick()
	return s.toolResult("mouse_move", true), nil
}

func (s *Server) handleMouseDrag(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := mouseDragArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	if err := s.agent.SendMouse(runtime.MouseMsg{X: args.X1, Y: args.Y1, Button: runtime.MouseLeft, Action: runtime.MousePress}); err != nil {
		return s.toolError("mouse_drag", err), nil
	}
	if err := s.agent.SendMouse(runtime.MouseMsg{X: args.X2, Y: args.Y2, Button: runtime.MouseLeft, Action: runtime.MouseMove}); err != nil {
		return s.toolError("mouse_drag", err), nil
	}
	if err := s.agent.SendMouse(runtime.MouseMsg{X: args.X2, Y: args.Y2, Button: runtime.MouseLeft, Action: runtime.MouseRelease}); err != nil {
		return s.toolError("mouse_drag", err), nil
	}
	s.agent.Tick()
	return s.toolResult("mouse_drag", true), nil
}

func (s *Server) handleMouseScrollUp(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleMouseScroll(req, runtime.MouseWheelUp, "mouse_scroll_up")
}

func (s *Server) handleMouseScrollDown(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleMouseScroll(req, runtime.MouseWheelDown, "mouse_scroll_down")
}

func (s *Server) handleClickWidget(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := actionArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("click_widget", err), nil
	}
	target, result, err := s.resolveActionTarget(snap, args)
	if err != nil {
		return s.toolError("click_widget", err), nil
	}
	if target.ID == "" {
		return s.toolResult("click_widget", result), nil
	}
	centerX := target.Bounds.X + target.Bounds.Width/2
	centerY := target.Bounds.Y + target.Bounds.Height/2
	if err := s.clickMouse(centerX, centerY, runtime.MouseLeft); err != nil {
		return s.toolError("click_widget", err), nil
	}
	result.WidgetID = target.ID
	result.ResolvedTo = target.ID
	return s.toolResult("click_widget", result), nil
}

func (s *Server) handleClipboardRead(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.clipboardAllowed() {
		return nil, clipboardDeniedError("clipboard_read")
	}
	text, err := s.clipboard().Read()
	if err != nil {
		return s.toolError("clipboard_read", err), nil
	}
	return s.toolResult("clipboard_read", text), nil
}

func (s *Server) handleClipboardWrite(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.clipboardAllowed() {
		return nil, clipboardDeniedError("clipboard_write")
	}
	args := textArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	if err := s.clipboard().Write(args.Text); err != nil {
		return s.toolError("clipboard_write", err), nil
	}
	return s.toolResult("clipboard_write", true), nil
}

func (s *Server) handleClipboardClear(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.clipboardAllowed() {
		return nil, clipboardDeniedError("clipboard_clear")
	}
	if err := s.clipboard().Write(""); err != nil {
		return s.toolError("clipboard_clear", err), nil
	}
	return s.toolResult("clipboard_clear", true), nil
}

func (s *Server) handleClipboardHasText(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.clipboardAllowed() {
		return nil, clipboardDeniedError("clipboard_has_text")
	}
	text, err := s.clipboard().Read()
	if err != nil {
		return s.toolError("clipboard_has_text", err), nil
	}
	return s.toolResult("clipboard_has_text", text != ""), nil
}

func (s *Server) handleClipboardReadPrimary(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.toolError("clipboard_read_primary", errors.New("primary clipboard not supported")), nil
}

func (s *Server) handleClipboardWritePrimary(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.toolError("clipboard_write_primary", errors.New("primary clipboard not supported")), nil
}

func (s *Server) handleSelectAll(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget, err := s.focusedWidget(ctx)
	if err != nil {
		return s.toolError("select_all", err), nil
	}
	selectable, ok := widget.(widgets.Selectable)
	if !ok {
		return s.toolError("select_all", errors.New("focused widget does not support selection")), nil
	}
	selectable.SelectAll()
	return s.toolResult("select_all", true), nil
}

func (s *Server) handleSelectRange(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var args struct {
		Start int `json:"start"`
		End   int `json:"end"`
	}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	widget, err := s.focusedWidget(ctx)
	if err != nil {
		return s.toolError("select_range", err), nil
	}
	selectable, ok := widget.(widgets.Selectable)
	if !ok {
		return s.toolError("select_range", errors.New("focused widget does not support selection")), nil
	}
	selectable.SetSelection(widgets.Selection{Start: args.Start, End: args.End})
	return s.toolResult("select_range", true), nil
}

func (s *Server) handleSelectWord(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget, err := s.focusedWidget(ctx)
	if err != nil {
		return s.toolError("select_word", err), nil
	}
	selectable, ok := widget.(widgets.Selectable)
	if !ok {
		return s.toolError("select_word", errors.New("focused widget does not support selection")), nil
	}
	selectable.SelectWord()
	return s.toolResult("select_word", true), nil
}

func (s *Server) handleSelectLine(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget, err := s.focusedWidget(ctx)
	if err != nil {
		return s.toolError("select_line", err), nil
	}
	selectable, ok := widget.(widgets.Selectable)
	if !ok {
		return s.toolError("select_line", errors.New("focused widget does not support selection")), nil
	}
	selectable.SelectLine()
	return s.toolResult("select_line", true), nil
}

func (s *Server) handleSelectNone(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget, err := s.focusedWidget(ctx)
	if err != nil {
		return s.toolError("select_none", err), nil
	}
	selectable, ok := widget.(widgets.Selectable)
	if !ok {
		return s.toolError("select_none", errors.New("focused widget does not support selection")), nil
	}
	selectable.SelectNone()
	return s.toolResult("select_none", true), nil
}

func (s *Server) handleGetSelection(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget, err := s.focusedWidget(ctx)
	if err != nil {
		return s.toolError("get_selection", err), nil
	}
	selectable, ok := widget.(widgets.Selectable)
	if !ok {
		return s.toolError("get_selection", errors.New("focused widget does not support selection")), nil
	}
	text := selectable.GetSelectedText()
	return s.toolResult("get_selection", map[string]any{"text": text}), nil
}

func (s *Server) handleGetSelectionBounds(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget, err := s.focusedWidget(ctx)
	if err != nil {
		return s.toolError("get_selection_bounds", err), nil
	}
	selectable, ok := widget.(widgets.Selectable)
	if !ok {
		return s.toolError("get_selection_bounds", errors.New("focused widget does not support selection")), nil
	}
	sel := selectable.GetSelection()
	return s.toolResult("get_selection_bounds", map[string]any{
		"start": sel.Start,
		"end":   sel.End,
	}), nil
}

func (s *Server) handleHasSelection(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	widget, err := s.focusedWidget(ctx)
	if err != nil {
		return s.toolError("has_selection", err), nil
	}
	selectable, ok := widget.(widgets.Selectable)
	if !ok {
		return s.toolError("has_selection", errors.New("focused widget does not support selection")), nil
	}
	return s.toolResult("has_selection", selectable.HasSelection()), nil
}

func (s *Server) handleCopy(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleClipboardTarget(ctx, "copy")
}

func (s *Server) handleCut(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleClipboardTarget(ctx, "cut")
}

func (s *Server) handlePaste(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.handleClipboardTarget(ctx, "paste")
}

func (s *Server) handlePasteText(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := textArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	return s.handleClipboardTargetWithText(ctx, args.Text)
}

func (s *Server) handleGetCursorPosition(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pos, err := s.cursorPosition(ctx)
	if err != nil {
		return s.toolError("get_cursor_position", err), nil
	}
	return s.toolResult("get_cursor_position", pos), nil
}

func (s *Server) handleSetCursorPosition(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := cursorPosArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	if err := s.setCursorPosition(ctx, args.X, args.Y); err != nil {
		return s.toolError("set_cursor_position", err), nil
	}
	return s.toolResult("set_cursor_position", true), nil
}

func (s *Server) handleGetCursorOffset(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	offset, err := s.cursorOffset(ctx)
	if err != nil {
		return s.toolError("get_cursor_offset", err), nil
	}
	return s.toolResult("get_cursor_offset", offset), nil
}

func (s *Server) handleSetCursorOffset(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := cursorOffsetArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	if err := s.setCursorOffset(ctx, args.Offset); err != nil {
		return s.toolError("set_cursor_offset", err), nil
	}
	return s.toolResult("set_cursor_offset", true), nil
}

func (s *Server) handleCursorToStart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.setCursorOffset(ctx, 0); err != nil {
		return s.toolError("cursor_to_start", err), nil
	}
	return s.toolResult("cursor_to_start", true), nil
}

func (s *Server) handleCursorToEnd(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	length, err := s.cursorLength(ctx)
	if err != nil {
		return s.toolError("cursor_to_end", err), nil
	}
	if err := s.setCursorOffset(ctx, length); err != nil {
		return s.toolError("cursor_to_end", err), nil
	}
	return s.toolResult("cursor_to_end", true), nil
}

func (s *Server) handleCursorWordLeft(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.cursorWord(ctx, true); err != nil {
		return s.toolError("cursor_word_left", err), nil
	}
	return s.toolResult("cursor_word_left", true), nil
}

func (s *Server) handleCursorWordRight(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if err := s.cursorWord(ctx, false); err != nil {
		return s.toolError("cursor_word_right", err), nil
	}
	return s.toolResult("cursor_word_right", true), nil
}

func (s *Server) handleTick(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.agent.Tick()
	return s.toolResult("tick", true), nil
}

func (s *Server) handleWaitMS(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := waitArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	time.Sleep(time.Duration(args.MS) * time.Millisecond)
	return s.toolResult("wait_ms", true), nil
}

func (s *Server) handleWaitForWidget(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := waitLabelArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	timeout := waitTimeout(args.TimeoutMS)
	if err := s.agent.WaitForWidget(args.Label, timeout); err != nil {
		return s.toolError("wait_for_widget", err), nil
	}
	return s.toolResult("wait_for_widget", true), nil
}

func (s *Server) handleWaitForWidgetGone(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := waitLabelArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	timeout := waitTimeout(args.TimeoutMS)
	if err := s.agent.WaitForWidgetGone(args.Label, timeout); err != nil {
		return s.toolError("wait_for_widget_gone", err), nil
	}
	return s.toolResult("wait_for_widget_gone", true), nil
}

func (s *Server) handleWaitForText(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.textAllowed() {
		return nil, textDeniedError("wait_for_text")
	}
	args := waitTextArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	timeout := waitTimeout(args.TimeoutMS)
	if err := s.agent.WaitForText(args.Text, timeout); err != nil {
		return s.toolError("wait_for_text", err), nil
	}
	return s.toolResult("wait_for_text", true), nil
}

func (s *Server) handleWaitForTextGone(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.textAllowed() {
		return nil, textDeniedError("wait_for_text_gone")
	}
	args := waitTextArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	timeout := waitTimeout(args.TimeoutMS)
	if err := s.agent.WaitForTextGone(args.Text, timeout); err != nil {
		return s.toolError("wait_for_text_gone", err), nil
	}
	return s.toolResult("wait_for_text_gone", true), nil
}

func (s *Server) handleWaitForFocus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := waitLabelArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	timeout := waitTimeout(args.TimeoutMS)
	if err := s.agent.WaitForFocus(args.Label, timeout); err != nil {
		return s.toolError("wait_for_focus", err), nil
	}
	return s.toolResult("wait_for_focus", true), nil
}

func (s *Server) handleWaitForValue(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := waitValueArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	timeout := waitTimeout(args.TimeoutMS)
	if err := s.agent.WaitForValue(args.Label, args.Value, timeout); err != nil {
		return s.toolError("wait_for_value", err), nil
	}
	return s.toolResult("wait_for_value", true), nil
}

func (s *Server) handleWaitForEnabled(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := waitLabelArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	timeout := waitTimeout(args.TimeoutMS)
	if err := s.agent.WaitForEnabled(args.Label, timeout); err != nil {
		return s.toolError("wait_for_enabled", err), nil
	}
	return s.toolResult("wait_for_enabled", true), nil
}

func (s *Server) handleWaitForIdle(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := waitIdleArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	timeout := waitTimeout(args.TimeoutMS)
	if err := s.agent.WaitForIdle(timeout); err != nil {
		return s.toolError("wait_for_idle", err), nil
	}
	return s.toolResult("wait_for_idle", true), nil
}

func (s *Server) handleResize(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := resizeArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	if err := s.agent.SendResize(args.Width, args.Height); err != nil {
		return s.toolError("resize", err), nil
	}
	return s.toolResult("resize", true), nil
}

func (s *Server) handleResizeWidth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := resizeWidthArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	_, height := s.agent.Dimensions()
	if err := s.agent.SendResize(args.Width, height); err != nil {
		return s.toolError("resize_width", err), nil
	}
	return s.toolResult("resize_width", true), nil
}

func (s *Server) handleResizeHeight(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := resizeHeightArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	width, _ := s.agent.Dimensions()
	if err := s.agent.SendResize(width, args.Height); err != nil {
		return s.toolError("resize_height", err), nil
	}
	return s.toolResult("resize_height", true), nil
}

func (s *Server) handleDiffSnapshots(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := diffArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	diff := diffSnapshots(args.Before, args.After)
	return s.toolResult("diff_snapshots", diff), nil
}

func (s *Server) handleWidgetsChanged(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := snapshotArg{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	current, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("widgets_changed", err), nil
	}
	diff := diffSnapshots(args.Since, current)
	return s.toolResult("widgets_changed", len(diff.WidgetsAdded) > 0 || len(diff.WidgetsRemoved) > 0 || len(diff.WidgetsModified) > 0), nil
}

func (s *Server) handleTextChanged(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if !s.textAllowed() {
		return nil, textDeniedError("text_changed")
	}
	args := snapshotArg{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	current, err := s.currentSnapshot(ctx, true)
	if err != nil {
		return s.toolError("text_changed", err), nil
	}
	return s.toolResult("text_changed", current.Text != args.Since.Text), nil
}

func (s *Server) handleGetCapabilities(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	data := Capabilities{
		AllowText:      s.opts.AllowText || s.opts.TestBypassTextGating,
		AllowClipboard: s.opts.AllowClipboard || s.opts.TestBypassClipboardGating,
		Transport:      s.opts.Transport,
		Subscriptions:  true,
	}
	return s.toolResult("get_capabilities", data), nil
}

func (s *Server) handleGetAppInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	width, height := s.agent.Dimensions()
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError("get_app_info", err), nil
	}
	info := AppInfo{Width: width, Height: height, LayerCount: snap.LayerCount}
	if !s.startedAt.IsZero() {
		info.UptimeMs = time.Since(s.startedAt).Milliseconds()
	}
	return s.toolResult("get_app_info", info), nil
}

func (s *Server) handlePing(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.toolResult("ping", true), nil
}

func (s *Server) toolResult(tool string, data any) *mcp.CallToolResult {
	payload := map[string]any{
		"_schema": SchemaVersion,
		"_tool":   tool,
		"data":    data,
	}
	encoded, _ := json.Marshal(payload)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: string(encoded)},
		},
		StructuredContent: payload,
	}
}

func (s *Server) toolError(tool string, err error) *mcp.CallToolResult {
	payload := map[string]any{
		"_schema": SchemaVersion,
		"_tool":   tool,
		"error":   err.Error(),
	}
	encoded, _ := json.Marshal(payload)
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: string(encoded)},
		},
		StructuredContent: payload,
	}
}

func (s *Server) currentSnapshot(ctx context.Context, includeText bool) (Snapshot, error) {
	snap, err := s.agent.SnapshotWithContext(ctx, agent.SnapshotOptions{
		IncludeText: includeText,
	})
	if err != nil {
		return Snapshot{}, err
	}
	return snapshotFromAgent(snap, includeText), nil
}

func (s *Server) currentDimensions() Dimensions {
	width, height := s.agent.Dimensions()
	return Dimensions{Width: width, Height: height}
}

func (s *Server) widgetByID(ctx context.Context, req mcp.CallToolRequest) *WidgetInfo {
	args := idArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return nil
	}
	info := findWidgetByID(snap.Widgets, args.ID, false)
	if info == nil {
		return nil
	}
	return info
}

func (s *Server) widgetByExplicitOrID(ctx context.Context, id string) *WidgetInfo {
	if strings.TrimSpace(id) == "" {
		return nil
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return nil
	}
	return findWidgetByID(snap.Widgets, id, false)
}

func hasAction(widget WidgetInfo, name string) bool {
	for _, action := range widget.Actions {
		if action == name {
			return true
		}
	}
	return false
}

func indexWidgets(widgets []WidgetInfo) map[string]WidgetInfo {
	index := make(map[string]WidgetInfo, len(widgets))
	for _, widget := range widgets {
		index[widget.ID] = widget
	}
	return index
}

func collectWidgets(index map[string]WidgetInfo, ids []string) []WidgetInfo {
	out := make([]WidgetInfo, 0, len(ids))
	for _, id := range ids {
		if widget, ok := index[id]; ok {
			out = append(out, widget)
		}
	}
	return out
}

func walkDescendants(index map[string]WidgetInfo, ids []string, out *[]WidgetInfo) {
	for _, id := range ids {
		widget, ok := index[id]
		if !ok {
			continue
		}
		*out = append(*out, widget)
		walkDescendants(index, widget.ChildrenIDs, out)
	}
}

func stateMatches(widget StateSet, filter StateSet) bool {
	if filter.Focused && !widget.Focused {
		return false
	}
	if filter.Disabled && !widget.Disabled {
		return false
	}
	if filter.Hidden && !widget.Hidden {
		return false
	}
	if filter.Pressed && !widget.Pressed {
		return false
	}
	if filter.Selected && !widget.Selected {
		return false
	}
	if filter.ReadOnly && !widget.ReadOnly {
		return false
	}
	if filter.Required && !widget.Required {
		return false
	}
	if filter.Invalid && !widget.Invalid {
		return false
	}
	if filter.Busy && !widget.Busy {
		return false
	}
	if filter.Checked != nil {
		if widget.Checked == nil || *widget.Checked != *filter.Checked {
			return false
		}
	}
	if filter.Expanded != nil {
		if widget.Expanded == nil || *widget.Expanded != *filter.Expanded {
			return false
		}
	}
	return true
}

func visibleBounds(bounds Rect, dims Dimensions) bool {
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return false
	}
	if dims.Width <= 0 || dims.Height <= 0 {
		return false
	}
	if bounds.X >= dims.Width || bounds.Y >= dims.Height {
		return false
	}
	if bounds.X+bounds.Width <= 0 || bounds.Y+bounds.Height <= 0 {
		return false
	}
	return true
}

func cellInfoFromCell(cell backend.Cell, allowText bool) CellInfo {
	char := "?"
	if allowText {
		if cell.Rune == 0 {
			char = " "
		} else {
			char = string(cell.Rune)
		}
	}
	style := styleInfoFromStyle(cell.Style)
	return CellInfo{Char: char, Style: style}
}

func styleInfoFromStyle(style backend.Style) StyleInfo {
	attrs := style.Attributes()
	return StyleInfo{
		FG:            colorValueFromColor(style.ForegroundColor()),
		BG:            colorValueFromColor(style.BackgroundColor()),
		Bold:          attrs&backend.AttrBold != 0,
		Italic:        attrs&backend.AttrItalic != 0,
		Underline:     attrs&backend.AttrUnderline != 0,
		Dim:           attrs&backend.AttrDim != 0,
		Blink:         attrs&backend.AttrBlink != 0,
		Reverse:       attrs&backend.AttrReverse != 0,
		Strikethrough: attrs&backend.AttrStrikeThrough != 0,
	}
}

func colorValueFromColor(color backend.Color) any {
	if color == backend.ColorDefault {
		return "default"
	}
	if color.IsRGB() {
		r, g, b := color.RGB()
		return fmt.Sprintf("#%02x%02x%02x", r, g, b)
	}
	if color >= backend.ColorBlack && color <= backend.ColorWhite {
		return colorName(color)
	}
	return int(color)
}

func colorName(color backend.Color) string {
	switch color {
	case backend.ColorBlack:
		return "black"
	case backend.ColorRed:
		return "red"
	case backend.ColorGreen:
		return "green"
	case backend.ColorYellow:
		return "yellow"
	case backend.ColorBlue:
		return "blue"
	case backend.ColorMagenta:
		return "magenta"
	case backend.ColorCyan:
		return "cyan"
	case backend.ColorWhite:
		return "white"
	default:
		return "default"
	}
}

func findWidgetByID(widgets []WidgetInfo, id string, strict bool) *WidgetInfo {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	if strings.HasPrefix(id, "layer") && strings.Contains(id, ":") {
		for i := range widgets {
			if widgets[i].ID == id {
				return &widgets[i]
			}
		}
		return nil
	}
	var matches []WidgetInfo
	for _, widget := range widgets {
		explicit := explicitIDFromWidgetID(widget.ID)
		if explicit == "" {
			continue
		}
		base := explicitBaseID(explicit)
		if explicit == id || base == id {
			matches = append(matches, widget)
		}
	}
	if len(matches) == 0 {
		return nil
	}
	if strict && len(matches) > 1 {
		return nil
	}
	return &matches[0]
}

func findWidgetAtPosition(snap Snapshot, x, y int) *WidgetInfo {
	var candidates []WidgetInfo
	for _, widget := range snap.Widgets {
		if x >= widget.Bounds.X && x < widget.Bounds.X+widget.Bounds.Width &&
			y >= widget.Bounds.Y && y < widget.Bounds.Y+widget.Bounds.Height {
			candidates = append(candidates, widget)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	best := candidates[0]
	for _, candidate := range candidates[1:] {
		if layerFromID(candidate.ID) > layerFromID(best.ID) {
			best = candidate
			continue
		}
	}
	return &best
}

func nextFocusable(snap Snapshot, startID string, forward bool) *WidgetInfo {
	focusables := make([]WidgetInfo, 0)
	layer := 0
	if startID == "" {
		startID = snap.FocusedID
	}
	if startID != "" {
		layer = layerFromID(startID)
	}
	for _, widget := range snap.Widgets {
		if layerFromID(widget.ID) != layer {
			continue
		}
		if hasAction(widget, "focus") {
			focusables = append(focusables, widget)
		}
	}
	if len(focusables) == 0 {
		return nil
	}
	startIndex := -1
	for i, widget := range focusables {
		if widget.ID == startID {
			startIndex = i
			break
		}
	}
	if startIndex == -1 {
		return &focusables[0]
	}
	if forward {
		return &focusables[(startIndex+1)%len(focusables)]
	}
	return &focusables[(startIndex-1+len(focusables))%len(focusables)]
}

func (s *Server) resolveLabel(snap Snapshot, label string, index *int, layer *int, strict bool) (*WidgetInfo, ActionResult, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return nil, ActionResult{}, errors.New("label is required")
	}
	matches := labelMatches(snap, label)
	if layer != nil {
		filtered := matches[:0]
		for _, match := range matches {
			if match.Layer == *layer {
				filtered = append(filtered, match)
			}
		}
		matches = filtered
	}
	if len(matches) == 0 {
		return nil, ActionResult{}, errors.New("widget not found")
	}
	if index != nil {
		if *index < 0 || *index >= len(matches) {
			return nil, ActionResult{}, fmt.Errorf("index %d out of range", *index)
		}
		chosen := matches[*index]
		result := ActionResult{Status: "ok", WidgetID: chosen.Widget.ID, ResolvedTo: chosen.Widget.ID}
		return &chosen.Widget, result, nil
	}
	if len(matches) == 1 {
		chosen := matches[0]
		result := ActionResult{Status: "ok", WidgetID: chosen.Widget.ID, ResolvedTo: chosen.Widget.ID}
		return &chosen.Widget, result, nil
	}
	if strict {
		return nil, ActionResult{}, errors.New("multiple widgets match label")
	}
	chosen, reason := resolveLabelMatches(matches)
	result := ActionResult{
		Status:           "ambiguous",
		WidgetID:         chosen.Widget.ID,
		ResolvedTo:       chosen.Widget.ID,
		Message:          fmt.Sprintf("multiple widgets match label %q", label),
		Matches:          matchSummaries(matches, snap),
		ResolutionReason: reason,
	}
	return &chosen.Widget, result, nil
}

func (s *Server) resolveActionTarget(snap Snapshot, args actionArgs) (WidgetInfo, ActionResult, error) {
	if strings.TrimSpace(args.ID) != "" {
		info := findWidgetByID(snap.Widgets, args.ID, false)
		if info == nil {
			return WidgetInfo{}, ActionResult{}, errors.New("widget not found")
		}
		return *info, ActionResult{Status: "ok", WidgetID: info.ID, ResolvedTo: info.ID}, nil
	}
	widget, result, err := s.resolveLabel(snap, args.Label, args.Index, args.Layer, s.opts.StrictLabelMatching)
	if err != nil {
		return WidgetInfo{}, ActionResult{}, err
	}
	if widget == nil {
		return WidgetInfo{}, result, errors.New("widget not found")
	}
	return *widget, result, nil
}

func labelMatches(snap Snapshot, label string) []labelMatch {
	label = strings.ToLower(label)
	var matches []labelMatch
	for i, widget := range snap.Widgets {
		if strings.Contains(strings.ToLower(widget.Label), label) {
			matches = append(matches, labelMatch{
				Widget:       widget,
				Order:        i,
				Layer:        layerFromID(widget.ID),
				FullyVisible: widgetVisible(widget, snap.Dimensions),
				VisibleArea:  visibleArea(widget, snap.Dimensions),
				Focused:      snap.FocusedID == widget.ID,
			})
		}
	}
	return matches
}

type labelMatch struct {
	Widget       WidgetInfo
	Order        int
	Layer        int
	FullyVisible bool
	VisibleArea  int
	Focused      bool
}

func resolveLabelMatches(matches []labelMatch) (labelMatch, string) {
	for _, match := range matches {
		if match.Focused {
			return match, "focused"
		}
	}
	best := matches[0]
	reason := "topmost_layer"
	for _, match := range matches[1:] {
		if match.Layer > best.Layer {
			best = match
			reason = "topmost_layer"
		}
	}
	for _, match := range matches {
		if match.Layer != best.Layer {
			continue
		}
		if match.FullyVisible && !best.FullyVisible {
			best = match
			reason = "fully_visible"
		}
		if match.FullyVisible == best.FullyVisible && match.VisibleArea > best.VisibleArea {
			best = match
			reason = "visibility"
		}
	}
	for _, match := range matches {
		if match.Layer == best.Layer && match.FullyVisible == best.FullyVisible && match.VisibleArea == best.VisibleArea {
			if match.Order < best.Order {
				best = match
				reason = "dom_order"
			}
		}
	}
	return best, reason
}

func matchSummaries(matches []labelMatch, snap Snapshot) []WidgetMatch {
	index := indexWidgets(snap.Widgets)
	out := make([]WidgetMatch, 0, len(matches))
	for _, match := range matches {
		context := ""
		if parent, ok := index[match.Widget.ParentID]; ok {
			context = parent.Label
		}
		out = append(out, WidgetMatch{
			ID:      match.Widget.ID,
			Label:   match.Widget.Label,
			Context: context,
			Layer:   match.Layer,
		})
	}
	return out
}

func widgetVisible(widget WidgetInfo, dims Dimensions) bool {
	if !visibleBounds(widget.Bounds, dims) {
		return false
	}
	return widget.Bounds.X >= 0 && widget.Bounds.Y >= 0 &&
		widget.Bounds.X+widget.Bounds.Width <= dims.Width &&
		widget.Bounds.Y+widget.Bounds.Height <= dims.Height
}

func visibleArea(widget WidgetInfo, dims Dimensions) int {
	if !visibleBounds(widget.Bounds, dims) {
		return 0
	}
	x1 := max(widget.Bounds.X, 0)
	y1 := max(widget.Bounds.Y, 0)
	x2 := min(widget.Bounds.X+widget.Bounds.Width, dims.Width)
	y2 := min(widget.Bounds.Y+widget.Bounds.Height, dims.Height)
	if x2 <= x1 || y2 <= y1 {
		return 0
	}
	return (x2 - x1) * (y2 - y1)
}

func (s *Server) widgetStringProperty(ctx context.Context, req mcp.CallToolRequest, fn func(WidgetInfo) string) string {
	widget := s.widgetByID(ctx, req)
	if widget == nil {
		return ""
	}
	return fn(*widget)
}

func (s *Server) clearWidget(ctx context.Context, id string) error {
	return s.agent.WithWidgetByID(ctx, id, func(w runtime.Widget, acc accessibility.Accessible) error {
		if clearer, ok := w.(interface{ Clear() }); ok {
			clearer.Clear()
			return nil
		}
		if setter, ok := w.(interface{ SetText(string) }); ok {
			setter.SetText("")
			return nil
		}
		return errors.New("widget does not support clear")
	})
}

func (s *Server) handleCheckToggle(ctx context.Context, req mcp.CallToolRequest, mode string) (*mcp.CallToolResult, error) {
	args := actionArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError(mode, err), nil
	}
	target, result, err := s.resolveActionTarget(snap, args)
	if err != nil {
		return s.toolError(mode, err), nil
	}
	if target.ID == "" {
		return s.toolResult(mode, result), nil
	}
	if err := s.agent.FocusByID(target.ID); err != nil {
		return s.toolError(mode, err), nil
	}
	desired := target.State.Checked
	switch mode {
	case "check":
		if desired != nil && *desired {
			return s.toolResult(mode, result), nil
		}
	case "uncheck":
		if desired != nil && !*desired {
			return s.toolResult(mode, result), nil
		}
	}
	if err := s.agent.SendKeyMsg(runtime.KeyMsg{Key: terminal.KeyRune, Rune: ' '}); err != nil {
		return s.toolError(mode, err), nil
	}
	s.agent.Tick()
	result.WidgetID = target.ID
	result.ResolvedTo = target.ID
	return s.toolResult(mode, result), nil
}

func (s *Server) handleExpandCollapse(ctx context.Context, req mcp.CallToolRequest, expand bool) (*mcp.CallToolResult, error) {
	tool := "expand"
	key := terminal.KeyRight
	if !expand {
		tool = "collapse"
		key = terminal.KeyLeft
	}

	args := actionArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError(tool, err), nil
	}
	target, result, err := s.resolveActionTarget(snap, args)
	if err != nil {
		return s.toolError(tool, err), nil
	}
	if target.ID == "" {
		return s.toolResult(tool, result), nil
	}
	if err := s.agent.FocusByID(target.ID); err != nil {
		return s.toolError(tool, err), nil
	}
	if expand {
		if target.State.Expanded != nil && *target.State.Expanded {
			return s.toolResult(tool, result), nil
		}
	} else if target.State.Expanded != nil && !*target.State.Expanded {
		return s.toolResult(tool, result), nil
	}
	if err := s.agent.SendKey(key); err != nil {
		return s.toolError(tool, err), nil
	}
	s.agent.Tick()
	result.WidgetID = target.ID
	result.ResolvedTo = target.ID
	return s.toolResult(tool, result), nil
}

func (s *Server) handleScrollAction(ctx context.Context, req mcp.CallToolRequest, tool string, fn func(scroll.Controller)) (*mcp.CallToolResult, error) {
	args := actionArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	return s.handleScrollActionWithArgs(ctx, tool, args, fn)
}

func (s *Server) handleScrollActionWithArgs(ctx context.Context, tool string, args actionArgs, fn func(scroll.Controller)) (*mcp.CallToolResult, error) {
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return s.toolError(tool, err), nil
	}
	target, result, err := s.resolveActionTarget(snap, args)
	if err != nil {
		return s.toolError(tool, err), nil
	}
	if target.ID == "" {
		return s.toolResult(tool, result), nil
	}
	err = s.agent.WithWidgetByID(ctx, target.ID, func(w runtime.Widget, acc accessibility.Accessible) error {
		controller, ok := w.(scroll.Controller)
		if !ok {
			return errors.New("widget is not scrollable")
		}
		fn(controller)
		return nil
	})
	if err != nil {
		return s.toolError(tool, err), nil
	}
	result.WidgetID = target.ID
	result.ResolvedTo = target.ID
	return s.toolResult(tool, result), nil
}

func (s *Server) sendKeySequence(sequence string) error {
	keys, err := keybind.ParseKeySequence(sequence)
	if err != nil {
		return err
	}
	for _, press := range keys.Sequence {
		if err := s.agent.SendKeyMsg(runtime.KeyMsg{
			Key:   press.Key,
			Rune:  press.Rune,
			Alt:   press.Alt,
			Ctrl:  press.Ctrl,
			Shift: press.Shift,
		}); err != nil {
			return err
		}
		s.agent.Tick()
	}
	return nil
}

func (s *Server) pressKeyTool(name string, key terminal.Key, runeVal ...rune) (*mcp.CallToolResult, error) {
	msg := runtime.KeyMsg{Key: key}
	if len(runeVal) > 0 {
		msg.Rune = runeVal[0]
	}
	if err := s.agent.SendKeyMsg(msg); err != nil {
		return s.toolError(name, err), nil
	}
	s.agent.Tick()
	return s.toolResult(name, true), nil
}

func (s *Server) handleMouseClickButton(req mcp.CallToolRequest, button runtime.MouseButton, tool string) (*mcp.CallToolResult, error) {
	args := mouseArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	if err := s.clickMouse(args.X, args.Y, button); err != nil {
		return s.toolError(tool, err), nil
	}
	return s.toolResult(tool, true), nil
}

func (s *Server) clickMouse(x, y int, button runtime.MouseButton) error {
	if err := s.agent.SendMouse(runtime.MouseMsg{X: x, Y: y, Button: button, Action: runtime.MousePress}); err != nil {
		return err
	}
	if err := s.agent.SendMouse(runtime.MouseMsg{X: x, Y: y, Button: button, Action: runtime.MouseRelease}); err != nil {
		return err
	}
	s.agent.Tick()
	return nil
}

func (s *Server) handleMouseScroll(req mcp.CallToolRequest, button runtime.MouseButton, name string) (*mcp.CallToolResult, error) {
	args := mouseScrollArgs{}
	if err := req.BindArguments(&args); err != nil {
		return nil, newMCPError(mcp.INVALID_PARAMS, err.Error(), nil)
	}
	amount := args.Amount
	if amount <= 0 {
		amount = 1
	}
	for i := 0; i < amount; i++ {
		if err := s.agent.SendMouse(runtime.MouseMsg{X: args.X, Y: args.Y, Button: button, Action: runtime.MousePress}); err != nil {
			return s.toolError(name, err), nil
		}
	}
	s.agent.Tick()
	return s.toolResult(name, true), nil
}

func parseMouseButton(button string) (runtime.MouseButton, error) {
	switch strings.ToLower(strings.TrimSpace(button)) {
	case "none", "":
		return runtime.MouseNone, nil
	case "left":
		return runtime.MouseLeft, nil
	case "middle":
		return runtime.MouseMiddle, nil
	case "right":
		return runtime.MouseRight, nil
	case "wheel_up", "wheelup", "up":
		return runtime.MouseWheelUp, nil
	case "wheel_down", "wheeldown", "down":
		return runtime.MouseWheelDown, nil
	default:
		return runtime.MouseNone, fmt.Errorf("unknown button %q", button)
	}
}

func (s *Server) clipboard() clipboard.Clipboard {
	if s == nil || s.app == nil {
		return clipboard.UnavailableClipboard{}
	}
	cb := s.app.Services().Clipboard()
	if cb == nil {
		return clipboard.UnavailableClipboard{}
	}
	return cb
}

func (s *Server) focusedWidget(ctx context.Context) (runtime.Widget, error) {
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return nil, err
	}
	if snap.FocusedID == "" {
		return nil, errors.New("no focused widget")
	}
	var widget runtime.Widget
	err = s.agent.WithWidgetByID(ctx, snap.FocusedID, func(w runtime.Widget, acc accessibility.Accessible) error {
		widget = w
		return nil
	})
	return widget, err
}

func (s *Server) handleClipboardTarget(ctx context.Context, action string) (*mcp.CallToolResult, error) {
	if !s.clipboardAllowed() {
		return nil, clipboardDeniedError(action)
	}
	widget, err := s.focusedWidget(ctx)
	if err != nil {
		return s.toolError(action, err), nil
	}
	target, ok := widget.(clipboard.Target)
	if !ok {
		return s.toolError(action, errors.New("focused widget does not support clipboard")), nil
	}
	switch action {
	case "copy":
		if text, ok := target.ClipboardCopy(); ok {
			_ = s.clipboard().Write(text)
			return s.toolResult("copy", true), nil
		}
	case "cut":
		if text, ok := target.ClipboardCut(); ok {
			_ = s.clipboard().Write(text)
			return s.toolResult("cut", true), nil
		}
	case "paste":
		text, _ := s.clipboard().Read()
		if ok := target.ClipboardPaste(text); ok {
			return s.toolResult("paste", true), nil
		}
	}
	return s.toolError(action, errors.New("clipboard action failed")), nil
}

func (s *Server) handleClipboardTargetWithText(ctx context.Context, text string) (*mcp.CallToolResult, error) {
	if !s.clipboardAllowed() {
		return nil, clipboardDeniedError("paste_text")
	}
	widget, err := s.focusedWidget(ctx)
	if err != nil {
		return s.toolError("paste_text", err), nil
	}
	target, ok := widget.(clipboard.Target)
	if !ok {
		return s.toolError("paste_text", errors.New("focused widget does not support clipboard")), nil
	}
	if ok := target.ClipboardPaste(text); !ok {
		return s.toolError("paste_text", errors.New("clipboard paste failed")), nil
	}
	return s.toolResult("paste_text", true), nil
}

func (s *Server) cursorPosition(ctx context.Context) (CursorPosition, error) {
	return s.withCursor(ctx, func(cursor cursorHandler) (CursorPosition, error) {
		x, y := cursor.Position()
		return CursorPosition{X: x, Y: y}, nil
	})
}

func (s *Server) setCursorPosition(ctx context.Context, x, y int) error {
	_, err := s.withCursor(ctx, func(cursor cursorHandler) (CursorPosition, error) {
		cursor.SetPosition(x, y)
		return CursorPosition{}, nil
	})
	return err
}

func (s *Server) cursorOffset(ctx context.Context) (int, error) {
	return s.withCursorOffset(ctx, func(offset cursorOffset) (int, error) {
		return offset.Offset(), nil
	})
}

func (s *Server) setCursorOffset(ctx context.Context, offset int) error {
	_, err := s.withCursorOffset(ctx, func(cursor cursorOffset) (int, error) {
		cursor.SetOffset(offset)
		return 0, nil
	})
	return err
}

func (s *Server) cursorLength(ctx context.Context) (int, error) {
	return s.withCursorText(ctx, func(text cursorText) (int, error) {
		return len([]rune(text.Text())), nil
	})
}

func (s *Server) cursorWord(ctx context.Context, left bool) error {
	_, err := s.withCursorWord(ctx, func(word cursorWord) (bool, error) {
		if left {
			word.WordLeft()
		} else {
			word.WordRight()
		}
		return true, nil
	})
	return err
}

type cursorHandler interface {
	Position() (int, int)
	SetPosition(int, int)
}

type cursorOffset interface {
	Offset() int
	SetOffset(int)
}

type cursorText interface {
	Text() string
}

type cursorWord interface {
	WordLeft()
	WordRight()
}

func (s *Server) withCursor(ctx context.Context, fn func(cursor cursorHandler) (CursorPosition, error)) (CursorPosition, error) {
	var result CursorPosition
	err := s.withFocusedWidget(ctx, func(w runtime.Widget) error {
		cursor, ok := w.(interface {
			CursorPosition() (int, int)
			SetCursorPosition(int, int)
		})
		if !ok {
			return errors.New("focused widget does not support cursor positioning")
		}
		pos, err := fn(cursorAdapter{cursor})
		if err != nil {
			return err
		}
		result = pos
		return nil
	})
	return result, err
}

func (s *Server) withCursorOffset(ctx context.Context, fn func(cursor cursorOffset) (int, error)) (int, error) {
	var value int
	err := s.withFocusedWidget(ctx, func(w runtime.Widget) error {
		cursor, ok := w.(interface {
			CursorOffset() int
			SetCursorOffset(int)
		})
		if !ok {
			return errors.New("focused widget does not support cursor offset")
		}
		result, err := fn(cursorOffsetAdapter{cursor})
		if err != nil {
			return err
		}
		value = result
		return nil
	})
	return value, err
}

func (s *Server) withCursorText(ctx context.Context, fn func(text cursorText) (int, error)) (int, error) {
	var value int
	err := s.withFocusedWidget(ctx, func(w runtime.Widget) error {
		text, ok := w.(interface{ Text() string })
		if !ok {
			return errors.New("focused widget does not expose text")
		}
		result, err := fn(text)
		if err != nil {
			return err
		}
		value = result
		return nil
	})
	return value, err
}

func (s *Server) withCursorWord(ctx context.Context, fn func(word cursorWord) (bool, error)) (bool, error) {
	var ok bool
	err := s.withFocusedWidget(ctx, func(w runtime.Widget) error {
		word, ok := w.(interface {
			CursorWordLeft()
			CursorWordRight()
		})
		if !ok {
			return errors.New("focused widget does not support word navigation")
		}
		_, err := fn(cursorWordAdapter{word})
		return err
	})
	return ok, err
}

func (s *Server) withFocusedWidget(ctx context.Context, fn func(w runtime.Widget) error) error {
	if s == nil {
		return errors.New("mcp server is nil")
	}
	snap, err := s.currentSnapshot(ctx, false)
	if err != nil {
		return err
	}
	if snap.FocusedID == "" {
		return errors.New("no focused widget")
	}
	return s.agent.WithWidgetByID(ctx, snap.FocusedID, func(w runtime.Widget, acc accessibility.Accessible) error {
		return fn(w)
	})
}

type cursorAdapter struct {
	cursor interface {
		CursorPosition() (int, int)
		SetCursorPosition(int, int)
	}
}

func (c cursorAdapter) Position() (int, int) {
	return c.cursor.CursorPosition()
}

func (c cursorAdapter) SetPosition(x, y int) {
	c.cursor.SetCursorPosition(x, y)
}

type cursorOffsetAdapter struct {
	cursor interface {
		CursorOffset() int
		SetCursorOffset(int)
	}
}

func (c cursorOffsetAdapter) Offset() int {
	return c.cursor.CursorOffset()
}

func (c cursorOffsetAdapter) SetOffset(offset int) {
	c.cursor.SetCursorOffset(offset)
}

type cursorWordAdapter struct {
	cursor interface {
		CursorWordLeft()
		CursorWordRight()
	}
}

func (c cursorWordAdapter) WordLeft() {
	c.cursor.CursorWordLeft()
}

func (c cursorWordAdapter) WordRight() {
	c.cursor.CursorWordRight()
}

func waitTimeout(ms int) time.Duration {
	if ms <= 0 {
		ms = 2000
	}
	return time.Duration(ms) * time.Millisecond
}

func diffSnapshots(before Snapshot, after Snapshot) Diff {
	diff := Diff{
		TextChanged:       before.Text != after.Text,
		DimensionsChanged: before.Dimensions != after.Dimensions,
		LayerCountChanged: before.LayerCount != after.LayerCount,
		FocusChanged:      before.FocusedID != after.FocusedID,
	}
	beforeIndex := indexWidgets(before.Widgets)
	afterIndex := indexWidgets(after.Widgets)

	for id := range beforeIndex {
		if _, ok := afterIndex[id]; !ok {
			diff.WidgetsRemoved = append(diff.WidgetsRemoved, id)
		}
	}
	for id := range afterIndex {
		if _, ok := beforeIndex[id]; !ok {
			diff.WidgetsAdded = append(diff.WidgetsAdded, id)
		}
	}
	for id, beforeWidget := range beforeIndex {
		afterWidget, ok := afterIndex[id]
		if !ok {
			continue
		}
		changes := widgetChanges(beforeWidget, afterWidget)
		if len(changes) > 0 {
			diff.WidgetsModified = append(diff.WidgetsModified, WidgetChange{
				ID:      id,
				Changes: changes,
			})
		}
	}
	return diff
}

func widgetChanges(before WidgetInfo, after WidgetInfo) map[string]ValueChange {
	changes := make(map[string]ValueChange)
	if before.Label != after.Label {
		changes["label"] = ValueChange{Old: before.Label, New: after.Label}
	}
	if before.Value != after.Value {
		changes["value"] = ValueChange{Old: before.Value, New: after.Value}
	}
	if before.Description != after.Description {
		changes["description"] = ValueChange{Old: before.Description, New: after.Description}
	}
	if before.Role != after.Role {
		changes["role"] = ValueChange{Old: before.Role, New: after.Role}
	}
	if before.Bounds != after.Bounds {
		changes["bounds"] = ValueChange{Old: before.Bounds, New: after.Bounds}
	}
	if !equalStringSlice(before.Actions, after.Actions) {
		changes["actions"] = ValueChange{Old: before.Actions, New: after.Actions}
	}
	if !equalStringSlice(before.ChildrenIDs, after.ChildrenIDs) {
		changes["children_ids"] = ValueChange{Old: before.ChildrenIDs, New: after.ChildrenIDs}
	}
	if before.ParentID != after.ParentID {
		changes["parent_id"] = ValueChange{Old: before.ParentID, New: after.ParentID}
	}
	stateChanges := stateDiff(before.State, after.State)
	for key, change := range stateChanges {
		changes["state."+key] = change
	}
	return changes
}

func stateDiff(before StateSet, after StateSet) map[string]ValueChange {
	changes := map[string]ValueChange{}
	if before.Focused != after.Focused {
		changes["focused"] = ValueChange{Old: before.Focused, New: after.Focused}
	}
	if before.Disabled != after.Disabled {
		changes["disabled"] = ValueChange{Old: before.Disabled, New: after.Disabled}
	}
	if before.Selected != after.Selected {
		changes["selected"] = ValueChange{Old: before.Selected, New: after.Selected}
	}
	if before.ReadOnly != after.ReadOnly {
		changes["readonly"] = ValueChange{Old: before.ReadOnly, New: after.ReadOnly}
	}
	if before.Required != after.Required {
		changes["required"] = ValueChange{Old: before.Required, New: after.Required}
	}
	if before.Invalid != after.Invalid {
		changes["invalid"] = ValueChange{Old: before.Invalid, New: after.Invalid}
	}
	if !boolPtrEqual(before.Checked, after.Checked) {
		changes["checked"] = ValueChange{Old: before.Checked, New: after.Checked}
	}
	if !boolPtrEqual(before.Expanded, after.Expanded) {
		changes["expanded"] = ValueChange{Old: before.Expanded, New: after.Expanded}
	}
	return changes
}

func boolPtrEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
