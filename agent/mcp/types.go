package mcp

import "time"

const SchemaVersion = "fluffy-mcp/v1"

type Capabilities struct {
	AllowText      bool   `json:"allow_text"`
	AllowClipboard bool   `json:"allow_clipboard"`
	Transport      string `json:"transport"`
	Subscriptions  bool   `json:"subscriptions"`
}

type AppInfo struct {
	Width      int   `json:"width"`
	Height     int   `json:"height"`
	LayerCount int   `json:"layer_count"`
	UptimeMs   int64 `json:"uptime_ms,omitempty"`
}

type Dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Rect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type StateSet struct {
	Focused  bool  `json:"focused,omitempty"`
	Disabled bool  `json:"disabled,omitempty"`
	Hidden   bool  `json:"hidden,omitempty"`
	Checked  *bool `json:"checked,omitempty"`
	Selected bool  `json:"selected,omitempty"`
	Expanded *bool `json:"expanded,omitempty"`
	Pressed  bool  `json:"pressed,omitempty"`
	ReadOnly bool  `json:"readonly,omitempty"`
	Required bool  `json:"required,omitempty"`
	Invalid  bool  `json:"invalid,omitempty"`
	Busy     bool  `json:"busy,omitempty"`
}

type WidgetInfo struct {
	ID          string   `json:"id"`
	Role        string   `json:"role"`
	Label       string   `json:"label,omitempty"`
	Value       string   `json:"value,omitempty"`
	Description string   `json:"description,omitempty"`
	Bounds      Rect     `json:"bounds"`
	State       StateSet `json:"state,omitempty"`
	Actions     []string `json:"actions,omitempty"`
	ChildrenIDs []string `json:"children_ids,omitempty"`
	ParentID    string   `json:"parent_id,omitempty"`
}

type WidgetNode struct {
	ID          string       `json:"id"`
	Role        string       `json:"role"`
	Label       string       `json:"label,omitempty"`
	Value       string       `json:"value,omitempty"`
	Description string       `json:"description,omitempty"`
	Bounds      Rect         `json:"bounds"`
	State       StateSet     `json:"state,omitempty"`
	Actions     []string     `json:"actions,omitempty"`
	Children    []WidgetNode `json:"children,omitempty"`
}

type Snapshot struct {
	Timestamp  time.Time    `json:"timestamp"`
	Dimensions Dimensions   `json:"dimensions"`
	LayerCount int          `json:"layer_count"`
	FocusedID  string       `json:"focused_id,omitempty"`
	Widgets    []WidgetInfo `json:"widgets"`
	Text       string       `json:"text,omitempty"`
}

type TreeSnapshot struct {
	Timestamp  time.Time    `json:"timestamp"`
	Dimensions Dimensions   `json:"dimensions"`
	LayerCount int          `json:"layer_count"`
	FocusedID  string       `json:"focused_id,omitempty"`
	Widgets    []WidgetNode `json:"widgets"`
	Text       string       `json:"text,omitempty"`
}

type CursorPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type SelectionBounds struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type CellInfo struct {
	Char  string    `json:"char"`
	Style StyleInfo `json:"style"`
}

type StyleInfo struct {
	FG            any  `json:"fg,omitempty"`
	BG            any  `json:"bg,omitempty"`
	Bold          bool `json:"bold,omitempty"`
	Italic        bool `json:"italic,omitempty"`
	Underline     bool `json:"underline,omitempty"`
	Dim           bool `json:"dim,omitempty"`
	Blink         bool `json:"blink,omitempty"`
	Reverse       bool `json:"reverse,omitempty"`
	Strikethrough bool `json:"strikethrough,omitempty"`
}

type WidgetMatch struct {
	ID      string `json:"id"`
	Label   string `json:"label,omitempty"`
	Context string `json:"context,omitempty"`
	Layer   int    `json:"layer,omitempty"`
}

type ActionResult struct {
	Status           string        `json:"status"`
	WidgetID         string        `json:"widget_id,omitempty"`
	ResolvedTo       string        `json:"resolved_to,omitempty"`
	Message          string        `json:"message,omitempty"`
	Matches          []WidgetMatch `json:"matches,omitempty"`
	ResolutionReason string        `json:"resolution_reason,omitempty"`
}

type ValueChange struct {
	Old any `json:"old"`
	New any `json:"new"`
}

type WidgetChange struct {
	ID      string                 `json:"id"`
	Changes map[string]ValueChange `json:"changes"`
}

type Diff struct {
	WidgetsAdded      []string       `json:"widgets_added,omitempty"`
	WidgetsRemoved    []string       `json:"widgets_removed,omitempty"`
	WidgetsModified   []WidgetChange `json:"widgets_modified,omitempty"`
	TextChanged       bool           `json:"text_changed"`
	DimensionsChanged bool           `json:"dimensions_changed"`
	LayerCountChanged bool           `json:"layer_count_changed"`
	FocusChanged      bool           `json:"focus_changed"`
}
