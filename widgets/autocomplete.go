package widgets

import (
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	"github.com/odvcencio/fluffyui/terminal"
)

// AutoComplete provides an input with suggestion list.
type AutoComplete struct {
	FocusableBase

	input          *Input
	options        []string
	suggestions    []string
	selected       int
	maxSuggestions int
	label          string
	style          backend.Style
	suggestionSty  backend.Style
	selectedSty    backend.Style

	provider func(query string) []string
	onSelect func(value string)
}

// NewAutoComplete creates a new AutoComplete widget.
func NewAutoComplete() *AutoComplete {
	ac := &AutoComplete{
		input:          NewInput(),
		maxSuggestions: 6,
		label:          "Auto Complete",
		style:          backend.DefaultStyle(),
		suggestionSty:  backend.DefaultStyle(),
		selectedSty:    backend.DefaultStyle().Reverse(true),
	}
	ac.input.SetPlaceholder("Type to search")
	ac.input.SetOnChange(func(text string) {
		ac.updateSuggestions(text)
	})
	ac.Base.Role = accessibility.RoleTextbox
	ac.syncA11y()
	return ac
}

// Input returns the underlying input widget.
func (a *AutoComplete) Input() *Input {
	if a == nil {
		return nil
	}
	return a.input
}

// SetOptions sets the candidate options for filtering.
func (a *AutoComplete) SetOptions(options []string) {
	if a == nil {
		return
	}
	a.options = options
	a.updateSuggestions(a.query())
}

// SetProvider sets a custom suggestion provider.
func (a *AutoComplete) SetProvider(fn func(query string) []string) {
	if a == nil {
		return
	}
	a.provider = fn
	a.updateSuggestions(a.query())
}

// SetMaxSuggestions limits the number of suggestions shown.
func (a *AutoComplete) SetMaxSuggestions(limit int) {
	if a == nil {
		return
	}
	if limit <= 0 {
		limit = 1
	}
	a.maxSuggestions = limit
}

// SetOnSelect registers a selection callback.
func (a *AutoComplete) SetOnSelect(fn func(value string)) {
	if a == nil {
		return
	}
	a.onSelect = fn
}

// SetLabel updates the accessibility label.
func (a *AutoComplete) SetLabel(label string) {
	if a == nil {
		return
	}
	a.label = label
	a.syncA11y()
	if a.input != nil {
		a.input.SetLabel(label)
	}
}

// Query returns the current query text.
func (a *AutoComplete) Query() string {
	if a == nil {
		return ""
	}
	return a.query()
}

// SetQuery updates the query text and refreshes suggestions.
func (a *AutoComplete) SetQuery(query string) {
	if a == nil {
		return
	}
	if a.input != nil {
		a.input.SetText(query)
	}
	a.updateSuggestions(query)
	a.Invalidate()
}

// StyleType returns the selector type name.
func (a *AutoComplete) StyleType() string {
	return "AutoComplete"
}

// Measure returns desired size.
func (a *AutoComplete) Measure(constraints runtime.Constraints) runtime.Size {
	inputSize := runtime.Size{}
	if a.input != nil {
		inputSize = a.input.Measure(constraints)
	}
	height := inputSize.Height
	if height <= 0 {
		height = 1
	}
	visible := min(len(a.suggestions), a.maxSuggestions)
	if visible > 0 {
		height += visible
	}
	width := inputSize.Width
	for _, s := range a.suggestions {
		w := textWidth(s)
		if w > width {
			width = w
		}
	}
	if width <= 0 {
		width = constraints.MinWidth
	}
	return constraints.Constrain(runtime.Size{Width: width, Height: height})
}

// Layout positions input and suggestion list.
func (a *AutoComplete) Layout(bounds runtime.Rect) {
	a.FocusableBase.Layout(bounds)
	content := a.ContentBounds()
	if a.input != nil {
		a.input.Layout(runtime.Rect{X: content.X, Y: content.Y, Width: content.Width, Height: 1})
	}
}

// Render draws the input and suggestions.
func (a *AutoComplete) Render(ctx runtime.RenderContext) {
	if a == nil {
		return
	}
	a.syncA11y()
	outer := a.bounds
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	baseStyle := resolveBaseStyle(ctx, a, a.style, true)
	ctx.Buffer.Fill(outer, ' ', baseStyle)
	if a.input != nil {
		a.input.Render(ctx)
	}
	content := a.ContentBounds()
	startY := content.Y + 1
	maxRows := min(len(a.suggestions), a.maxSuggestions)
	for i := 0; i < maxRows; i++ {
		line := a.suggestions[i]
		style := baseStyle
		if i == a.selected {
			style = a.selectedSty
		} else if a.suggestionSty != (backend.Style{}) {
			style = mergeBackendStyles(baseStyle, a.suggestionSty)
		}
		writePadded(ctx.Buffer, content.X, startY+i, content.Width, truncateString(line, content.Width), style)
	}
}

// HandleMessage processes keyboard input.
func (a *AutoComplete) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if a == nil || !a.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if ok {
		switch key.Key {
		case terminal.KeyUp:
			if a.selected > 0 {
				a.selected--
				return runtime.Handled()
			}
		case terminal.KeyDown:
			if a.selected < len(a.suggestions)-1 && a.selected < a.maxSuggestions-1 {
				a.selected++
				return runtime.Handled()
			}
		case terminal.KeyEnter:
			if a.selected >= 0 && a.selected < len(a.suggestions) {
				value := a.suggestions[a.selected]
				if a.input != nil {
					a.input.SetText(value)
				}
				if a.onSelect != nil {
					a.onSelect(value)
				}
				a.suggestions = nil
				a.selected = 0
				return runtime.Handled()
			}
		case terminal.KeyEscape:
			a.suggestions = nil
			a.selected = 0
			return runtime.Handled()
		}
	}
	if a.input != nil {
		return a.input.HandleMessage(msg)
	}
	return runtime.Unhandled()
}

// Focus forwards focus to the input.
func (a *AutoComplete) Focus() {
	a.FocusableBase.Focus()
	if a.input != nil {
		a.input.Focus()
	}
}

// Blur clears focus.
func (a *AutoComplete) Blur() {
	a.FocusableBase.Blur()
	if a.input != nil {
		a.input.Blur()
	}
}

// ChildWidgets returns child widgets.
func (a *AutoComplete) ChildWidgets() []runtime.Widget {
	if a == nil || a.input == nil {
		return nil
	}
	return []runtime.Widget{a.input}
}

func (a *AutoComplete) updateSuggestions(query string) {
	if a == nil {
		return
	}
	query = strings.TrimSpace(query)
	if query == "" {
		a.suggestions = nil
		a.selected = 0
		return
	}
	if a.provider != nil {
		a.suggestions = a.provider(query)
	} else {
		needle := strings.ToLower(query)
		var matches []string
		for _, opt := range a.options {
			if strings.Contains(strings.ToLower(opt), needle) {
				matches = append(matches, opt)
			}
		}
		a.suggestions = matches
	}
	if a.maxSuggestions > 0 && len(a.suggestions) > a.maxSuggestions {
		a.suggestions = a.suggestions[:a.maxSuggestions]
	}
	if a.selected >= len(a.suggestions) {
		a.selected = 0
	}
}

func (a *AutoComplete) query() string {
	if a == nil || a.input == nil {
		return ""
	}
	return a.input.Text()
}

func (a *AutoComplete) syncA11y() {
	if a == nil {
		return
	}
	if a.Base.Role == "" {
		a.Base.Role = accessibility.RoleTextbox
	}
	label := strings.TrimSpace(a.label)
	if label == "" {
		label = "Auto Complete"
	}
	a.Base.Label = label
	if a.input != nil {
		a.Base.Value = &accessibility.ValueInfo{Text: a.input.Text()}
	}
}

var _ runtime.Widget = (*AutoComplete)(nil)
var _ runtime.Focusable = (*AutoComplete)(nil)
var _ runtime.ChildProvider = (*AutoComplete)(nil)
var _ Searchable = (*AutoComplete)(nil)
