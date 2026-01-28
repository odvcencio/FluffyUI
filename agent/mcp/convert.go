package mcp

import (
	"strconv"
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/agent"
	"github.com/odvcencio/fluffyui/runtime"
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
		ID:          info.ID,
		Role:        roleToMCP(info.Role),
		Label:       info.Label,
		Value:       info.Value,
		Description: info.Description,
		Bounds:      rectFromRuntime(info.Bounds),
		State:       stateFromAgent(info),
		Actions:     info.Actions,
		ChildrenIDs: childrenIDs,
		ParentID:    parentID,
	}
}

func widgetNodeFromAgent(info agent.WidgetInfo) WidgetNode {
	children := make([]WidgetNode, 0, len(info.Children))
	for _, child := range info.Children {
		children = append(children, widgetNodeFromAgent(child))
	}
	return WidgetNode{
		ID:          info.ID,
		Role:        roleToMCP(info.Role),
		Label:       info.Label,
		Value:       info.Value,
		Description: info.Description,
		Bounds:      rectFromRuntime(info.Bounds),
		State:       stateFromAgent(info),
		Actions:     info.Actions,
		Children:    children,
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
		Focused:  info.Focused,
		Disabled: state.Disabled,
		Checked:  state.Checked,
		Expanded: state.Expanded,
		Selected: state.Selected,
		ReadOnly: state.ReadOnly,
		Required: state.Required,
		Invalid:  state.Invalid,
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
