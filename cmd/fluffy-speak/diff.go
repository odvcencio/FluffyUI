package main

import (
	mcp "github.com/odvcencio/fluffyui/agent/mcp"
)

// diffSnapshots compares two snapshots and returns the changes.
// Ported from agent/mcp/tools.go for local use without a round-trip.
func diffSnapshots(before, after mcp.Snapshot) mcp.Diff {
	diff := mcp.Diff{
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
			diff.WidgetsModified = append(diff.WidgetsModified, mcp.WidgetChange{
				ID:      id,
				Changes: changes,
			})
		}
	}
	return diff
}

func indexWidgets(widgets []mcp.WidgetInfo) map[string]mcp.WidgetInfo {
	index := make(map[string]mcp.WidgetInfo, len(widgets))
	for _, w := range widgets {
		index[w.ID] = w
	}
	return index
}

func widgetChanges(before, after mcp.WidgetInfo) map[string]mcp.ValueChange {
	changes := make(map[string]mcp.ValueChange)
	if before.Label != after.Label {
		changes["label"] = mcp.ValueChange{Old: before.Label, New: after.Label}
	}
	if before.Value != after.Value {
		changes["value"] = mcp.ValueChange{Old: before.Value, New: after.Value}
	}
	if before.Description != after.Description {
		changes["description"] = mcp.ValueChange{Old: before.Description, New: after.Description}
	}
	if before.Role != after.Role {
		changes["role"] = mcp.ValueChange{Old: before.Role, New: after.Role}
	}
	if before.Live != after.Live {
		changes["live"] = mcp.ValueChange{Old: before.Live, New: after.Live}
	}
	if before.Relevant != after.Relevant {
		changes["relevant"] = mcp.ValueChange{Old: before.Relevant, New: after.Relevant}
	}
	if before.Atomic != after.Atomic {
		changes["atomic"] = mcp.ValueChange{Old: before.Atomic, New: after.Atomic}
	}
	if before.Landmark != after.Landmark {
		changes["landmark"] = mcp.ValueChange{Old: before.Landmark, New: after.Landmark}
	}
	if before.Owns != after.Owns {
		changes["owns"] = mcp.ValueChange{Old: before.Owns, New: after.Owns}
	}
	if before.FlowTo != after.FlowTo {
		changes["flow_to"] = mcp.ValueChange{Old: before.FlowTo, New: after.FlowTo}
	}
	if before.Level != after.Level {
		changes["level"] = mcp.ValueChange{Old: before.Level, New: after.Level}
	}
	if before.Orientation != after.Orientation {
		changes["orientation"] = mcp.ValueChange{Old: before.Orientation, New: after.Orientation}
	}
	if before.ActiveDescendant != after.ActiveDescendant {
		changes["active_descendant"] = mcp.ValueChange{Old: before.ActiveDescendant, New: after.ActiveDescendant}
	}
	if before.PosInSet != after.PosInSet {
		changes["pos_in_set"] = mcp.ValueChange{Old: before.PosInSet, New: after.PosInSet}
	}
	if before.SetSize != after.SetSize {
		changes["set_size"] = mcp.ValueChange{Old: before.SetSize, New: after.SetSize}
	}
	if before.HasPopup != after.HasPopup {
		changes["has_popup"] = mcp.ValueChange{Old: before.HasPopup, New: after.HasPopup}
	}
	if before.ErrorMessage != after.ErrorMessage {
		changes["error_message"] = mcp.ValueChange{Old: before.ErrorMessage, New: after.ErrorMessage}
	}
	if before.Current != after.Current {
		changes["current"] = mcp.ValueChange{Old: before.Current, New: after.Current}
	}
	if before.Autocomplete != after.Autocomplete {
		changes["autocomplete"] = mcp.ValueChange{Old: before.Autocomplete, New: after.Autocomplete}
	}
	if before.Placeholder != after.Placeholder {
		changes["placeholder"] = mcp.ValueChange{Old: before.Placeholder, New: after.Placeholder}
	}
	if before.Sort != after.Sort {
		changes["sort"] = mcp.ValueChange{Old: before.Sort, New: after.Sort}
	}
	if before.KeyShortcuts != after.KeyShortcuts {
		changes["key_shortcuts"] = mcp.ValueChange{Old: before.KeyShortcuts, New: after.KeyShortcuts}
	}
	if before.Details != after.Details {
		changes["details"] = mcp.ValueChange{Old: before.Details, New: after.Details}
	}
	if before.RoleDescription != after.RoleDescription {
		changes["role_description"] = mcp.ValueChange{Old: before.RoleDescription, New: after.RoleDescription}
	}
	if !valueInfoMCPEqual(before.ValueInfo, after.ValueInfo) {
		changes["value_info"] = mcp.ValueChange{Old: before.ValueInfo, New: after.ValueInfo}
	}
	stateChanges := stateDiff(before.State, after.State)
	for key, change := range stateChanges {
		changes["state."+key] = change
	}
	return changes
}

func stateDiff(before, after mcp.StateSet) map[string]mcp.ValueChange {
	changes := map[string]mcp.ValueChange{}
	if before.Focused != after.Focused {
		changes["focused"] = mcp.ValueChange{Old: before.Focused, New: after.Focused}
	}
	if before.Disabled != after.Disabled {
		changes["disabled"] = mcp.ValueChange{Old: before.Disabled, New: after.Disabled}
	}
	if before.Selected != after.Selected {
		changes["selected"] = mcp.ValueChange{Old: before.Selected, New: after.Selected}
	}
	if before.ReadOnly != after.ReadOnly {
		changes["readonly"] = mcp.ValueChange{Old: before.ReadOnly, New: after.ReadOnly}
	}
	if before.Required != after.Required {
		changes["required"] = mcp.ValueChange{Old: before.Required, New: after.Required}
	}
	if before.Invalid != after.Invalid {
		changes["invalid"] = mcp.ValueChange{Old: before.Invalid, New: after.Invalid}
	}
	if !boolPtrEqual(before.Checked, after.Checked) {
		changes["checked"] = mcp.ValueChange{Old: before.Checked, New: after.Checked}
	}
	if !boolPtrEqual(before.Expanded, after.Expanded) {
		changes["expanded"] = mcp.ValueChange{Old: before.Expanded, New: after.Expanded}
	}
	if before.Busy != after.Busy {
		changes["busy"] = mcp.ValueChange{Old: before.Busy, New: after.Busy}
	}
	if before.Hidden != after.Hidden {
		changes["hidden"] = mcp.ValueChange{Old: before.Hidden, New: after.Hidden}
	}
	if !boolPtrEqual(before.Pressed, after.Pressed) {
		changes["pressed"] = mcp.ValueChange{Old: before.Pressed, New: after.Pressed}
	}
	if before.Modal != after.Modal {
		changes["modal"] = mcp.ValueChange{Old: before.Modal, New: after.Modal}
	}
	if before.Multiline != after.Multiline {
		changes["multiline"] = mcp.ValueChange{Old: before.Multiline, New: after.Multiline}
	}
	if before.Multiselectable != after.Multiselectable {
		changes["multiselectable"] = mcp.ValueChange{Old: before.Multiselectable, New: after.Multiselectable}
	}
	return changes
}

func valueInfoMCPEqual(a, b *mcp.ValueInfoMCP) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Min == b.Min && a.Max == b.Max && a.Current == b.Current && a.Text == b.Text
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
