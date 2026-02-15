package widgets

import (
	"strings"
	"unicode"

	"github.com/odvcencio/fluffyui/accessibility"
	"github.com/odvcencio/fluffyui/backend"
	"github.com/odvcencio/fluffyui/clipboard"
	"github.com/odvcencio/fluffyui/forms"
	"github.com/odvcencio/fluffyui/runtime"
	uistyle "github.com/odvcencio/fluffyui/style"
	"github.com/odvcencio/fluffyui/terminal"
)

// MaskedInput is a text input widget with a format mask.
// The mask defines the expected format for input using special characters:
//   - # = digit (0-9)
//   - A = letter (a-zA-Z)
//   - * = alphanumeric (letter or digit)
//   - ? = any character
//   - \ = escape next character as literal
//
// Any other character in the mask is treated as a literal that is
// automatically inserted and cannot be edited.
type MaskedInput struct {
	FocusableBase

	mask        string // format mask, e.g., "(###) ###-####"
	placeholder rune   // character shown for unfilled positions
	value       string // raw value without mask characters

	label       string
	style       backend.Style
	focusStyle  backend.Style
	services    runtime.Services
	styleSet    bool
	focusSet    bool
	validators  []forms.Validator
	valErrors   []forms.ValidationError
	valMessages []string

	cursorPos int // cursor position in the masked display

	onSubmit func(value string)
	onChange func(value string)
}

// MaskedInputOption is a functional option for MaskedInput.
type MaskedInputOption func(*MaskedInput)

// WithPlaceholder sets the placeholder character for unfilled positions.
func WithPlaceholder(r rune) MaskedInputOption {
	return func(m *MaskedInput) {
		m.placeholder = r
	}
}

// WithMaskedLabel sets the accessibility label.
func WithMaskedLabel(label string) MaskedInputOption {
	return func(m *MaskedInput) {
		m.label = label
	}
}

// WithMaskedStyle sets the normal style.
func WithMaskedStyle(s backend.Style) MaskedInputOption {
	return func(m *MaskedInput) {
		m.style = s
		m.styleSet = true
	}
}

// WithMaskedFocusStyle sets the focused style.
func WithMaskedFocusStyle(s backend.Style) MaskedInputOption {
	return func(m *MaskedInput) {
		m.focusStyle = s
		m.focusSet = true
	}
}

// NewMaskedInput creates a new masked input widget.
func NewMaskedInput(mask string, opts ...MaskedInputOption) *MaskedInput {
	m := &MaskedInput{
		mask:        mask,
		placeholder: '_',
		style:       backend.DefaultStyle(),
		focusStyle:  backend.DefaultStyle().Bold(true),
	}
	for _, opt := range opts {
		opt(m)
	}
	m.Base.Role = accessibility.RoleTextbox
	m.cursorPos = m.firstEditablePosition()
	m.syncA11y()
	return m
}

// Bind attaches app services.
func (m *MaskedInput) Bind(services runtime.Services) {
	if m == nil {
		return
	}
	m.services = services
}

// Unbind releases app services.
func (m *MaskedInput) Unbind() {
	if m == nil {
		return
	}
	m.services = runtime.Services{}
}

// Blur removes focus and announces validation errors if any.
func (m *MaskedInput) Blur() {
	if m == nil {
		return
	}
	m.FocusableBase.Blur()
	if len(m.validators) > 0 {
		errs := m.Validate()
		if len(errs) > 0 && len(m.valMessages) > 0 {
			if announcer := m.services.Announcer(); announcer != nil {
				announcer.Announce(strings.Join(m.valMessages, ", "), accessibility.PriorityAssertive)
			}
		}
	}
}

// isMaskComplete reports whether all editable slots are filled.
func (m *MaskedInput) isMaskComplete() bool {
	slots := m.parseMask()
	editableCount := 0
	for _, slot := range slots {
		if !slot.literal {
			editableCount++
		}
	}
	return len([]rune(m.value)) >= editableCount
}

// SetMask updates the mask pattern.
func (m *MaskedInput) SetMask(mask string) {
	if m == nil {
		return
	}
	m.mask = mask
	m.value = ""
	m.cursorPos = m.firstEditablePosition()
	m.syncA11y()
}

// Mask returns the current mask pattern.
func (m *MaskedInput) Mask() string {
	if m == nil {
		return ""
	}
	return m.mask
}

// SetPlaceholder sets the placeholder character.
func (m *MaskedInput) SetPlaceholder(r rune) {
	if m == nil {
		return
	}
	m.placeholder = r
}

// Placeholder returns the placeholder character.
func (m *MaskedInput) Placeholder() rune {
	if m == nil {
		return '_'
	}
	return m.placeholder
}

// Value returns the raw input value without mask characters.
func (m *MaskedInput) Value() string {
	if m == nil {
		return ""
	}
	return m.value
}

// SetValue sets the raw value (without mask characters).
func (m *MaskedInput) SetValue(value string) {
	if m == nil {
		return
	}
	// Filter value to only include valid characters for editable slots
	m.value = m.filterValue(value)
	m.cursorPos = m.lastFilledPosition() + 1
	if m.cursorPos > m.maskDisplayLength() {
		m.cursorPos = m.maskDisplayLength()
	}
	m.syncA11y()
}

// MaskedValue returns the formatted value with mask characters included.
func (m *MaskedInput) MaskedValue() string {
	if m == nil {
		return ""
	}
	return m.formatWithMask()
}

// SetLabel sets the accessibility label.
func (m *MaskedInput) SetLabel(label string) {
	if m == nil {
		return
	}
	m.label = label
	m.syncA11y()
}

// SetStyle sets the normal style.
func (m *MaskedInput) SetStyle(style backend.Style) {
	if m == nil {
		return
	}
	m.style = style
	m.styleSet = true
}

// SetFocusStyle sets the focused style.
func (m *MaskedInput) SetFocusStyle(style backend.Style) {
	if m == nil {
		return
	}
	m.focusStyle = style
	m.focusSet = true
}

// StyleType returns the selector type name.
func (m *MaskedInput) StyleType() string {
	return "MaskedInput"
}

// SetOnSubmit sets the callback for when Enter is pressed.
func (m *MaskedInput) SetOnSubmit(fn func(value string)) {
	if m == nil {
		return
	}
	m.onSubmit = fn
}

// SetOnChange sets the callback for when value changes.
func (m *MaskedInput) SetOnChange(fn func(value string)) {
	if m == nil {
		return
	}
	m.onChange = fn
}

// SetValidators updates validation rules.
func (m *MaskedInput) SetValidators(validators ...forms.Validator) {
	if m == nil {
		return
	}
	m.validators = validators
}

// Validate runs validation rules and returns validation errors.
func (m *MaskedInput) Validate() []forms.ValidationError {
	if m == nil {
		return nil
	}
	errs, messages := validateValue(m.value, m.validators)
	m.valErrors = errs
	m.valMessages = messages
	return errs
}

// Errors returns the latest validation error messages.
func (m *MaskedInput) Errors() []string {
	if m == nil {
		return nil
	}
	if len(m.validators) > 0 {
		m.Validate()
	}
	if len(m.valMessages) == 0 {
		return nil
	}
	out := make([]string, len(m.valMessages))
	copy(out, m.valMessages)
	return out
}

// Valid reports whether validation passes.
func (m *MaskedInput) Valid() bool {
	if m == nil {
		return true
	}
	return len(m.Validate()) == 0
}

// CursorPos returns the current cursor position in the masked display.
func (m *MaskedInput) CursorPos() int {
	if m == nil {
		return 0
	}
	return m.cursorPos
}

// Measure returns the size needed for the input.
func (m *MaskedInput) Measure(constraints runtime.Constraints) runtime.Size {
	return m.measureWithStyle(constraints, func(contentConstraints runtime.Constraints) runtime.Size {
		// Input is 1 line tall, fills available width
		return runtime.Size{
			Width:  contentConstraints.MaxWidth,
			Height: 1,
		}
	})
}

// Layout stores the assigned bounds.
func (m *MaskedInput) Layout(bounds runtime.Rect) {
	if m == nil {
		return
	}
	m.FocusableBase.Layout(bounds)
}

// Render draws the masked input field.
func (m *MaskedInput) Render(ctx runtime.RenderContext) {
	if m == nil {
		return
	}
	outer := m.bounds
	content := m.ContentBounds()
	if outer.Width == 0 || outer.Height == 0 {
		return
	}

	style := m.style
	resolved := ctx.ResolveStyle(m)
	if !resolved.IsZero() {
		final := resolved
		if m.styleSet {
			final = final.Merge(uistyle.FromBackend(m.style))
		}
		if m.focused && m.focusSet {
			final = final.Merge(uistyle.FromBackend(m.focusStyle))
		}
		style = final.ToBackend()
	} else if m.focused {
		style = m.focusStyle
	}

	// Clear the input area
	ctx.Buffer.Fill(outer, ' ', style)

	if content.Width == 0 || content.Height == 0 {
		return
	}

	// Get the formatted display string
	display := m.formatWithMask()
	displayRunes := []rune(display)

	// Calculate visible portion
	visibleStart := 0
	if m.cursorPos >= content.Width {
		visibleStart = m.cursorPos - content.Width + 1
	}

	visibleEnd := visibleStart + content.Width
	if visibleEnd > len(displayRunes) {
		visibleEnd = len(displayRunes)
	}

	var visibleRunes []rune
	if visibleStart < len(displayRunes) {
		visibleRunes = displayRunes[visibleStart:visibleEnd]
	}

	// Draw text
	for idx, ch := range visibleRunes {
		screenX := content.X + idx
		ctx.Buffer.Set(screenX, content.Y, ch, style)
	}

	// Draw cursor if focused
	if m.focused {
		cursorX := content.X + m.cursorPos - visibleStart
		if cursorX >= content.X && cursorX < content.X+content.Width {
			var cursorChar rune = ' '
			if m.cursorPos < len(displayRunes) {
				cursorChar = displayRunes[m.cursorPos]
			}
			cursorStyle := style.Reverse(true)
			ctx.Buffer.Set(cursorX, content.Y, cursorChar, cursorStyle)
		}
	}
}

// HandleMessage processes keyboard input.
func (m *MaskedInput) HandleMessage(msg runtime.Message) runtime.HandleResult {
	if m == nil || !m.focused {
		return runtime.Unhandled()
	}

	key, ok := msg.(runtime.KeyMsg)
	if !ok {
		return runtime.Unhandled()
	}

	switch key.Key {
	case terminal.KeyCtrlC:
		if m.copyToClipboard() {
			return runtime.Handled()
		}

	case terminal.KeyCtrlV:
		if m.pasteFromClipboard() {
			return runtime.Handled()
		}

	case terminal.KeyEnter:
		if m.onSubmit != nil {
			m.onSubmit(m.value)
		}
		return runtime.WithCommand(runtime.Submit{Text: m.value})

	case terminal.KeyBackspace:
		if m.deleteAtCursor(true) {
			m.notifyChange()
		}
		return runtime.Handled()

	case terminal.KeyDelete:
		if m.deleteAtCursor(false) {
			m.notifyChange()
		}
		return runtime.Handled()

	case terminal.KeyLeft:
		m.moveCursor(-1)
		return runtime.Handled()

	case terminal.KeyRight:
		m.moveCursor(1)
		return runtime.Handled()

	case terminal.KeyHome:
		m.cursorPos = m.firstEditablePosition()
		return runtime.Handled()

	case terminal.KeyEnd:
		m.cursorPos = m.lastFilledPosition() + 1
		if pos := m.lastEditablePosition(); m.cursorPos > pos+1 {
			m.cursorPos = pos + 1
		}
		return runtime.Handled()

	case terminal.KeyRune:
		if m.insertChar(key.Rune) {
			m.notifyChange()
		}
		return runtime.Handled()

	case terminal.KeyTab:
		if key.Shift {
			return runtime.WithCommand(runtime.FocusPrev{})
		}
		return runtime.WithCommand(runtime.FocusNext{})

	case terminal.KeyEscape:
		return runtime.WithCommand(runtime.Cancel{})
	}

	return runtime.Unhandled()
}

// ClipboardCopy returns the masked value for copying.
func (m *MaskedInput) ClipboardCopy() (string, bool) {
	if m == nil {
		return "", false
	}
	return m.MaskedValue(), true
}

// ClipboardCut returns the masked value and clears the input.
func (m *MaskedInput) ClipboardCut() (string, bool) {
	if m == nil {
		return "", false
	}
	text := m.MaskedValue()
	m.value = ""
	m.cursorPos = m.firstEditablePosition()
	m.notifyChange()
	return text, true
}

// ClipboardPaste inserts text, extracting valid characters.
func (m *MaskedInput) ClipboardPaste(text string) bool {
	if m == nil || text == "" {
		return false
	}
	// Extract only valid characters and insert them
	for _, ch := range text {
		m.insertChar(ch)
	}
	m.notifyChange()
	return true
}

func (m *MaskedInput) copyToClipboard() bool {
	cb := m.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, ok := m.ClipboardCopy()
	if !ok {
		return false
	}
	_ = cb.Write(text)
	return true
}

func (m *MaskedInput) pasteFromClipboard() bool {
	cb := m.services.Clipboard()
	if cb == nil || !cb.Available() {
		return false
	}
	text, err := cb.Read()
	if err != nil || text == "" {
		return false
	}
	return m.ClipboardPaste(text)
}

func (m *MaskedInput) notifyChange() {
	m.syncA11y()
	if m.onChange != nil {
		m.onChange(m.value)
	}
	if m.isMaskComplete() {
		if announcer := m.services.Announcer(); announcer != nil {
			announcer.Announce("Input complete", accessibility.PriorityPolite)
		}
	}
}

func (m *MaskedInput) syncA11y() {
	if m == nil {
		return
	}
	label := strings.TrimSpace(m.label)
	if label == "" {
		label = "Masked Input"
	}
	if m.Base.Role == "" {
		m.Base.Role = accessibility.RoleTextbox
	}
	m.Base.Label = label
	m.Base.Value = &accessibility.ValueInfo{Text: m.MaskedValue()}
}

// parseMask returns the mask slots (type and literal status).
type maskSlot struct {
	slotType rune // '#', 'A', '*', '?', or the literal character
	literal  bool // true if this is a literal character
}

func (m *MaskedInput) parseMask() []maskSlot {
	if m == nil || m.mask == "" {
		return nil
	}

	slots := make([]maskSlot, 0, len(m.mask))
	escaped := false

	for _, ch := range m.mask {
		if escaped {
			slots = append(slots, maskSlot{slotType: ch, literal: true})
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		switch ch {
		case '#', 'A', '*', '?':
			slots = append(slots, maskSlot{slotType: ch, literal: false})
		default:
			slots = append(slots, maskSlot{slotType: ch, literal: true})
		}
	}

	return slots
}

func (m *MaskedInput) maskDisplayLength() int {
	return len(m.parseMask())
}

func (m *MaskedInput) isValidForSlot(ch rune, slotType rune) bool {
	switch slotType {
	case '#':
		return unicode.IsDigit(ch)
	case 'A':
		return unicode.IsLetter(ch)
	case '*':
		return unicode.IsLetter(ch) || unicode.IsDigit(ch)
	case '?':
		return true
	default:
		return false
	}
}

func (m *MaskedInput) formatWithMask() string {
	if m == nil {
		return ""
	}
	slots := m.parseMask()
	if len(slots) == 0 {
		return m.value
	}

	result := make([]rune, len(slots))
	valueRunes := []rune(m.value)
	valueIdx := 0

	for i, slot := range slots {
		if slot.literal {
			result[i] = slot.slotType
		} else if valueIdx < len(valueRunes) {
			result[i] = valueRunes[valueIdx]
			valueIdx++
		} else {
			result[i] = m.placeholder
		}
	}

	return string(result)
}

func (m *MaskedInput) filterValue(value string) string {
	if m == nil {
		return ""
	}
	slots := m.parseMask()
	if len(slots) == 0 {
		return value
	}

	var result []rune
	slotIdx := 0

	for _, ch := range value {
		// Find the next editable slot
		for slotIdx < len(slots) && slots[slotIdx].literal {
			slotIdx++
		}
		if slotIdx >= len(slots) {
			break
		}

		if m.isValidForSlot(ch, slots[slotIdx].slotType) {
			result = append(result, ch)
			slotIdx++
		}
	}

	return string(result)
}

func (m *MaskedInput) firstEditablePosition() int {
	slots := m.parseMask()
	for i, slot := range slots {
		if !slot.literal {
			return i
		}
	}
	return 0
}

func (m *MaskedInput) lastEditablePosition() int {
	slots := m.parseMask()
	last := 0
	for i, slot := range slots {
		if !slot.literal {
			last = i
		}
	}
	return last
}

func (m *MaskedInput) lastFilledPosition() int {
	slots := m.parseMask()
	valueRunes := []rune(m.value)
	valueIdx := 0
	lastFilled := -1

	for i, slot := range slots {
		if slot.literal {
			continue
		}
		if valueIdx < len(valueRunes) {
			lastFilled = i
			valueIdx++
		}
	}

	return lastFilled
}

func (m *MaskedInput) displayToValueIndex(displayPos int) int {
	slots := m.parseMask()
	valueIdx := 0
	for i := 0; i < displayPos && i < len(slots); i++ {
		if !slots[i].literal {
			valueIdx++
		}
	}
	return valueIdx
}

func (m *MaskedInput) moveCursor(delta int) {
	slots := m.parseMask()
	if len(slots) == 0 {
		return
	}

	newPos := m.cursorPos + delta
	if newPos < 0 {
		newPos = 0
	}
	if newPos > len(slots) {
		newPos = len(slots)
	}

	// Skip literal characters when moving
	if delta > 0 {
		for newPos < len(slots) && slots[newPos].literal {
			newPos++
		}
	} else if delta < 0 {
		for newPos > 0 && newPos < len(slots) && slots[newPos].literal {
			newPos--
		}
		// Also skip back if we landed on a literal at position 0
		if newPos == 0 && len(slots) > 0 && slots[0].literal {
			for newPos < len(slots) && slots[newPos].literal {
				newPos++
			}
		}
	}

	m.cursorPos = newPos
}

func (m *MaskedInput) insertChar(ch rune) bool {
	slots := m.parseMask()
	if len(slots) == 0 {
		return false
	}

	// Find the current editable position
	pos := m.cursorPos
	for pos < len(slots) && slots[pos].literal {
		pos++
	}
	if pos >= len(slots) {
		return false
	}

	// Validate the character for this slot
	if !m.isValidForSlot(ch, slots[pos].slotType) {
		return false
	}

	// Calculate value index and insert
	valueIdx := m.displayToValueIndex(pos)
	valueRunes := []rune(m.value)

	// Count total editable slots
	editableCount := 0
	for _, slot := range slots {
		if !slot.literal {
			editableCount++
		}
	}

	// Insert or replace at valueIdx
	if valueIdx < len(valueRunes) {
		// Replace existing
		valueRunes[valueIdx] = ch
	} else if len(valueRunes) < editableCount {
		// Append
		valueRunes = append(valueRunes, ch)
	} else {
		return false
	}

	m.value = string(valueRunes)

	// Move cursor to next editable position
	m.cursorPos = pos + 1
	for m.cursorPos < len(slots) && slots[m.cursorPos].literal {
		m.cursorPos++
	}

	return true
}

func (m *MaskedInput) deleteAtCursor(backward bool) bool {
	slots := m.parseMask()
	if len(slots) == 0 || m.value == "" {
		return false
	}

	pos := m.cursorPos
	if backward {
		// Find previous editable position
		pos--
		for pos >= 0 && slots[pos].literal {
			pos--
		}
		if pos < 0 {
			return false
		}
	} else {
		// Find current or next editable position
		for pos < len(slots) && slots[pos].literal {
			pos++
		}
		if pos >= len(slots) {
			return false
		}
	}

	valueIdx := m.displayToValueIndex(pos)
	valueRunes := []rune(m.value)

	if valueIdx >= len(valueRunes) {
		return false
	}

	// Remove character at valueIdx
	valueRunes = append(valueRunes[:valueIdx], valueRunes[valueIdx+1:]...)
	m.value = string(valueRunes)

	if backward {
		m.cursorPos = pos
	}

	return true
}

var _ runtime.Widget = (*MaskedInput)(nil)
var _ runtime.Focusable = (*MaskedInput)(nil)
var _ runtime.Bindable = (*MaskedInput)(nil)
var _ runtime.Unbindable = (*MaskedInput)(nil)
var _ Validatable = (*MaskedInput)(nil)
var _ clipboard.Target = (*MaskedInput)(nil)
