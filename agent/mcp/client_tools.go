package mcp

import "time"

type actionTarget struct {
	Label string
	ID    string
	Index *int
	Layer *int
}

func (t actionTarget) args() map[string]any {
	args := map[string]any{}
	if t.Label != "" {
		args["label"] = t.Label
	}
	if t.ID != "" {
		args["id"] = t.ID
	}
	if t.Index != nil {
		args["index"] = *t.Index
	}
	if t.Layer != nil {
		args["layer"] = *t.Layer
	}
	return args
}

func labelQueryPayload(label string, index *int, layer *int) map[string]any {
	args := map[string]any{
		"label": label,
	}
	if index != nil {
		args["index"] = *index
	}
	if layer != nil {
		args["layer"] = *layer
	}
	return args
}

func idPayload(id string) map[string]any {
	return map[string]any{"id": id}
}

// Snapshot & screen.
func (c *Client) Snapshot() (Snapshot, error) {
	return callTool[Snapshot](c, "snapshot", nil)
}

func (c *Client) SnapshotWithText() (Snapshot, error) {
	return callTool[Snapshot](c, "snapshot", map[string]any{"include_text": true})
}

func (c *Client) SnapshotText() (string, error) {
	return callTool[string](c, "snapshot_text", nil)
}

func (c *Client) SnapshotRegion(x, y, width, height int) (string, error) {
	return callTool[string](c, "snapshot_region", map[string]any{
		"x":      x,
		"y":      y,
		"width":  width,
		"height": height,
	})
}

func (c *Client) GetDimensions() (Dimensions, error) {
	return callTool[Dimensions](c, "get_dimensions", nil)
}

func (c *Client) GetLayerCount() (int, error) {
	return callTool[int](c, "get_layer_count", nil)
}

func (c *Client) GetCell(x, y int) (CellInfo, error) {
	return callTool[CellInfo](c, "get_cell", map[string]any{"x": x, "y": y})
}

// Widget queries.
func (c *Client) FindByLabel(label string) (*WidgetInfo, error) {
	return callTool[*WidgetInfo](c, "find_by_label", labelQueryPayload(label, nil, nil))
}

func (c *Client) FindByLabelAt(label string, index int) (*WidgetInfo, error) {
	return callTool[*WidgetInfo](c, "find_by_label", labelQueryPayload(label, &index, nil))
}

func (c *Client) FindByLabelInLayer(label string, layer int) (*WidgetInfo, error) {
	return callTool[*WidgetInfo](c, "find_by_label", labelQueryPayload(label, nil, &layer))
}

func (c *Client) FindByLabelAtLayer(label string, index, layer int) (*WidgetInfo, error) {
	return callTool[*WidgetInfo](c, "find_by_label", labelQueryPayload(label, &index, &layer))
}

func (c *Client) FindByRole(role string) ([]WidgetInfo, error) {
	return callTool[[]WidgetInfo](c, "find_by_role", map[string]any{"role": role})
}

func (c *Client) FindByID(id string) (*WidgetInfo, error) {
	return callTool[*WidgetInfo](c, "find_by_id", map[string]any{"id": id})
}

func (c *Client) FindByValue(value string) ([]WidgetInfo, error) {
	return callTool[[]WidgetInfo](c, "find_by_value", map[string]any{"value": value})
}

func (c *Client) FindByState(state StateSet) ([]WidgetInfo, error) {
	return callTool[[]WidgetInfo](c, "find_by_state", map[string]any{"state": state})
}

func (c *Client) FindAtPosition(x, y int) (*WidgetInfo, error) {
	return callTool[*WidgetInfo](c, "find_at_position", map[string]any{"x": x, "y": y})
}

func (c *Client) FindFocused() (*WidgetInfo, error) {
	return callTool[*WidgetInfo](c, "find_focused", nil)
}

func (c *Client) FindAll() ([]WidgetInfo, error) {
	return callTool[[]WidgetInfo](c, "find_all", nil)
}

func (c *Client) FindFocusable() ([]WidgetInfo, error) {
	return callTool[[]WidgetInfo](c, "find_focusable", nil)
}

func (c *Client) FindActionable() ([]WidgetInfo, error) {
	return callTool[[]WidgetInfo](c, "find_actionable", nil)
}

// Widget tree navigation.
func (c *Client) GetChildren(id string) ([]WidgetInfo, error) {
	return callTool[[]WidgetInfo](c, "get_children", idPayload(id))
}

func (c *Client) GetParent(id string) (*WidgetInfo, error) {
	return callTool[*WidgetInfo](c, "get_parent", idPayload(id))
}

func (c *Client) GetSiblings(id string) ([]WidgetInfo, error) {
	return callTool[[]WidgetInfo](c, "get_siblings", idPayload(id))
}

func (c *Client) GetDescendants(id string) ([]WidgetInfo, error) {
	return callTool[[]WidgetInfo](c, "get_descendants", idPayload(id))
}

func (c *Client) GetAncestors(id string) ([]WidgetInfo, error) {
	return callTool[[]WidgetInfo](c, "get_ancestors", idPayload(id))
}

func (c *Client) NextFocusable() (*WidgetInfo, error) {
	return callTool[*WidgetInfo](c, "get_next_focusable", nil)
}

func (c *Client) NextFocusableFrom(id string) (*WidgetInfo, error) {
	return callTool[*WidgetInfo](c, "get_next_focusable", idPayload(id))
}

func (c *Client) PrevFocusable() (*WidgetInfo, error) {
	return callTool[*WidgetInfo](c, "get_prev_focusable", nil)
}

func (c *Client) PrevFocusableFrom(id string) (*WidgetInfo, error) {
	return callTool[*WidgetInfo](c, "get_prev_focusable", idPayload(id))
}

// Widget properties.
func (c *Client) GetLabel(id string) (string, error) {
	return callTool[string](c, "get_label", idPayload(id))
}

func (c *Client) GetRole(id string) (string, error) {
	return callTool[string](c, "get_role", idPayload(id))
}

func (c *Client) GetValue(id string) (string, error) {
	return callTool[string](c, "get_value", idPayload(id))
}

func (c *Client) GetDescription(id string) (string, error) {
	return callTool[string](c, "get_description", idPayload(id))
}

func (c *Client) GetBounds(id string) (Rect, error) {
	return callTool[Rect](c, "get_bounds", idPayload(id))
}

func (c *Client) GetState(id string) (StateSet, error) {
	return callTool[StateSet](c, "get_state", idPayload(id))
}

func (c *Client) GetActions(id string) ([]string, error) {
	return callTool[[]string](c, "get_actions", idPayload(id))
}

func (c *Client) IsFocused(id string) (bool, error) {
	return callTool[bool](c, "is_focused", idPayload(id))
}

func (c *Client) IsEnabled(id string) (bool, error) {
	return callTool[bool](c, "is_enabled", idPayload(id))
}

func (c *Client) IsVisible(id string) (bool, error) {
	return callTool[bool](c, "is_visible", idPayload(id))
}

func (c *Client) IsChecked(id string) (bool, error) {
	return callTool[bool](c, "is_checked", idPayload(id))
}

func (c *Client) IsExpanded(id string) (bool, error) {
	return callTool[bool](c, "is_expanded", idPayload(id))
}

func (c *Client) IsSelected(id string) (bool, error) {
	return callTool[bool](c, "is_selected", idPayload(id))
}

func (c *Client) HasFocus() (bool, error) {
	return callTool[bool](c, "has_focus", nil)
}

// Semantic actions.
func (c *Client) Activate(label string) (ActionResult, error) {
	return c.action("activate", actionTarget{Label: label})
}

func (c *Client) ActivateAt(label string, index int) (ActionResult, error) {
	return c.action("activate", actionTarget{Label: label, Index: &index})
}

func (c *Client) ActivateLayer(label string, layer int) (ActionResult, error) {
	return c.action("activate", actionTarget{Label: label, Layer: &layer})
}

func (c *Client) ActivateID(id string) (ActionResult, error) {
	return c.action("activate", actionTarget{ID: id})
}

func (c *Client) Focus(label string) (ActionResult, error) {
	return c.action("focus", actionTarget{Label: label})
}

func (c *Client) FocusAt(label string, index int) (ActionResult, error) {
	return c.action("focus", actionTarget{Label: label, Index: &index})
}

func (c *Client) FocusLayer(label string, layer int) (ActionResult, error) {
	return c.action("focus", actionTarget{Label: label, Layer: &layer})
}

func (c *Client) FocusID(id string) (ActionResult, error) {
	return c.action("focus", actionTarget{ID: id})
}

func (c *Client) Blur() (bool, error) {
	return callTool[bool](c, "blur", nil)
}

func (c *Client) TypeInto(label, text string) (ActionResult, error) {
	return c.typeInto(actionTarget{Label: label}, text)
}

func (c *Client) TypeIntoAt(label string, index int, text string) (ActionResult, error) {
	return c.typeInto(actionTarget{Label: label, Index: &index}, text)
}

func (c *Client) TypeIntoLayer(label string, layer int, text string) (ActionResult, error) {
	return c.typeInto(actionTarget{Label: label, Layer: &layer}, text)
}

func (c *Client) TypeIntoID(id, text string) (ActionResult, error) {
	return c.typeInto(actionTarget{ID: id}, text)
}

func (c *Client) Clear(label string) (ActionResult, error) {
	return c.action("clear", actionTarget{Label: label})
}

func (c *Client) ClearID(id string) (ActionResult, error) {
	return c.action("clear", actionTarget{ID: id})
}

func (c *Client) SelectOption(label, option string) (ActionResult, error) {
	args := actionTarget{Label: label}.args()
	args["option"] = option
	return callTool[ActionResult](c, "select_option", args)
}

func (c *Client) SelectOptionID(id, option string) (ActionResult, error) {
	args := actionTarget{ID: id}.args()
	args["option"] = option
	return callTool[ActionResult](c, "select_option", args)
}

func (c *Client) SelectIndex(label string, index int) (ActionResult, error) {
	args := actionTarget{Label: label}.args()
	args["index"] = index
	return callTool[ActionResult](c, "select_index", args)
}

func (c *Client) SelectIndexID(id string, index int) (ActionResult, error) {
	args := actionTarget{ID: id}.args()
	args["index"] = index
	return callTool[ActionResult](c, "select_index", args)
}

func (c *Client) Toggle(label string) (ActionResult, error) {
	return c.action("toggle", actionTarget{Label: label})
}

func (c *Client) ToggleID(id string) (ActionResult, error) {
	return c.action("toggle", actionTarget{ID: id})
}

func (c *Client) Check(label string) (ActionResult, error) {
	return c.action("check", actionTarget{Label: label})
}

func (c *Client) CheckID(id string) (ActionResult, error) {
	return c.action("check", actionTarget{ID: id})
}

func (c *Client) Uncheck(label string) (ActionResult, error) {
	return c.action("uncheck", actionTarget{Label: label})
}

func (c *Client) UncheckID(id string) (ActionResult, error) {
	return c.action("uncheck", actionTarget{ID: id})
}

func (c *Client) Expand(label string) (ActionResult, error) {
	return c.action("expand", actionTarget{Label: label})
}

func (c *Client) ExpandID(id string) (ActionResult, error) {
	return c.action("expand", actionTarget{ID: id})
}

func (c *Client) Collapse(label string) (ActionResult, error) {
	return c.action("collapse", actionTarget{Label: label})
}

func (c *Client) CollapseID(id string) (ActionResult, error) {
	return c.action("collapse", actionTarget{ID: id})
}

func (c *Client) ScrollTo(label string) (ActionResult, error) {
	return c.action("scroll_to", actionTarget{Label: label})
}

func (c *Client) ScrollToID(id string) (ActionResult, error) {
	return c.action("scroll_to", actionTarget{ID: id})
}

func (c *Client) ScrollBy(label string, delta int) (ActionResult, error) {
	args := actionTarget{Label: label}.args()
	args["delta"] = delta
	return callTool[ActionResult](c, "scroll_by", args)
}

func (c *Client) ScrollByID(id string, delta int) (ActionResult, error) {
	args := actionTarget{ID: id}.args()
	args["delta"] = delta
	return callTool[ActionResult](c, "scroll_by", args)
}

func (c *Client) ScrollToTop(label string) (ActionResult, error) {
	return c.action("scroll_to_top", actionTarget{Label: label})
}

func (c *Client) ScrollToTopID(id string) (ActionResult, error) {
	return c.action("scroll_to_top", actionTarget{ID: id})
}

func (c *Client) ScrollToBottom(label string) (ActionResult, error) {
	return c.action("scroll_to_bottom", actionTarget{Label: label})
}

func (c *Client) ScrollToBottomID(id string) (ActionResult, error) {
	return c.action("scroll_to_bottom", actionTarget{ID: id})
}

// Raw keyboard.
func (c *Client) PressKey(key string) (bool, error) {
	return callTool[bool](c, "press_key", map[string]any{"key": key})
}

func (c *Client) PressKeys(keys string) (bool, error) {
	return callTool[bool](c, "press_keys", map[string]any{"keys": keys})
}

func (c *Client) PressChord(chord string) (bool, error) {
	return callTool[bool](c, "press_chord", map[string]any{"chord": chord})
}

func (c *Client) PressRune(r rune) (bool, error) {
	return callTool[bool](c, "press_rune", map[string]any{"rune": string(r)})
}

func (c *Client) TypeString(text string) (bool, error) {
	return callTool[bool](c, "type_string", map[string]any{"text": text})
}

func (c *Client) PressEnter() (bool, error) {
	return callTool[bool](c, "press_enter", nil)
}

func (c *Client) PressEscape() (bool, error) {
	return callTool[bool](c, "press_escape", nil)
}

func (c *Client) PressTab() (bool, error) {
	return callTool[bool](c, "press_tab", nil)
}

func (c *Client) PressShiftTab() (bool, error) {
	return callTool[bool](c, "press_shift_tab", nil)
}

func (c *Client) PressSpace() (bool, error) {
	return callTool[bool](c, "press_space", nil)
}

func (c *Client) PressBackspace() (bool, error) {
	return callTool[bool](c, "press_backspace", nil)
}

func (c *Client) PressDelete() (bool, error) {
	return callTool[bool](c, "press_delete", nil)
}

func (c *Client) PressUp() (bool, error) {
	return callTool[bool](c, "press_up", nil)
}

func (c *Client) PressDown() (bool, error) {
	return callTool[bool](c, "press_down", nil)
}

func (c *Client) PressLeft() (bool, error) {
	return callTool[bool](c, "press_left", nil)
}

func (c *Client) PressRight() (bool, error) {
	return callTool[bool](c, "press_right", nil)
}

func (c *Client) PressHome() (bool, error) {
	return callTool[bool](c, "press_home", nil)
}

func (c *Client) PressEnd() (bool, error) {
	return callTool[bool](c, "press_end", nil)
}

func (c *Client) PressPageUp() (bool, error) {
	return callTool[bool](c, "press_page_up", nil)
}

func (c *Client) PressPageDown() (bool, error) {
	return callTool[bool](c, "press_page_down", nil)
}

func (c *Client) PressF1() (bool, error)  { return callTool[bool](c, "press_f1", nil) }
func (c *Client) PressF2() (bool, error)  { return callTool[bool](c, "press_f2", nil) }
func (c *Client) PressF3() (bool, error)  { return callTool[bool](c, "press_f3", nil) }
func (c *Client) PressF4() (bool, error)  { return callTool[bool](c, "press_f4", nil) }
func (c *Client) PressF5() (bool, error)  { return callTool[bool](c, "press_f5", nil) }
func (c *Client) PressF6() (bool, error)  { return callTool[bool](c, "press_f6", nil) }
func (c *Client) PressF7() (bool, error)  { return callTool[bool](c, "press_f7", nil) }
func (c *Client) PressF8() (bool, error)  { return callTool[bool](c, "press_f8", nil) }
func (c *Client) PressF9() (bool, error)  { return callTool[bool](c, "press_f9", nil) }
func (c *Client) PressF10() (bool, error) { return callTool[bool](c, "press_f10", nil) }
func (c *Client) PressF11() (bool, error) { return callTool[bool](c, "press_f11", nil) }
func (c *Client) PressF12() (bool, error) { return callTool[bool](c, "press_f12", nil) }

// Raw mouse.
func (c *Client) MouseClick(x, y int) (bool, error) {
	return callTool[bool](c, "mouse_click", map[string]any{"x": x, "y": y})
}

func (c *Client) MouseDoubleClick(x, y int) (bool, error) {
	return callTool[bool](c, "mouse_double_click", map[string]any{"x": x, "y": y})
}

func (c *Client) MouseRightClick(x, y int) (bool, error) {
	return callTool[bool](c, "mouse_right_click", map[string]any{"x": x, "y": y})
}

func (c *Client) MousePress(x, y int, button string) (bool, error) {
	return callTool[bool](c, "mouse_press", map[string]any{"x": x, "y": y, "button": button})
}

func (c *Client) MouseRelease(x, y int, button string) (bool, error) {
	return callTool[bool](c, "mouse_release", map[string]any{"x": x, "y": y, "button": button})
}

func (c *Client) MouseMove(x, y int) (bool, error) {
	return callTool[bool](c, "mouse_move", map[string]any{"x": x, "y": y})
}

func (c *Client) MouseDrag(x1, y1, x2, y2 int) (bool, error) {
	return callTool[bool](c, "mouse_drag", map[string]any{"x1": x1, "y1": y1, "x2": x2, "y2": y2})
}

func (c *Client) MouseScrollUp(x, y, amount int) (bool, error) {
	return callTool[bool](c, "mouse_scroll_up", map[string]any{"x": x, "y": y, "amount": amount})
}

func (c *Client) MouseScrollDown(x, y, amount int) (bool, error) {
	return callTool[bool](c, "mouse_scroll_down", map[string]any{"x": x, "y": y, "amount": amount})
}

func (c *Client) ClickWidget(label string) (ActionResult, error) {
	return c.action("click_widget", actionTarget{Label: label})
}

// Clipboard.
func (c *Client) ClipboardRead() (string, error) {
	return callTool[string](c, "clipboard_read", nil)
}

func (c *Client) ClipboardWrite(text string) (bool, error) {
	return callTool[bool](c, "clipboard_write", map[string]any{"text": text})
}

func (c *Client) ClipboardClear() (bool, error) {
	return callTool[bool](c, "clipboard_clear", nil)
}

func (c *Client) ClipboardHasText() (bool, error) {
	return callTool[bool](c, "clipboard_has_text", nil)
}

func (c *Client) ClipboardReadPrimary() (string, error) {
	return callTool[string](c, "clipboard_read_primary", nil)
}

func (c *Client) ClipboardWritePrimary(text string) (bool, error) {
	return callTool[bool](c, "clipboard_write_primary", map[string]any{"text": text})
}

func (c *Client) SelectAll() (bool, error) {
	return callTool[bool](c, "select_all", nil)
}

func (c *Client) SelectRange(start, end int) (bool, error) {
	return callTool[bool](c, "select_range", map[string]any{"start": start, "end": end})
}

func (c *Client) SelectWord() (bool, error) {
	return callTool[bool](c, "select_word", nil)
}

func (c *Client) SelectLine() (bool, error) {
	return callTool[bool](c, "select_line", nil)
}

func (c *Client) SelectNone() (bool, error) {
	return callTool[bool](c, "select_none", nil)
}

func (c *Client) GetSelection() (string, error) {
	return callTool[string](c, "get_selection", nil)
}

func (c *Client) GetSelectionBounds() (SelectionBounds, error) {
	return callTool[SelectionBounds](c, "get_selection_bounds", nil)
}

func (c *Client) HasSelection() (bool, error) {
	return callTool[bool](c, "has_selection", nil)
}

func (c *Client) Copy() (bool, error) {
	return callTool[bool](c, "copy", nil)
}

func (c *Client) Cut() (bool, error) {
	return callTool[bool](c, "cut", nil)
}

func (c *Client) Paste() (bool, error) {
	return callTool[bool](c, "paste", nil)
}

func (c *Client) PasteText(text string) (bool, error) {
	return callTool[bool](c, "paste_text", map[string]any{"text": text})
}

// Cursor & caret.
func (c *Client) GetCursorPosition() (CursorPosition, error) {
	return callTool[CursorPosition](c, "get_cursor_position", nil)
}

func (c *Client) SetCursorPosition(x, y int) (bool, error) {
	return callTool[bool](c, "set_cursor_position", map[string]any{"x": x, "y": y})
}

func (c *Client) GetCursorOffset() (int, error) {
	return callTool[int](c, "get_cursor_offset", nil)
}

func (c *Client) SetCursorOffset(offset int) (bool, error) {
	return callTool[bool](c, "set_cursor_offset", map[string]any{"offset": offset})
}

func (c *Client) CursorToStart() (bool, error) {
	return callTool[bool](c, "cursor_to_start", nil)
}

func (c *Client) CursorToEnd() (bool, error) {
	return callTool[bool](c, "cursor_to_end", nil)
}

func (c *Client) CursorWordLeft() (bool, error) {
	return callTool[bool](c, "cursor_word_left", nil)
}

func (c *Client) CursorWordRight() (bool, error) {
	return callTool[bool](c, "cursor_word_right", nil)
}

// Wait & sync.
func (c *Client) Tick() (bool, error) {
	return callTool[bool](c, "tick", nil)
}

func (c *Client) WaitMS(ms int) (bool, error) {
	return callTool[bool](c, "wait_ms", map[string]any{"ms": ms})
}

func (c *Client) WaitForWidget(label string, timeout time.Duration) (bool, error) {
	return callTool[bool](c, "wait_for_widget", waitLabelPayload(label, timeout))
}

func (c *Client) WaitForWidgetGone(label string, timeout time.Duration) (bool, error) {
	return callTool[bool](c, "wait_for_widget_gone", waitLabelPayload(label, timeout))
}

func (c *Client) WaitForText(text string, timeout time.Duration) (bool, error) {
	return callTool[bool](c, "wait_for_text", waitTextPayload(text, timeout))
}

func (c *Client) WaitForTextGone(text string, timeout time.Duration) (bool, error) {
	return callTool[bool](c, "wait_for_text_gone", waitTextPayload(text, timeout))
}

func (c *Client) WaitForFocus(label string, timeout time.Duration) (bool, error) {
	return callTool[bool](c, "wait_for_focus", waitLabelPayload(label, timeout))
}

func (c *Client) WaitForValue(label, value string, timeout time.Duration) (bool, error) {
	return callTool[bool](c, "wait_for_value", map[string]any{
		"label":      label,
		"value":      value,
		"timeout_ms": timeoutMs(timeout),
	})
}

func (c *Client) WaitForEnabled(label string, timeout time.Duration) (bool, error) {
	return callTool[bool](c, "wait_for_enabled", waitLabelPayload(label, timeout))
}

func (c *Client) WaitForIdle(timeout time.Duration) (bool, error) {
	return callTool[bool](c, "wait_for_idle", map[string]any{"timeout_ms": timeoutMs(timeout)})
}

// Screen control.
func (c *Client) Resize(width, height int) (bool, error) {
	return callTool[bool](c, "resize", map[string]any{"width": width, "height": height})
}

func (c *Client) ResizeWidth(width int) (bool, error) {
	return callTool[bool](c, "resize_width", map[string]any{"width": width})
}

func (c *Client) ResizeHeight(height int) (bool, error) {
	return callTool[bool](c, "resize_height", map[string]any{"height": height})
}

// Comparison.
func (c *Client) DiffSnapshots(before, after Snapshot) (Diff, error) {
	return callTool[Diff](c, "diff_snapshots", map[string]any{
		"before": before,
		"after":  after,
	})
}

func (c *Client) WidgetsChanged(since Snapshot) (bool, error) {
	return callTool[bool](c, "widgets_changed", map[string]any{"since": since})
}

func (c *Client) TextChanged(since Snapshot) (bool, error) {
	return callTool[bool](c, "text_changed", map[string]any{"since": since})
}

// Meta.
func (c *Client) GetCapabilities() (Capabilities, error) {
	return callTool[Capabilities](c, "get_capabilities", nil)
}

func (c *Client) GetAppInfo() (AppInfo, error) {
	return callTool[AppInfo](c, "get_app_info", nil)
}

func (c *Client) Ping() (bool, error) {
	return callTool[bool](c, "ping", nil)
}

func (c *Client) action(tool string, target actionTarget) (ActionResult, error) {
	return callTool[ActionResult](c, tool, target.args())
}

func (c *Client) typeInto(target actionTarget, text string) (ActionResult, error) {
	args := target.args()
	args["text"] = text
	return callTool[ActionResult](c, "type_into", args)
}

func waitLabelPayload(label string, timeout time.Duration) map[string]any {
	args := map[string]any{"label": label}
	if timeout > 0 {
		args["timeout_ms"] = timeoutMs(timeout)
	}
	return args
}

func waitTextPayload(text string, timeout time.Duration) map[string]any {
	args := map[string]any{"text": text}
	if timeout > 0 {
		args["timeout_ms"] = timeoutMs(timeout)
	}
	return args
}

func timeoutMs(timeout time.Duration) int {
	if timeout <= 0 {
		return 0
	}
	return int(timeout / time.Millisecond)
}
