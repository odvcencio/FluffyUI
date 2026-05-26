package mcp

import (
	"strconv"
	"strings"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/agent"
	"m31labs.dev/fluffyui/runtime"
)

func snapshotFromAgent(snap agent.Snapshot, includeText bool) Snapshot {
	result := Snapshot{
		Timestamp:  snap.Timestamp,
		Dimensions: Dimensions{Width: snap.Width, Height: snap.Height},
		LayerCount: snap.LayerCount,
		FocusedID:  snap.FocusedID,
	}
	flattened := make([]WidgetInfo, 0, len(snap.Widgets))
	flattenWidgets(snap.Widgets, "", &flattened)
	result.Widgets = flattened
	if includeText {
		result.Text = snap.Text
	}
	return result
}

func treeSnapshotFromAgent(snap agent.Snapshot, includeText bool) TreeSnapshot {
	result := TreeSnapshot{
		Timestamp:  snap.Timestamp,
		Dimensions: Dimensions{Width: snap.Width, Height: snap.Height},
		LayerCount: snap.LayerCount,
		FocusedID:  snap.FocusedID,
	}
	nodes := make([]WidgetNode, 0, len(snap.Widgets))
	for _, widget := range snap.Widgets {
		nodes = append(nodes, widgetNodeFromAgent(widget))
	}
	result.Widgets = nodes
	if includeText {
		result.Text = snap.Text
	}
	return result
}

func flattenWidgets(widgets []agent.WidgetInfo, parentID string, out *[]WidgetInfo) {
	for _, entry := range widgets {
		info := widgetInfoFromAgent(entry, parentID)
		*out = append(*out, info)
		if len(entry.Children) > 0 {
			flattenWidgets(entry.Children, entry.ID, out)
		}
	}
}

func widgetInfoFromAgent(info agent.WidgetInfo, parentID string) WidgetInfo {
	childrenIDs := make([]string, 0, len(info.Children))
	for _, child := range info.Children {
		childrenIDs = append(childrenIDs, child.ID)
	}
	return WidgetInfo{
		ID:               info.ID,
		Role:             roleToMCP(info.Role),
		Label:            info.Label,
		Value:            info.Value,
		Description:      info.Description,
		Bounds:           rectFromRuntime(info.Bounds),
		State:            stateFromAgent(info),
		Actions:          info.Actions,
		ChildrenIDs:      childrenIDs,
		ParentID:         parentID,
		Live:             info.Live,
		Relevant:         info.Relevant,
		Atomic:           info.Atomic,
		Landmark:         info.Landmark,
		LabelledBy:       info.LabelledBy,
		DescribedBy:      info.DescribedBy,
		Controls:         info.Controls,
		Owns:             info.Owns,
		FlowTo:           info.FlowTo,
		Level:            info.Level,
		Orientation:      info.Orientation,
		ActiveDescendant: info.ActiveDescendant,
		PosInSet:         info.PosInSet,
		SetSize:          info.SetSize,
		HasPopup:         info.HasPopup,
		ErrorMessage:     info.ErrorMessage,
		Current:          info.Current,
		Autocomplete:     info.Autocomplete,
		Placeholder:      info.Placeholder,
		Sort:             info.Sort,
		KeyShortcuts:     info.KeyShortcuts,
		Details:          info.Details,
		RoleDescription:  info.RoleDescription,
		ValueInfo:        valueInfoFromAgent(info.ValueInfo),
	}
}

func widgetNodeFromAgent(info agent.WidgetInfo) WidgetNode {
	children := make([]WidgetNode, 0, len(info.Children))
	for _, child := range info.Children {
		children = append(children, widgetNodeFromAgent(child))
	}
	return WidgetNode{
		ID:               info.ID,
		Role:             roleToMCP(info.Role),
		Label:            info.Label,
		Value:            info.Value,
		Description:      info.Description,
		Bounds:           rectFromRuntime(info.Bounds),
		State:            stateFromAgent(info),
		Actions:          info.Actions,
		Children:         children,
		Live:             info.Live,
		Relevant:         info.Relevant,
		Atomic:           info.Atomic,
		Landmark:         info.Landmark,
		LabelledBy:       info.LabelledBy,
		DescribedBy:      info.DescribedBy,
		Controls:         info.Controls,
		Owns:             info.Owns,
		FlowTo:           info.FlowTo,
		Level:            info.Level,
		Orientation:      info.Orientation,
		ActiveDescendant: info.ActiveDescendant,
		PosInSet:         info.PosInSet,
		SetSize:          info.SetSize,
		HasPopup:         info.HasPopup,
		ErrorMessage:     info.ErrorMessage,
		Current:          info.Current,
		Autocomplete:     info.Autocomplete,
		Placeholder:      info.Placeholder,
		Sort:             info.Sort,
		KeyShortcuts:     info.KeyShortcuts,
		Details:          info.Details,
		RoleDescription:  info.RoleDescription,
		ValueInfo:        valueInfoFromAgent(info.ValueInfo),
	}
}

func rectFromRuntime(rect runtime.Rect) Rect {
	return Rect{
		X:      rect.X,
		Y:      rect.Y,
		Width:  rect.Width,
		Height: rect.Height,
	}
}

func stateFromAgent(info agent.WidgetInfo) StateSet {
	state := info.State
	return StateSet{
		Focused:         info.Focused,
		Disabled:        state.Disabled,
		Checked:         state.Checked,
		Expanded:        state.Expanded,
		Pressed:         state.Pressed,
		Selected:        state.Selected,
		ReadOnly:        state.ReadOnly,
		Required:        state.Required,
		Invalid:         state.Invalid,
		Hidden:          state.Hidden,
		Busy:            state.Busy,
		Modal:           state.Modal,
		Multiline:       state.Multiline,
		Multiselectable: state.Multiselectable,
	}
}

func roleToMCP(role accessibility.Role) string {
	switch role {
	case accessibility.RoleButton:
		return "button"
	case accessibility.RoleCheckbox:
		return "checkbox"
	case accessibility.RoleRadio:
		return "radio"
	case accessibility.RoleTextbox:
		return "input"
	case accessibility.RoleList:
		return "list"
	case accessibility.RoleListItem:
		return "listitem"
	case accessibility.RoleTree:
		return "tree"
	case accessibility.RoleTreeItem:
		return "treeitem"
	case accessibility.RoleDialog:
		return "dialog"
	case accessibility.RoleMenu:
		return "menu"
	case accessibility.RoleMenuItem:
		return "menuitem"
	case accessibility.RoleTab:
		return "tab"
	case accessibility.RoleTabPanel:
		return "tabpanel"
	case accessibility.RoleProgressBar:
		return "progressbar"
	case accessibility.RoleStatus:
		return "status"
	case accessibility.RoleAlert:
		return "alert"
	case accessibility.RoleText:
		return "text"
	case accessibility.RoleGroup:
		return "container"
	case accessibility.RoleChart:
		return "chart"
	case accessibility.RoleSlider:
		return "slider"
	case accessibility.RoleTable:
		return "table"
	case accessibility.RoleRow:
		return "row"
	case accessibility.RoleCell:
		return "cell"
	case accessibility.RoleTabList:
		return "tablist"
	case accessibility.RoleWindow:
		return "window"
	case accessibility.RoleApplication:
		return "application"
	case accessibility.RoleCombobox:
		return "combobox"
	case accessibility.RoleSwitch:
		return "switch"
	case accessibility.RoleSpinButton:
		return "spinbutton"
	case accessibility.RoleHeading:
		return "heading"
	case accessibility.RoleLink:
		return "link"
	case accessibility.RoleSeparator:
		return "separator"
	case accessibility.RoleLog:
		return "log"
	case accessibility.RoleTimer:
		return "timer"
	case accessibility.RoleFeed:
		return "feed"
	case accessibility.RoleToolbar:
		return "toolbar"
	case accessibility.RoleSearchbox:
		return "searchbox"
	case accessibility.RoleNone:
		return "none"
	case accessibility.RoleImg:
		return "img"
	case accessibility.RoleNote:
		return "note"
	case accessibility.RoleScrollbar:
		return "scrollbar"
	case accessibility.RoleListbox:
		return "listbox"
	case accessibility.RoleOption:
		return "option"
	case accessibility.RoleRadioGroup:
		return "radiogroup"
	case accessibility.RoleGrid:
		return "grid"
	case accessibility.RoleGridCell:
		return "gridcell"
	case accessibility.RoleColumnHeader:
		return "columnheader"
	case accessibility.RoleRowHeader:
		return "rowheader"
	case accessibility.RoleRowGroup:
		return "rowgroup"
	case accessibility.RoleAlertDialog:
		return "alertdialog"
	case accessibility.RoleMenuItemCheckbox:
		return "menuitemcheckbox"
	case accessibility.RoleMenuItemRadio:
		return "menuitemradio"
	case accessibility.RoleMenuBar:
		return "menubar"
	case accessibility.RoleTreeGrid:
		return "treegrid"
	case accessibility.RoleDocument:
		return "document"
	case accessibility.RoleMarquee:
		return "marquee"
	case accessibility.RolePresentation:
		return "presentation"
	case accessibility.RoleComment:
		return "comment"
	case accessibility.RoleMark:
		return "mark"
	case accessibility.RoleSuggestion:
		return "suggestion"
	case accessibility.RoleCode:
		return "code"
	case accessibility.RoleTime:
		return "time"
	case accessibility.RoleImage:
		return "img"
	case accessibility.RoleTooltip:
		return "tooltip"
	case accessibility.RoleMeter:
		return "meter"
	default:
		return "unknown"
	}
}

func layerFromID(id string) int {
	if !strings.HasPrefix(id, "layer") {
		return 0
	}
	parts := strings.SplitN(id, ":", 2)
	if len(parts) == 0 {
		return 0
	}
	layerPart := strings.TrimPrefix(parts[0], "layer")
	layer, err := strconv.Atoi(layerPart)
	if err != nil {
		return 0
	}
	return layer
}

func explicitIDFromWidgetID(id string) string {
	parts := strings.Split(id, ":")
	if len(parts) < 4 {
		return ""
	}
	return parts[len(parts)-1]
}

func explicitBaseID(explicit string) string {
	explicit = strings.TrimSpace(explicit)
	if explicit == "" {
		return ""
	}
	if idx := strings.LastIndex(explicit, "#"); idx > 0 {
		return explicit[:idx]
	}
	return explicit
}

func valueInfoFromAgent(vi *accessibility.ValueInfo) *ValueInfoMCP {
	if vi == nil {
		return nil
	}
	return &ValueInfoMCP{
		Min:     vi.Min,
		Max:     vi.Max,
		Current: vi.Current,
		Text:    vi.Text,
	}
}
