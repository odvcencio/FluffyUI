package widgets

import (
	"fmt"
	"strings"

	"m31labs.dev/fluffyui/accessibility"
	"m31labs.dev/fluffyui/backend"
	"m31labs.dev/fluffyui/runtime"
	"m31labs.dev/fluffyui/terminal"
)

// TagInput is a focusable widget for entering and managing tags.
// Tags are displayed as [tag1] [tag2] followed by an input area.
type TagInput struct {
	FocusableBase

	tags     []string
	inputBuf string
	maxTags  int
	onChange func(tags []string)
	services runtime.Services

	style    backend.Style
	styleSet bool
}

// NewTagInput creates a tag input widget.
func NewTagInput() *TagInput {
	t := &TagInput{
		style: backend.DefaultStyle(),
	}
	t.Base.Role = accessibility.RoleGroup
	t.Base.Label = "Tag input"
	return t
}

// SetMaxTags sets the maximum number of tags allowed (0 = unlimited).
func (t *TagInput) SetMaxTags(max int) {
	if t == nil {
		return
	}
	t.maxTags = max
}

// SetOnChange sets the change handler.
func (t *TagInput) SetOnChange(fn func(tags []string)) {
	if t == nil {
		return
	}
	t.onChange = fn
}

// Tags returns a copy of the current tags.
func (t *TagInput) Tags() []string {
	if t == nil || len(t.tags) == 0 {
		return nil
	}
	out := make([]string, len(t.tags))
	copy(out, t.tags)
	return out
}

// SetTags replaces the current tags.
func (t *TagInput) SetTags(tags []string) {
	if t == nil {
		return
	}
	t.tags = make([]string, len(tags))
	copy(t.tags, tags)
	t.syncA11y()
}

// AddTag adds a tag if not duplicate and under max limit.
func (t *TagInput) AddTag(tag string) bool {
	if t == nil {
		return false
	}
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return false
	}
	// Check for duplicates.
	for _, existing := range t.tags {
		if existing == tag {
			return false
		}
	}
	// Check max limit.
	if t.maxTags > 0 && len(t.tags) >= t.maxTags {
		return false
	}
	t.tags = append(t.tags, tag)
	t.syncA11y()
	if t.onChange != nil {
		t.onChange(t.Tags())
	}
	if announcer := t.services.Announcer(); announcer != nil {
		announcer.Announce(fmt.Sprintf("Added tag %s", tag), accessibility.PriorityPolite)
	}
	return true
}

// RemoveTag removes a tag by value.
func (t *TagInput) RemoveTag(tag string) bool {
	if t == nil {
		return false
	}
	for i, existing := range t.tags {
		if existing == tag {
			t.tags = append(t.tags[:i], t.tags[i+1:]...)
			t.syncA11y()
			if t.onChange != nil {
				t.onChange(t.Tags())
			}
			if announcer := t.services.Announcer(); announcer != nil {
				announcer.Announce(fmt.Sprintf("Removed tag %s", tag), accessibility.PriorityPolite)
			}
			return true
		}
	}
	return false
}

// InputBuffer returns the current input text.
func (t *TagInput) InputBuffer() string {
	if t == nil {
		return ""
	}
	return t.inputBuf
}

// SetStyle sets the widget style.
func (t *TagInput) SetStyle(style backend.Style) {
	if t == nil {
		return
	}
	t.style = style
	t.styleSet = true
}

// StyleType returns the selector type name.
func (t *TagInput) StyleType() string {
	return "TagInput"
}

// Bind attaches app services.
func (t *TagInput) Bind(services runtime.Services) {
	if t == nil {
		return
	}
	t.services = services
}

// Unbind releases app services.
func (t *TagInput) Unbind() {
	if t == nil {
		return
	}
	t.services = runtime.Services{}
}

// Measure returns the desired size.
func (t *TagInput) Measure(constraints runtime.Constraints) runtime.Size {
	return t.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		text := t.renderText()
		width := textWidth(text)
		if width < 3 {
			width = 3
		}
		return contentConstraints.Constrain(runtime.Size{Width: width, Height: 1})
	})
}

// Render draws the tag input widget.
func (t *TagInput) Render(ctx runtime.RenderContext) {
	if t == nil {
		return
	}
	t.syncA11y()
	bounds := t.ContentBounds()
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return
	}
	style := resolveBaseStyle(ctx, t, t.style, t.styleSet)
	text := t.renderText()
	writePadded(ctx.Buffer, bounds.X, bounds.Y, bounds.Width, truncateString(text, bounds.Width), style)
}

func (t *TagInput) renderText() string {
	var parts []string
	for _, tag := range t.tags {
		parts = append(parts, "["+tag+"]")
	}
	input := "|" + t.inputBuf + "_"
	parts = append(parts, input)
	return strings.Join(parts, " ")
}

// HandleMessage processes keyboard input.
func (t *TagInput) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if t == nil || !t.focused {
		return runtime.Unhandled()
	}
	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}
	switch key.Key {
	case terminal.KeyEnter, terminal.KeyTab:
		tag := strings.TrimSpace(t.inputBuf)
		if tag != "" {
			t.AddTag(tag)
			t.inputBuf = ""
		}
		return runtime.Handled()
	case terminal.KeyBackspace:
		if len(t.inputBuf) > 0 {
			// Remove last character from input buffer.
			runes := []rune(t.inputBuf)
			t.inputBuf = string(runes[:len(runes)-1])
		} else if len(t.tags) > 0 {
			// Remove last tag.
			removed := t.tags[len(t.tags)-1]
			t.RemoveTag(removed)
		}
		return runtime.Handled()
	case terminal.KeyRune:
		t.inputBuf += string(key.Rune)
		return runtime.Handled()
	}
	return runtime.Unhandled()
}

func (t *TagInput) syncA11y() {
	if t == nil {
		return
	}
	if t.Base.Role == "" {
		t.Base.Role = accessibility.RoleGroup
	}
	if len(t.tags) > 0 {
		t.Base.Label = fmt.Sprintf("Tag input, %d tags", len(t.tags))
	} else {
		t.Base.Label = "Tag input, empty"
	}
}

var _ runtime.Widget = (*TagInput)(nil)
var _ runtime.Focusable = (*TagInput)(nil)
var _ runtime.Bindable = (*TagInput)(nil)
var _ runtime.Unbindable = (*TagInput)(nil)
