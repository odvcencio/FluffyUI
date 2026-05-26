//go:build !js

package agent

import (
	"time"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/runtime"
)

// Snapshot captures a structured view of the current UI state.
type Snapshot struct {
	Timestamp  time.Time    `json:"timestamp"`
	Width      int          `json:"width"`
	Height     int          `json:"height"`
	LayerCount int          `json:"layer_count,omitempty"`
	Text       string       `json:"text,omitempty"`
	Widgets    []WidgetInfo `json:"widgets,omitempty"`
	FocusedID  string       `json:"focused_id,omitempty"`
	Focused    *WidgetInfo  `json:"focused,omitempty"`
}

// SnapshotOptions configures snapshot output.
type SnapshotOptions struct {
	IncludeText bool
}

// WidgetInfo describes a widget in the UI tree.
type WidgetInfo struct {
	ID          string                   `json:"id"`
	Role        accessibility.Role       `json:"type"`
	Label       string                   `json:"label,omitempty"`
	Description string                   `json:"description,omitempty"`
	Value       string                   `json:"value,omitempty"`
	ValueInfo   *accessibility.ValueInfo `json:"value_info,omitempty"`
	State       accessibility.StateSet   `json:"state,omitempty"`
	Bounds      runtime.Rect             `json:"bounds"`
	Children    []WidgetInfo             `json:"children,omitempty"`
	Actions     []string                 `json:"actions,omitempty"`
	Focusable   bool                     `json:"focusable,omitempty"`
	Focused     bool                     `json:"focused,omitempty"`

	// ARIA-like live region, landmark, and relationship fields
	Live        string `json:"live,omitempty"`
	Relevant    string `json:"relevant,omitempty"`
	Atomic      bool   `json:"atomic,omitempty"`
	Landmark    string `json:"landmark,omitempty"`
	LabelledBy  string `json:"labelled_by,omitempty"`
	DescribedBy string `json:"described_by,omitempty"`
	Controls    string `json:"controls,omitempty"`
	Owns        string `json:"owns,omitempty"`
	FlowTo      string `json:"flow_to,omitempty"`

	// WAI-ARIA 1.2/1.3 properties
	Level            int    `json:"level,omitempty"`
	Orientation      string `json:"orientation,omitempty"`
	ActiveDescendant string `json:"active_descendant,omitempty"`
	PosInSet         int    `json:"pos_in_set,omitempty"`
	SetSize          int    `json:"set_size,omitempty"`
	HasPopup         string `json:"has_popup,omitempty"`
	ErrorMessage     string `json:"error_message,omitempty"`
	Current          string `json:"current,omitempty"`
	Autocomplete     string `json:"autocomplete,omitempty"`
	Placeholder      string `json:"placeholder,omitempty"`
	Sort             string `json:"sort,omitempty"`
	KeyShortcuts     string `json:"key_shortcuts,omitempty"`
	Details          string `json:"details,omitempty"`
	RoleDescription  string `json:"role_description,omitempty"`
}
