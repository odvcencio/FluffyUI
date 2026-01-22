package widgets

import (
	"strings"

	"github.com/odvcencio/fluffy-ui/accessibility"
	"github.com/odvcencio/fluffy-ui/backend"
	"github.com/odvcencio/fluffy-ui/runtime"
	"github.com/odvcencio/fluffy-ui/terminal"
)

// Tab represents a single tab.
type Tab struct {
	Title   string
	Content runtime.Widget
}

// Tabs is a tabbed container widget.
type Tabs struct {
	FocusableBase
	Tabs          []Tab
	selected      int
	label         string
	style         backend.Style
	selectedStyle backend.Style
}

// NewTabs creates a tab container.
func NewTabs(tabs ...Tab) *Tabs {
	w := &Tabs{
		Tabs:          tabs,
		selected:      0,
		label:         "Tabs",
		style:         backend.DefaultStyle(),
		selectedStyle: backend.DefaultStyle().Reverse(true),
	}
	w.Base.Role = accessibility.RoleTabList
	w.syncA11y()
	return w
}

// Measure returns the size of the selected tab.
func (t *Tabs) Measure(constraints runtime.Constraints) runtime.Size {
	if t == nil || len(t.Tabs) == 0 {
		return constraints.MinSize()
	}
	selected := t.selectedTab()
	if selected == nil || selected.Content == nil {
		return constraints.MinSize()
	}
	size := selected.Content.Measure(constraints)
	if size.Height < 1 {
		size.Height = 1
	}
	size.Height += 1
	return constraints.Constrain(size)
}

// Layout positions the selected tab content.
func (t *Tabs) Layout(bounds runtime.Rect) {
	t.Base.Layout(bounds)
	t.layoutSelected()
}

// Render draws tab titles and content.
func (t *Tabs) Render(ctx runtime.RenderContext) {
	if t == nil {
		return
	}
	t.syncA11y()
	bounds := t.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	x := bounds.X
	for i, tab := range t.Tabs {
		label := " " + tab.Title + " "
		style := t.style
		if i == t.selected {
			style = t.selectedStyle
		}
		if x < bounds.X+bounds.Width {
			available := bounds.Width - (x - bounds.X)
			label = truncateString(label, available)
			ctx.Buffer.SetString(x, bounds.Y, label, style)
			x += len(label)
		}
	}
	selected := t.selectedTab()
	if selected != nil && selected.Content != nil {
		selected.Content.Render(ctx)
	}
}

// HandleMessage switches tabs.
func (t *Tabs) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t == nil || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyLeft:
		t.setSelected(t.selected - 1)
		return runtime.Handled()
	case terminal.KeyRight:
		t.setSelected(t.selected + 1)
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

// SelectedIndex returns the current tab index.
func (t *Tabs) SelectedIndex() int {
	if t == nil || len(t.Tabs) == 0 {
		return 0
	}
	if t.selected < 0 {
		t.selected = 0
	}
	if t.selected >= len(t.Tabs) {
		t.selected = len(t.Tabs) - 1
	}
	return t.selected
}

// SetSelected updates the active tab index.
func (t *Tabs) SetSelected(index int) {
	t.setSelected(index)
}

// SetLabel updates the accessibility label.
func (t *Tabs) SetLabel(label string) {
	if t == nil {
		return
	}
	t.label = label
	t.syncA11y()
}

// ChildWidgets returns the selected tab content.
func (t *Tabs) ChildWidgets() []runtime.Widget {
	selected := t.selectedTab()
	if selected == nil || selected.Content == nil {
		return nil
	}
	return []runtime.Widget{selected.Content}
}

func (t *Tabs) selectedTab() *Tab {
	if t == nil || len(t.Tabs) == 0 {
		return nil
	}
	if t.selected < 0 {
		t.selected = 0
	}
	if t.selected >= len(t.Tabs) {
		t.selected = len(t.Tabs) - 1
	}
	return &t.Tabs[t.selected]
}

func (t *Tabs) setSelected(index int) {
	if len(t.Tabs) == 0 {
		t.selected = 0
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(t.Tabs) {
		index = len(t.Tabs) - 1
	}
	t.selected = index
	t.layoutSelected()
	t.syncA11y()
}

func (t *Tabs) layoutSelected() {
	selected := t.selectedTab()
	if selected == nil || selected.Content == nil {
		return
	}
	bounds := t.bounds
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	contentBounds := runtime.Rect{
		X:      bounds.X,
		Y:      bounds.Y + 1,
		Width:  bounds.Width,
		Height: max(0, bounds.Height-1),
	}
	selected.Content.Layout(contentBounds)
}

func (t *Tabs) syncA11y() {
	if t == nil {
		return
	}
	if t.Base.Role == "" {
		t.Base.Role = accessibility.RoleTabList
	}
	label := strings.TrimSpace(t.label)
	if label == "" {
		label = "Tabs"
	}
	t.Base.Label = label
	if tab := t.selectedTab(); tab != nil {
		t.Base.Value = &accessibility.ValueInfo{Text: tab.Title}
	} else {
		t.Base.Value = nil
	}
	if len(t.Tabs) > 0 {
		titles := make([]string, 0, len(t.Tabs))
		for _, tab := range t.Tabs {
			if strings.TrimSpace(tab.Title) != "" {
				titles = append(titles, tab.Title)
			}
		}
		t.Base.Description = strings.Join(titles, ", ")
	} else {
		t.Base.Description = ""
	}
}
