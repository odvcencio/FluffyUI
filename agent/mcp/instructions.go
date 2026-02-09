package mcp

// AgentInstructions contains comprehensive guidance for AI agents interacting
// with FluffyUI applications through the MCP server. It is served both as the
// MCP server's instruction string and as the fluffy://instructions resource.
const AgentInstructions = `# FluffyUI MCP Agent Guide

You are connected to a FluffyUI terminal application through its MCP (Model
Context Protocol) server. This guide teaches you how to efficiently observe,
navigate, and interact with the application's widget tree.

## Quick Start

1. Call get_app_info to learn the screen dimensions, layer count, and uptime.
2. Call find_all or snapshot to see every widget currently rendered.
3. Call find_focused to see which widget has keyboard focus.
4. Call snapshot_text (if text access is enabled) for a visual overview of the
   screen as plain text.

## Core Concepts

### Widget Tree

The UI is a tree of widgets. Every widget has:
- id        — stable identifier in the form "layer0:type:index:explicitID"
- role      — WAI-ARIA role string (see Role Reference below)
- label     — human-readable name (may be empty for layout containers)
- value     — current content (text input contents, slider position, etc.)
- state     — boolean flags: focused, disabled, hidden, checked, selected,
              expanded, pressed, readonly, required, invalid, busy, modal,
              multiline, multiselectable
- actions   — list of supported actions: "activate", "focus", "toggle", etc.
- bounds    — screen rectangle {x, y, width, height}
- parent_id — ID of the parent widget
- children_ids — IDs of direct children

### Layers

FluffyUI supports modal layers. Layer 0 is the base application. When a
dialog or popover opens, it creates layer 1 (or higher). Use get_layer_count
to check how many layers are active. Widget IDs encode their layer:
"layer0:button:0:save" is on layer 0, "layer1:dialog:0:confirm" is on layer 1.
Always interact with the topmost layer first.

### ARIA Properties

Widgets expose WAI-ARIA 1.2/1.3 properties for rich semantic information:
- live / relevant / atomic — live region announcements
- landmark — navigation landmark (main, nav, complementary, etc.)
- labelled_by / described_by — relationships to other widgets
- controls / owns / flow_to — structural relationships
- level — heading level or tree depth
- orientation — "horizontal" or "vertical"
- active_descendant — ID of the logically active child
- pos_in_set / set_size — position in an ordered set (e.g., tab 2 of 5)
- has_popup — "menu", "listbox", "dialog", etc.
- error_message — validation error text
- placeholder — hint text for empty inputs
- sort — column sort direction ("ascending", "descending", "none")
- key_shortcuts — keyboard shortcut string
- value_info — {min, max, current, text} for range-based widgets

## Tool Categories

### 1. Discovery & Observation

Start broad, then narrow down. These tools are read-only.

| Tool | Purpose | Parameters |
|------|---------|------------|
| get_app_info | App dimensions, layer count, uptime | (none) |
| get_capabilities | Server feature flags | (none) |
| get_dimensions | Screen width and height | (none) |
| get_layer_count | Number of modal layers | (none) |
| snapshot | Full widget tree with metadata | include_text (optional) |
| snapshot_text | Raw screen text (requires AllowText) | (none) |
| snapshot_region | Text from a screen rectangle | x, y, width, height |
| get_cell | Character and style at coordinates | x, y |

### 2. Finding Widgets

Use these to locate specific widgets without fetching the entire tree.

| Tool | Purpose | Parameters |
|------|---------|------------|
| find_by_role | All widgets of a given role | role |
| find_by_label | Widget by label substring | label, index?, layer? |
| find_by_id | Widget by exact ID | id |
| find_by_value | Widgets matching a value substring | value |
| find_by_state | Widgets matching state flags | state (object) |
| find_at_position | Widget at screen coordinates | x, y |
| find_focused | Currently focused widget | (none) |
| find_all | All widgets (can be large) | (none) |
| find_focusable | All focusable widgets | (none) |
| find_actionable | All widgets with actions | (none) |

Best practices:
- Prefer find_by_role over find_all when you know what you're looking for.
- Use find_by_label when you know the visible text of a widget.
- Use find_by_state with {"state": {"disabled": true}} to find disabled widgets.
- find_by_label supports an "index" param to disambiguate multiple matches.
- find_by_label supports a "layer" param to search only a specific modal layer.

### 3. Tree Traversal

Navigate the widget hierarchy starting from a known widget ID.

| Tool | Purpose | Parameters |
|------|---------|------------|
| get_children | Direct children | id |
| get_parent | Parent widget | id |
| get_siblings | Sibling widgets | id |
| get_descendants | All descendants (recursive) | id |
| get_ancestors | All ancestors up to root | id |
| get_next_focusable | Next widget in focus order | id? |
| get_prev_focusable | Previous widget in focus order | id? |

### 4. Widget Properties

Read individual properties of a widget by ID.

| Tool | Purpose | Parameters |
|------|---------|------------|
| get_label | Widget label text | id |
| get_role | Widget role string | id |
| get_value | Widget value | id |
| get_description | Widget description | id |
| get_bounds | Widget screen rectangle | id |
| get_state | Full state object | id |
| get_actions | List of supported actions | id |
| is_focused | Focus check | id |
| is_enabled | Enabled check | id |
| is_visible | Visibility check | id |
| is_checked | Checked check | id |
| is_expanded | Expanded check | id |
| is_selected | Selected check | id |
| has_focus | Whether any widget is focused | (none) |

### 5. Widget Actions

Interact with widgets. Most accept either "label" or "id" to target the widget.

| Tool | Purpose | Parameters |
|------|---------|------------|
| activate | Trigger a widget (click button, follow link) | label/id, index?, layer? |
| focus | Move focus to a widget | label/id, index?, layer? |
| blur | Remove focus | (none) |
| type_into | Type text into an input | label/id, text, index?, layer? |
| clear | Clear input contents | label/id, index?, layer? |
| select_option | Choose a dropdown option by label | label/id, option |
| select_index | Choose a dropdown option by index | label/id, index |
| toggle | Toggle a checkbox/switch | label/id, index?, layer? |
| check | Set checkbox to checked | label/id, index?, layer? |
| uncheck | Set checkbox to unchecked | label/id, index?, layer? |
| expand | Expand a collapsible | label/id, index?, layer? |
| collapse | Collapse a collapsible | label/id, index?, layer? |
| scroll_to | Scroll to make a widget visible | label/id, index?, layer? |
| scroll_by | Scroll by a delta amount | label/id, delta |
| scroll_to_top | Scroll to top | label/id, index?, layer? |
| scroll_to_bottom | Scroll to bottom | label/id, index?, layer? |
| click_widget | Click a widget by label | label/id, index?, layer? |

### 6. Direct Mutation

Set widget state directly without simulating input events.

| Tool | Purpose | Parameters |
|------|---------|------------|
| set_widget_value | Set value directly (text, number, bool) | widget_id, value |
| focus_widget | Move focus by widget ID | id |
| send_key | Send key event to focused or specific widget | key, widget_id? |

### 7. Keyboard Input

Simulate keyboard events. These go to the currently focused widget.

Common keys:
| Tool | Key |
|------|-----|
| press_enter | Enter/Return |
| press_escape | Escape |
| press_tab | Tab (next focus) |
| press_shift_tab | Shift+Tab (prev focus) |
| press_space | Space (toggle, activate) |
| press_backspace | Backspace |
| press_delete | Delete |

Arrow keys:
| Tool | Key |
|------|-----|
| press_up | Up arrow |
| press_down | Down arrow |
| press_left | Left arrow |
| press_right | Right arrow |
| press_home | Home |
| press_end | End |
| press_page_up | Page Up |
| press_page_down | Page Down |

Function keys: press_f1 through press_f12.

Advanced:
| Tool | Purpose | Parameters |
|------|---------|------------|
| press_key | Any named key | key |
| press_keys | Sequence of keys | keys (space-separated) |
| press_chord | Key combination | chord (e.g., "Ctrl+S") |
| press_rune | Single character | rune |
| type_string | Type a full string | text |

### 8. Mouse Input

Simulate mouse events at screen coordinates.

| Tool | Purpose | Parameters |
|------|---------|------------|
| mouse_click | Left click | x, y |
| mouse_double_click | Double click | x, y |
| mouse_right_click | Right click | x, y |
| mouse_press | Button press (hold) | x, y, button |
| mouse_release | Button release | x, y, button |
| mouse_move | Move cursor | x, y |
| mouse_drag | Drag from point to point | x1, y1, x2, y2 |
| mouse_scroll_up | Scroll up | x, y, amount? |
| mouse_scroll_down | Scroll down | x, y, amount? |

### 9. Text Selection & Clipboard

| Tool | Purpose | Parameters |
|------|---------|------------|
| select_all | Select all text | (none) |
| select_range | Select by character range | start, end |
| select_word | Select current word | (none) |
| select_line | Select current line | (none) |
| select_none | Clear selection | (none) |
| get_selection | Get selected text | (none) |
| get_selection_bounds | Get selection start/end | (none) |
| has_selection | Check if text is selected | (none) |
| copy | Copy selection to clipboard | (none) |
| cut | Cut selection to clipboard | (none) |
| paste | Paste from clipboard | (none) |
| paste_text | Paste specific text | text |
| clipboard_read | Read clipboard | (none) |
| clipboard_write | Write to clipboard | text |
| clipboard_clear | Clear clipboard | (none) |
| clipboard_has_text | Check clipboard content | (none) |
| clipboard_read_primary | Read primary selection | (none) |
| clipboard_write_primary | Write to primary selection | text |

### 10. Cursor Control

For text editing widgets with a cursor.

| Tool | Purpose | Parameters |
|------|---------|------------|
| get_cursor_position | Cursor x,y position | (none) |
| set_cursor_position | Move cursor to x,y | x, y |
| get_cursor_offset | Cursor character offset | (none) |
| set_cursor_offset | Move cursor to offset | offset |
| cursor_to_start | Move to beginning | (none) |
| cursor_to_end | Move to end | (none) |
| cursor_word_left | Move one word left | (none) |
| cursor_word_right | Move one word right | (none) |

### 11. Waiting & Synchronization

After performing actions, the UI may need time to update. Use these tools to
wait for specific conditions before reading state.

| Tool | Purpose | Parameters |
|------|---------|------------|
| tick | Wait one render cycle | (none) |
| wait_ms | Wait specific duration | ms |
| wait_for_widget | Wait for widget to appear | label, timeout_ms? |
| wait_for_widget_gone | Wait for widget to disappear | label, timeout_ms? |
| wait_for_text | Wait for text on screen | text, timeout_ms? |
| wait_for_text_gone | Wait for text to disappear | text, timeout_ms? |
| wait_for_focus | Wait for focus on widget | label, timeout_ms? |
| wait_for_value | Wait for widget to have value | label, value, timeout_ms? |
| wait_for_enabled | Wait for widget to be enabled | label, timeout_ms? |
| wait_for_idle | Wait for no pending updates | timeout_ms? |

### 12. Change Detection

Detect what changed between two points in time.

| Tool | Purpose | Parameters |
|------|---------|------------|
| diff_snapshots | Full diff between two snapshots | before, after |
| widgets_changed | Check if widgets changed since snapshot | since |
| text_changed | Check if screen text changed since snapshot | since |

Workflow: call snapshot → perform actions → call snapshot again → call
diff_snapshots with the before and after snapshots to see exactly what changed.

### 13. Screen Management

| Tool | Purpose | Parameters |
|------|---------|------------|
| resize | Set screen dimensions | width, height |
| resize_width | Set screen width only | width |
| resize_height | Set screen height only | height |

### 14. Batch Execution

| Tool | Purpose | Parameters |
|------|---------|------------|
| batch_execute | Run multiple tools sequentially | operations[] |

Each operation is {"tool": "tool_name", "params": {...}}. Up to 50 operations
per batch. Recursive batch_execute is not allowed.

### 15. Server Info

| Tool | Purpose | Parameters |
|------|---------|------------|
| get_app_info | App dimensions, layers, uptime | (none) |
| get_capabilities | Server feature flags | (none) |
| ping | Server health check | (none) |

## Resources (Subscriptions)

Resources provide live data via URI. Subscribe to get notifications on change.

| URI | Description | MIME |
|-----|-------------|------|
| fluffy://screen | Current screen text | text/plain |
| fluffy://widgets | Widget tree snapshot | application/json |
| fluffy://focused | Focused widget details | application/json |
| fluffy://clipboard | Clipboard contents | text/plain |
| fluffy://dimensions | Screen dimensions | application/json |
| fluffy://instructions | This guide | text/plain |
| fluffy://widget/{id} | Widget details by ID | application/json |
| fluffy://widget/{id}/value | Widget value by ID | application/json |
| fluffy://widget/{id}/children | Widget children by ID | application/json |
| fluffy://layer/{n} | Widget tree for layer n | application/json |

## Common Workflows

### Filling Out a Form

1. find_by_role("textbox") to discover all text inputs.
2. For each input: focus_widget(id) then type_into(id, "value").
   Or use set_widget_value(widget_id, "value") for direct mutation.
3. find_by_role("button") to find the submit button.
4. activate(id) to submit.
5. wait_for_idle() then check results.

Batch version:
  batch_execute({operations: [
    {tool: "focus_widget", params: {id: "layer0:textbox:0:name"}},
    {tool: "type_into", params: {id: "layer0:textbox:0:name", text: "Alice"}},
    {tool: "focus_widget", params: {id: "layer0:textbox:1:email"}},
    {tool: "type_into", params: {id: "layer0:textbox:1:email", text: "alice@example.com"}},
    {tool: "activate", params: {id: "layer0:button:0:submit"}}
  ]})

### Navigating Tabs

1. find_by_role("tab") to see all tabs.
2. Each tab shows pos_in_set and set_size (e.g., tab 2 of 5).
3. activate(id) on the desired tab.
4. The corresponding tabpanel will appear — use find_by_role("tabpanel") or
   get_children on the tablist parent to find it.

### Working with Tables

1. find_by_role("table") to find the table.
2. get_children(id) to get rows.
3. For each row, get_children(row_id) to get cells.
4. Cell values are in the "value" or "label" field.
5. For sortable columns, check the "sort" property on columnheader widgets.

### Handling Dialogs and Modals

1. get_layer_count to detect modal layers.
2. If layer_count > 1, a modal is open. Widgets on the top layer have IDs
   starting with "layer1:" (or higher).
3. find_by_role("dialog") to locate the dialog.
4. get_descendants(dialog_id) to see dialog contents.
5. Interact with dialog widgets normally.
6. press_escape or activate the close/cancel button to dismiss.
7. wait_for_widget_gone(dialog_label) to confirm dismissal.

### Selecting from Dropdowns

1. find_by_role("listbox") or find_by_role("combobox") to find selects.
2. select_option(id, "Option Label") to choose by text.
3. Or select_index(id, 2) to choose by position (0-based).
4. For autocomplete/combobox: focus → type_into to filter → select_option.

### Tree Navigation

1. find_by_role("tree") to find the tree widget.
2. get_children(tree_id) for top-level items (role: "treeitem").
3. expand(item_id) to expand a node.
4. get_children(item_id) to see child nodes.
5. activate(item_id) to select a node.

### Toggling Checkboxes and Switches

1. find_by_role("checkbox") or find_by_role("switch").
2. toggle(id) to flip the state.
3. Or check(id) / uncheck(id) for explicit state.
4. is_checked(id) to verify.

### Text Editing

1. find_by_role("textbox") and check state.multiline for multiline editors.
2. focus_widget(id) to focus.
3. type_into(id, text) to type.
4. Use cursor tools to position within existing text.
5. select_range(start, end) then cut/copy/paste for editing.
6. clear(id) to empty the field.

### Reading Progress

1. find_by_role("progressbar") to find progress indicators.
2. Check value_info for {min, max, current} values.
3. Or read the "value" field for a text representation.

### Detecting Changes After Actions

1. Call snapshot to capture the "before" state.
2. Perform your actions.
3. Call wait_for_idle to let the UI settle.
4. Call snapshot again for the "after" state.
5. Call diff_snapshots(before, after) to see exactly what changed.
   The diff includes: widgets_added, widgets_removed, widgets_modified,
   focus_changed, text_changed, dimensions_changed, layer_count_changed.

## Role Reference

Roles map to widget types. Use these strings with find_by_role.

### Interactive Widgets
| Role | FluffyUI Widget |
|------|----------------|
| button | Button |
| textbox | Input, TextArea |
| checkbox | Checkbox |
| radio | Radio button |
| switch | Toggle switch |
| slider | Slider |
| spinbutton | Stepper/NumericInput |
| combobox | AutoComplete, SelectDropdown |
| listbox | Select, MultiSelect, SelectDropdown |
| option | Individual option in a listbox |
| link | Link |
| menuitem | Menu item |
| menuitemcheckbox | Checkable menu item |
| menuitemradio | Radio menu item |
| tab | Tab header |

### Container Widgets
| Role | FluffyUI Widget |
|------|----------------|
| group | Panel, FieldSet, Group |
| tablist | Tab bar |
| tabpanel | Tab content area |
| menu | Menu |
| menubar | Menu bar |
| toolbar | Toolbar |
| dialog | Dialog |
| alertdialog | Alert dialog (requires acknowledgment) |
| tree | Tree |
| treeitem | Tree node |
| treegrid | Tree-table hybrid |
| table | Table |
| row | Table row |
| cell | Table cell |
| gridcell | Interactive table cell |
| columnheader | Table column header |
| rowheader | Table row header |
| rowgroup | Table row group (thead, tbody, tfoot) |
| radiogroup | Group of radio buttons |
| grid | Interactive grid |
| list | List container |
| listitem | List item |

### Status & Feedback
| Role | FluffyUI Widget |
|------|----------------|
| alert | Alert banner |
| status | Status bar |
| progressbar | Progress bar |
| meter | Meter/gauge |
| log | Log viewer |
| marquee | Scrolling content |
| timer | Timer display |
| tooltip | Tooltip |

### Structure & Landmark
| Role | FluffyUI Widget |
|------|----------------|
| application | Root application |
| window | Window |
| document | Document content area |
| heading | Section heading |
| separator | Divider/separator |
| text | Static text |
| img / image | Image display |
| navigation | Navigation landmark |
| complementary | Complementary landmark |
| note | Note/aside |
| feed | Feed of articles |
| searchbox | Search input |
| scrollbar | Scrollbar |

### Semantic/Annotation
| Role | FluffyUI Widget |
|------|----------------|
| code | Code block |
| mark | Highlighted text |
| comment | Comment |
| suggestion | Suggested edit |
| time | Time display |
| presentation | Decorative (no semantics) |
| none | No role |
| chart | Chart/graph |

## State Flags Reference

The state object returned by get_state has these boolean fields:

| Flag | Meaning |
|------|---------|
| focused | Widget has keyboard focus |
| disabled | Widget is non-interactive |
| hidden | Widget is not visible |
| checked | Checkbox/radio is checked (null = not a checkable widget) |
| selected | Item is selected |
| expanded | Collapsible is open (null = not expandable) |
| pressed | Toggle button is pressed (null = not a toggle) |
| readonly | Content cannot be edited |
| required | Input must have a value |
| invalid | Input has a validation error |
| busy | Widget is loading/processing |
| modal | Widget blocks interaction with other content |
| multiline | Text input accepts multiple lines |
| multiselectable | List allows multiple selection |

## Efficiency Guidelines

1. Start with get_app_info, not find_all. Get the lay of the land first.

2. Use find_by_role for targeted discovery. Searching by role is faster and
   returns less data than find_all or snapshot.

3. Use batch_execute for multi-step operations. One round trip for multiple
   actions reduces latency significantly.

4. Wait smartly. After actions that trigger UI updates (activate, type_into,
   toggle), call wait_for_idle or the specific wait_for_* tool before reading
   state. Don't use wait_ms with arbitrary delays.

5. Use diff_snapshots for precise change detection. Instead of re-scanning the
   entire UI, diff two snapshots to see exactly what changed.

6. Prefer ID-based targeting. Once you have a widget's ID from a find_* call,
   use that ID for subsequent interactions. IDs are stable within a session.

7. Use resources for monitoring. Subscribe to fluffy://focused to get notified
   when focus changes, or fluffy://widgets to track all widget tree changes.

8. Read the actions list. Before trying to interact with a widget, check its
   "actions" field. If "activate" is not listed, the widget can't be activated.
   If "focus" is not listed, it can't receive focus.

9. Respect layers. When a modal dialog is open, interact with the topmost
   layer. Check get_layer_count and target widgets on that layer.

10. Use value_info for range widgets. Sliders, progress bars, and meters have
    a value_info object with {min, max, current} that is more precise than
    the text "value" field.

## Error Handling

- "widget not found" — the widget ID does not exist. It may have been removed
  or the ID changed (e.g., after a re-render). Re-discover with find_by_role
  or find_by_label.
- "text access denied" — the server has AllowText disabled. Tools that read
  screen text (snapshot_text, find_by_label, find_by_value) are unavailable.
  Use find_by_role or snapshot (without include_text) instead.
- "clipboard access denied" — AllowClipboard is disabled.
- "rate limit exceeded" — slow down. Check the retry_after_ms field.
- "session expired" — reconnect.
- "authentication required" — provide a token in the initialize request.

## Access Control

Some tools require opt-in permissions:
- Text access (AllowText): snapshot_text, snapshot_region, find_by_label,
  find_by_value, snapshot with include_text, fluffy://screen resource.
- Clipboard access (AllowClipboard): clipboard_read, clipboard_write, and
  other clipboard_* tools, fluffy://clipboard resource.

If you get an access denied error, note which features are unavailable and
adapt your strategy to use alternative tools.
`
