package main

import (
	"fmt"
	"strings"

	mcp "m31labs.dev/fluffyui/agent/mcp"
)

// Announce generates utterances from snapshot changes using ARIA semantics.
// No hardcoded role checks — everything is driven by widget metadata.
func Announce(before, after mcp.Snapshot, diff mcp.Diff) []Utterance {
	var out []Utterance

	beforeIndex := indexWidgets(before.Widgets)
	afterIndex := indexWidgets(after.Widgets)

	// 1. Focus changed → announce the newly focused widget
	if diff.FocusChanged && after.FocusedID != "" {
		if w, ok := afterIndex[after.FocusedID]; ok {
			if !w.State.Hidden {
				text := formatWidget(w)
				if text != "" {
					out = append(out, Utterance{Text: text, Priority: "assertive"})
				}
			}
		}
	}

	// 2. Modified widgets in live regions
	for _, mod := range diff.WidgetsModified {
		w, ok := afterIndex[mod.ID]
		if !ok {
			continue
		}
		if w.Live == "" {
			continue
		}
		if w.State.Busy {
			continue
		}
		if !isRelevantChange(w.Relevant, mod.Changes) {
			continue
		}
		priority := liveToPriority(w.Live)
		if w.Atomic {
			text := formatWidget(w)
			if text != "" {
				out = append(out, Utterance{Text: text, Priority: priority})
			}
		} else {
			text := formatChanges(mod.Changes)
			if text != "" {
				out = append(out, Utterance{Text: text, Priority: priority})
			}
		}
	}

	// 3. New widgets added to live regions
	for _, id := range diff.WidgetsAdded {
		w, ok := afterIndex[id]
		if !ok {
			continue
		}
		// Check if the widget itself is a live region
		if w.Live != "" && isRelevantType(w.Relevant, "additions") {
			text := formatWidget(w)
			if text != "" {
				out = append(out, Utterance{Text: text, Priority: liveToPriority(w.Live)})
			}
			continue
		}
		// Check parent
		if w.ParentID != "" {
			if parent, ok := afterIndex[w.ParentID]; ok {
				if parent.Live != "" && isRelevantType(parent.Relevant, "additions") {
					text := formatWidget(w)
					if text != "" {
						out = append(out, Utterance{Text: text, Priority: liveToPriority(parent.Live)})
					}
				}
			}
		}
	}

	// 4. Widgets removed from live regions
	for _, id := range diff.WidgetsRemoved {
		w, ok := beforeIndex[id]
		if !ok {
			continue
		}
		// Check if the widget itself was in a live region
		if w.Live != "" && isRelevantType(w.Relevant, "removals") {
			label := w.Label
			if label == "" {
				label = w.Role
			}
			if label != "" {
				out = append(out, Utterance{
					Text:     fmt.Sprintf("Removed %s", label),
					Priority: liveToPriority(w.Live),
				})
			}
			continue
		}
		// Check parent
		if w.ParentID != "" {
			if parent, ok := beforeIndex[w.ParentID]; ok {
				if parent.Live != "" && isRelevantType(parent.Relevant, "removals") {
					label := w.Label
					if label == "" {
						label = w.Role
					}
					if label != "" {
						out = append(out, Utterance{
							Text:     fmt.Sprintf("Removed %s", label),
							Priority: liveToPriority(parent.Live),
						})
					}
				}
			}
		}
	}

	return out
}

// formatWidget builds a speech string: "label, role[, states][, value]"
func formatWidget(w mcp.WidgetInfo) string {
	// Skip hidden widgets entirely
	if w.State.Hidden {
		return ""
	}
	var parts []string

	// Resolve LabelledBy if present
	label := strings.TrimSpace(w.Label)
	if label != "" {
		parts = append(parts, label)
	}

	role := strings.TrimSpace(w.Role)
	if role != "" {
		parts = append(parts, role)
	}

	desc := strings.TrimSpace(w.Description)
	if desc != "" {
		parts = append(parts, desc)
	}

	if states := formatStates(w.State); states != "" {
		parts = append(parts, states)
	}

	value := strings.TrimSpace(w.Value)
	if value != "" {
		parts = append(parts, value)
	}

	if w.RoleDescription != "" {
		// Override role name with custom description
		for i, p := range parts {
			if p == role {
				parts[i] = w.RoleDescription
				break
			}
		}
	}
	if w.Level > 0 && role == "heading" {
		parts = append(parts, fmt.Sprintf("level %d", w.Level))
	}
	if w.PosInSet > 0 && w.SetSize > 0 {
		parts = append(parts, fmt.Sprintf("%d of %d", w.PosInSet, w.SetSize))
	}
	if w.Current != "" {
		parts = append(parts, "current "+w.Current)
	}
	if w.Placeholder != "" && value == "" {
		parts = append(parts, w.Placeholder)
	}
	if w.ErrorMessage != "" {
		parts = append(parts, "error: "+w.ErrorMessage)
	}

	return strings.Join(parts, ", ")
}

func formatStates(s mcp.StateSet) string {
	var parts []string
	if s.Checked != nil {
		if *s.Checked {
			parts = append(parts, "checked")
		} else {
			parts = append(parts, "unchecked")
		}
	}
	if s.Expanded != nil {
		if *s.Expanded {
			parts = append(parts, "expanded")
		} else {
			parts = append(parts, "collapsed")
		}
	}
	if s.Pressed != nil {
		if *s.Pressed {
			parts = append(parts, "pressed")
		} else {
			parts = append(parts, "not pressed")
		}
	}
	if s.Selected {
		parts = append(parts, "selected")
	}
	if s.Disabled {
		parts = append(parts, "disabled")
	}
	if s.ReadOnly {
		parts = append(parts, "read-only")
	}
	if s.Busy {
		parts = append(parts, "busy")
	}
	if s.Modal {
		parts = append(parts, "modal dialog")
	}
	if s.Multiline {
		parts = append(parts, "multiline")
	}
	if s.Multiselectable {
		parts = append(parts, "multiselectable")
	}
	return strings.Join(parts, " ")
}

func formatChanges(changes map[string]mcp.ValueChange) string {
	var parts []string
	if c, ok := changes["label"]; ok {
		if new, ok := c.New.(string); ok && new != "" {
			parts = append(parts, new)
		}
	}
	if c, ok := changes["value"]; ok {
		if new, ok := c.New.(string); ok && new != "" {
			parts = append(parts, new)
		}
	}
	if c, ok := changes["description"]; ok {
		if new, ok := c.New.(string); ok && new != "" {
			parts = append(parts, new)
		}
	}
	return strings.Join(parts, ", ")
}

func isRelevantChange(relevant string, changes map[string]mcp.ValueChange) bool {
	if relevant == "" || relevant == "all" {
		return true
	}
	if relevant == "text" {
		_, hasLabel := changes["label"]
		_, hasValue := changes["value"]
		_, hasDesc := changes["description"]
		return hasLabel || hasValue || hasDesc
	}
	// For "additions" or "removals", structural changes are handled separately
	return true
}

func isRelevantType(relevant string, changeType string) bool {
	if relevant == "" || relevant == "all" {
		return true
	}
	return relevant == changeType
}

func liveToPriority(live string) string {
	if live == "assertive" {
		return "assertive"
	}
	return "polite"
}
