package widgets

import (
	"strings"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/runtime"
	uistyle "github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/terminal"
)

// Combobox is a dropdown select that also accepts free-text input.
// It combines the behavior of an Input and a Select.
type Combobox struct {
	FocusableBase

	options  []string
	filtered []string
	text     string
	cursor   int // index in filtered list, -1 = none
	open     bool
	onChange func(string)

	style    backend.Style
	focusStyle backend.Style
	styleSet bool
	focusSet bool
	services runtime.Services
}

// NewCombobox creates a combobox with the given options.
func NewCombobox(options ...string) *Combobox {
	c := &Combobox{
		options:    options,
		filtered:   options,
		cursor:     -1,
		style:      backend.DefaultStyle(),
		focusStyle: backend.DefaultStyle().Reverse(true),
	}
	c.Base.Role = accessibility.RoleCombobox
	c.Base.HasPopup = "listbox"
	c.syncA11y()
	return c
}

// Text returns the current text.
func (c *Combobox) Text() string {
	if c == nil {
		return ""
	}
	return c.text
}

// SetText updates the text and filters options.
func (c *Combobox) SetText(s string) {
	if c == nil {
		return
	}
	c.text = s
	c.filter()
	c.syncA11y()
	c.relayout()
}

// Options returns the full option list.
func (c *Combobox) Options() []string {
	if c == nil {
		return nil
	}
	return c.options
}

// SetOptions replaces the option list and re-filters.
func (c *Combobox) SetOptions(opts []string) {
	if c == nil {
		return
	}
	c.options = opts
	c.filter()
	c.syncA11y()
	c.relayout()
}

// SetOnChange sets the change handler, called when text changes or an option is selected.
func (c *Combobox) SetOnChange(fn func(string)) {
	if c == nil {
		return
	}
	c.onChange = fn
}

// SetStyle sets the normal style.
func (c *Combobox) SetStyle(style backend.Style) {
	if c == nil {
		return
	}
	c.style = style
	c.styleSet = true
}

// SetFocusStyle sets the focused style.
func (c *Combobox) SetFocusStyle(style backend.Style) {
	if c == nil {
		return
	}
	c.focusStyle = style
	c.focusSet = true
}

// StyleType returns the selector type name.
func (c *Combobox) StyleType() string {
	return "Combobox"
}

// Bind attaches app services.
func (c *Combobox) Bind(services runtime.Services) {
	if c == nil {
		return
	}
	c.services = services
}

// Unbind releases app services.
func (c *Combobox) Unbind() {
	if c == nil {
		return
	}
	c.services = runtime.Services{}
}

// Measure returns the desired size.
func (c *Combobox) Measure(constraints runtime.Constraints) runtime.Size {
	return c.measureWithStyle(constraints, func(cc runtime.Constraints) runtime.Size {
		width := textWidth(c.text) + 4
		if width < 10 {
			width = 10
		}
		height := 1
		if c.open && len(c.filtered) > 0 {
			height += len(c.filtered)
		}
		return cc.Constrain(runtime.Size{Width: width, Height: height})
	})
}

// Render draws the combobox.
func (c *Combobox) Render(ctx runtime.RenderContext) {
	if c == nil {
		return
	}
	outer := c.bounds
	content := c.ContentBounds()
	if outer.Width <= 0 || outer.Height <= 0 {
		return
	}
	style := c.style
	resolved := ctx.ResolveStyle(c)
	if !resolved.IsZero() {
		final := resolved
		if c.styleSet {
			final = final.Merge(uistyle.FromBackend(c.style))
		}
		if c.focused && c.focusSet {
			final = final.Merge(uistyle.FromBackend(c.focusStyle))
		}
		style = final.ToBackend()
	} else if c.focused {
		style = c.focusStyle
	}
	ctx.Buffer.Fill(outer, ' ', style)
	if content.Width <= 0 || content.Height <= 0 {
		return
	}
	// Draw input line
	available := max(0, content.Width-4)
	text := "[" + truncateString(c.text, available) + " v]"
	writePadded(ctx.Buffer, content.X, content.Y, content.Width, text, style)

	// Draw filtered options below
	if c.open && len(c.filtered) > 0 {
		for i, opt := range c.filtered {
			y := content.Y + 1 + i
			if y >= content.Y+content.Height {
				break
			}
			optStyle := style
			if i == c.cursor {
				optStyle = c.focusStyle
			}
			label := " " + truncateString(opt, max(0, content.Width-1))
			writePadded(ctx.Buffer, content.X, y, content.Width, label, optStyle)
		}
	}
}

// HandleMessage handles keyboard input.
func (c *Combobox) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if c == nil || !c.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyUp:
		if c.open && c.cursor > 0 {
			c.cursor--
			c.Invalidate()
			return runtime.Handled()
		}
	case terminal.KeyDown:
		if !c.open {
			c.open = true
			if len(c.filtered) > 0 {
				c.cursor = 0
			}
			c.syncA11y()
			c.Invalidate()
			return runtime.Handled()
		}
		if c.cursor < len(c.filtered)-1 {
			c.cursor++
			c.Invalidate()
			return runtime.Handled()
		}
	case terminal.KeyEnter:
		if c.open && c.cursor >= 0 && c.cursor < len(c.filtered) {
			c.text = c.filtered[c.cursor]
			c.open = false
			c.cursor = -1
			c.filter()
			c.syncA11y()
			c.relayout()
			if announcer := c.services.Announcer(); announcer != nil {
				announcer.AnnounceChange(c)
			}
			if c.onChange != nil {
				c.onChange(c.text)
			}
			return runtime.Handled()
		}
	case terminal.KeyEscape:
		if c.open {
			c.open = false
			c.cursor = -1
			c.syncA11y()
			c.Invalidate()
			return runtime.Handled()
		}
	case terminal.KeyBackspace:
		if len(c.text) > 0 {
			runes := []rune(c.text)
			c.text = string(runes[:len(runes)-1])
			c.open = true
			c.filter()
			c.syncA11y()
			c.relayout()
			if c.onChange != nil {
				c.onChange(c.text)
			}
			return runtime.Handled()
		}
	case terminal.KeyRune:
		c.text += string(key.Rune)
		c.open = true
		c.filter()
		c.syncA11y()
		c.relayout()
		if c.onChange != nil {
			c.onChange(c.text)
		}
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (c *Combobox) filter() {
	if c.text == "" {
		c.filtered = c.options
	} else {
		lower := strings.ToLower(c.text)
		c.filtered = nil
		for _, opt := range c.options {
			if strings.Contains(strings.ToLower(opt), lower) {
				c.filtered = append(c.filtered, opt)
			}
		}
	}
	if c.cursor >= len(c.filtered) {
		c.cursor = len(c.filtered) - 1
	}
}

func (c *Combobox) syncA11y() {
	if c == nil {
		return
	}
	if c.Base.Role == "" {
		c.Base.Role = accessibility.RoleCombobox
	}
	c.Base.HasPopup = "listbox"
	c.Base.State.Expanded = accessibility.BoolPtr(c.open)
	c.Base.Autocomplete = "list"
	if c.text != "" {
		c.Base.Value = &accessibility.ValueInfo{Text: c.text}
	} else {
		c.Base.Value = nil
	}
}

func (c *Combobox) relayout() {
	if c == nil {
		return
	}
	c.Invalidate()
	c.services.Relayout()
}

var _ runtime.Widget = (*Combobox)(nil)
var _ runtime.Focusable = (*Combobox)(nil)
var _ runtime.Bindable = (*Combobox)(nil)
var _ runtime.Unbindable = (*Combobox)(nil)
